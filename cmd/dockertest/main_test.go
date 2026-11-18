package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDiagnosticsKeepsTCPEndpointsSeparateFromForwardingHeaders(t *testing.T) {
	info := instance{
		Hostname:     "dockertest",
		PodName:      "dockertest-abc",
		PodIP:        "192.0.2.10",
		InterfaceIPs: []string{"192.0.2.10/24", "2001:db8::10/64"},
	}
	server := httptest.NewServer(diagnosticHandler(info))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/diagnostics", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("X-Forwarded-For", "203.0.113.1, 198.51.100.2")
	request.Header.Add("X-Forwarded-For", "2001:db8::3")
	request.Header.Set("Forwarded", `for="[2001:db8::3]:1234"`)
	request.Header.Set("Authorization", "Bearer test-value")
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected status or caching policy: %s %v", response.Status, response.Header)
	}
	var got diagnostics
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	peerIP, peerPort, err := net.SplitHostPort(got.PeerAddr)
	if err != nil || peerIP != "127.0.0.1" || peerPort == "" || got.PeerAddr == got.LocalAddr {
		t.Fatalf("unexpected TCP peer: %q", got.PeerAddr)
	}
	if got.LocalAddr != server.Listener.Addr().String() {
		t.Fatalf("local endpoint = %q, want %q", got.LocalAddr, server.Listener.Addr())
	}
	if !reflect.DeepEqual(got.Instance, info) {
		t.Fatalf("instance = %+v, want %+v", got.Instance, info)
	}
	wantHeaders := map[string][]string{
		"X-Forwarded-For": {"203.0.113.1, 198.51.100.2", "2001:db8::3"},
		"Forwarded":       {`for="[2001:db8::3]:1234"`},
	}
	if !reflect.DeepEqual(got.ForwardingHeaders, wantHeaders) {
		t.Fatalf("headers = %v, want %v", got.ForwardingHeaders, wantHeaders)
	}
}

func TestDiagnosticsPageEscapesHeaderValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "<b>unverified</b>")
	out := httptest.NewRecorder()
	diagnosticHandler(instance{Hostname: "dockertest"}).ServeHTTP(out, req)
	if out.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", out.Code)
	}
	body := out.Body.String()
	if !strings.Contains(body, "&lt;b&gt;unverified&lt;/b&gt;") || !strings.Contains(body, "Not configured") {
		t.Fatalf("expected escaped forwarding header and indication of missing Pod metadata")
	}
}

func TestDiagnosticsShowsOnlyAllowlistedEnvironment(t *testing.T) {
	t.Setenv("DIAGNOSTICS_ENV_SHOW_ALL", "false")
	t.Setenv("DIAGNOSTICS_ENV_ALLOWLIST", " DOCKERTEST_TEST_LABEL, ,DOCKERTEST_TEST_EMPTY, DOCKERTEST_TEST_LABEL ")
	t.Setenv("DOCKERTEST_TEST_LABEL", "<example>")
	t.Setenv("DOCKERTEST_TEST_EMPTY", "")
	t.Setenv("DOCKERTEST_TEST_HIDDEN", "not for display")
	handler := diagnosticHandler(instance{})

	out := httptest.NewRecorder()
	handler.ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/api/diagnostics", nil))
	var got diagnostics
	if err := json.Unmarshal(out.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"DOCKERTEST_TEST_LABEL": "<example>", "DOCKERTEST_TEST_EMPTY": ""}
	if !reflect.DeepEqual(got.Environment, want) {
		t.Fatalf("environment = %v, want %v", got.Environment, want)
	}
	if got.Runtime.StartedAt.IsZero() || got.Runtime.UptimeSeconds < 0 ||
		got.Runtime.GoVersion == "" || got.Runtime.OS == "" || got.Runtime.Arch == "" ||
		got.Runtime.LogicalCPUs < 1 || got.Runtime.GOMAXPROCS < 1 {
		t.Fatalf("incomplete runtime diagnostics: %+v", got.Runtime)
	}

	out = httptest.NewRecorder()
	handler.ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/", nil))
	body := out.Body.String()
	if out.Code != http.StatusOK || !strings.Contains(body, got.Runtime.GoVersion) ||
		!strings.Contains(body, "&lt;example&gt;") || !strings.Contains(body, "(empty)") ||
		strings.Contains(body, "DOCKERTEST_TEST_HIDDEN") || strings.Contains(body, "not for display") {
		t.Fatal("expected runtime data and escaped allowlisted environment values only")
	}
}

func TestDiagnosticsDoesNotExposeEnvironmentByDefault(t *testing.T) {
	t.Setenv("DIAGNOSTICS_ENV_SHOW_ALL", "")
	t.Setenv("DIAGNOSTICS_ENV_ALLOWLIST", "")
	t.Setenv("DOCKERTEST_TEST_HIDDEN", "not for display")
	out := httptest.NewRecorder()
	diagnosticHandler(instance{}).ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/api/diagnostics", nil))
	var got diagnostics
	if err := json.Unmarshal(out.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Environment == nil || len(got.Environment) != 0 {
		t.Fatalf("expected an empty environment object, got %v", got.Environment)
	}
}

func TestDiagnosticsCanShowAllEnvironmentVariables(t *testing.T) {
	t.Setenv("DIAGNOSTICS_ENV_SHOW_ALL", "true")
	t.Setenv("DIAGNOSTICS_ENV_ALLOWLIST", "DOCKERTEST_TEST_LABEL")
	t.Setenv("DOCKERTEST_TEST_LABEL", "allowed")
	t.Setenv("DOCKERTEST_TEST_EXTRA", "<not-allowlisted>=example")
	t.Setenv("DOCKERTEST_TEST_EMPTY", "")
	handler := diagnosticHandler(instance{})

	out := httptest.NewRecorder()
	handler.ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/api/diagnostics", nil))
	var got diagnostics
	if err := json.Unmarshal(out.Body.Bytes(), &got); err != nil {
		t.Fatal("invalid diagnostics JSON")
	}
	for name, want := range map[string]string{
		"DOCKERTEST_TEST_LABEL": "allowed",
		"DOCKERTEST_TEST_EXTRA": "<not-allowlisted>=example",
		"DOCKERTEST_TEST_EMPTY": "",
	} {
		if value, ok := got.Environment[name]; !ok || value != want {
			t.Errorf("missing or incorrect environment entry for %s", name)
		}
	}

	out = httptest.NewRecorder()
	handler.ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/", nil))
	if out.Code != http.StatusOK || out.Header().Get("Cache-Control") != "no-store" ||
		!strings.Contains(out.Body.String(), "&lt;not-allowlisted&gt;=example") {
		t.Fatal("expected uncached HTML with the escaped environment value")
	}
}

package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

//go:embed index.html
var indexHTML string

var index = template.Must(template.New("index").Parse(indexHTML))

type instance struct {
	Hostname     string   `json:"hostname"`
	PodName      string   `json:"pod_name,omitempty"`
	PodIP        string   `json:"pod_ip,omitempty"`
	InterfaceIPs []string `json:"interface_ips"`
}

type diagnostics struct {
	Instance          instance            `json:"instance"`
	LocalAddr         string              `json:"local_addr"`
	PeerAddr          string              `json:"peer_addr"`
	ForwardingHeaders map[string][]string `json:"forwarding_headers"`
	Runtime           runtimeInfo         `json:"runtime"`
	Environment       map[string]string   `json:"environment"`
}

type runtimeInfo struct {
	GoVersion     string    `json:"go_version"`
	OS            string    `json:"os"`
	Arch          string    `json:"arch"`
	LogicalCPUs   int       `json:"logical_cpus"`
	GOMAXPROCS    int       `json:"gomaxprocs"`
	StartedAt     time.Time `json:"started_at"`
	UptimeSeconds int64     `json:"uptime_seconds"`
}

var forwardingHeaders = []string{
	"Forwarded",
	"X-Forwarded-For",
	"X-Real-IP",
	"X-Original-Forwarded-For",
}

func diagnosticEnvironment() map[string]string {
	environment := make(map[string]string)
	switch os.Getenv("DIAGNOSTICS_ENV_SHOW_ALL") {
	case "true":
		log.Print("warning: diagnostics expose all process environment variables")
		for _, entry := range os.Environ() {
			if name, value, ok := strings.Cut(entry, "="); ok {
				environment[name] = value
			}
		}
		return environment
	case "", "false":
	default:
		log.Print("invalid DIAGNOSTICS_ENV_SHOW_ALL: expected true or false; using the allowlist")
	}

	for name := range strings.SplitSeq(os.Getenv("DIAGNOSTICS_ENV_ALLOWLIST"), ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if value, ok := os.LookupEnv(name); ok {
			environment[name] = value
		}
	}
	return environment
}

func diagnosticHandler(info instance) http.Handler {
	startedAt := time.Now()
	environment := diagnosticEnvironment()

	report := func(w http.ResponseWriter, r *http.Request) diagnostics {
		w.Header().Set("Cache-Control", "no-store")
		d := diagnostics{
			Instance:          info,
			PeerAddr:          r.RemoteAddr,
			ForwardingHeaders: make(map[string][]string),
			Runtime: runtimeInfo{
				GoVersion:     runtime.Version(),
				OS:            runtime.GOOS,
				Arch:          runtime.GOARCH,
				LogicalCPUs:   runtime.NumCPU(),
				GOMAXPROCS:    runtime.GOMAXPROCS(0),
				StartedAt:     startedAt.UTC(),
				UptimeSeconds: int64(time.Since(startedAt) / time.Second),
			},
			Environment: environment,
		}
		if addr, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
			d.LocalAddr = addr.String()
		}
		for _, name := range forwardingHeaders {
			if values := r.Header.Values(name); len(values) != 0 {
				d.ForwardingHeaders[name] = values
			}
		}
		return d
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		var body bytes.Buffer
		if err := index.Execute(&body, report(w, r)); err != nil {
			log.Printf("render diagnostics: %v", err)
			http.Error(w, "Could not render diagnostics", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; frame-ancestors 'none'")
		if _, err := body.WriteTo(w); err != nil {
			log.Printf("write diagnostics page: %v", err)
		}
	})
	mux.HandleFunc("GET /api/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		data := report(w, r)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("write diagnostics JSON: %v", err)
		}
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("read hostname: %v", err)
	}
	info := instance{
		Hostname:     hostname,
		PodName:      os.Getenv("POD_NAME"),
		PodIP:        os.Getenv("POD_IP"),
		InterfaceIPs: []string{},
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("read interface addresses: %v", err)
	}
	for _, addr := range addrs {
		info.InterfaceIPs = append(info.InterfaceIPs, addr.String())
	}
	server := &http.Server{
		Addr:              ":8080",
		Handler:           diagnosticHandler(info),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}
	log.Printf("listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS test

WORKDIR /src
COPY go.mod ./
COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 go test ./...

FROM test AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /dockertest ./cmd/dockertest

FROM scratch
COPY --from=build /dockertest /dockertest
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/dockertest"]


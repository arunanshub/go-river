build:
    GOAMD64=v3 CGO_ENABLED=0 \
    go build -o bin/goriver -ldflags="-s -w" -trimpath ./cmd/goriver

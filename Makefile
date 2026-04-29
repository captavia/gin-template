all:
	go build -ldflags="-s -w" -trimpath -o ./build/ ./cmd/api
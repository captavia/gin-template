all:
	go build -ldflags="-s -w" -trimpath -tags=nomsgpack -o ./build/ ./cmd/api
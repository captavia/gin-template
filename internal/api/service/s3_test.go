package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/do/v2"

	"template/config"
)

func createTestInjector(t *testing.T, s3Endpoint string, secure bool, defaultBucket string) do.Injector {
	t.Helper()
	injector := do.New()

	cfg := &config.Config{
		S3: config.S3Config{
			Endpoint:        s3Endpoint,
			SSL:             secure,
			AccessKeyID:     "mockAccessKey",
			SecretAccessKey: "mockSecretKey1234567890",
			Bucket:          defaultBucket,
			Region:          "us-east-1",
		},
	}
	do.ProvideValue(injector, cfg)

	minioClient, err := minio.New(cfg.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
		Secure: cfg.S3.SSL,
		Region: cfg.S3.Region,
	})
	if err != nil {
		t.Fatalf("failed to create minio client: %v", err)
	}
	do.ProvideValue(injector, minioClient)

	do.Provide(injector, NewS3Service)
	return injector
}

func TestS3Service_DependencyInjection(t *testing.T) {
	injector := createTestInjector(t, "localhost:9000", false, "test-bucket")
	svc, err := do.Invoke[*S3Service](injector)
	if err != nil {
		t.Fatalf("failed to invoke S3Service: %v", err)
	}
	if svc == nil {
		t.Fatal("expected S3Service instance, got nil")
	}
	if svc.Client() == nil {
		t.Fatal("expected non-nil minio client")
	}
}

func TestS3Service_PathValidation(t *testing.T) {
	injector := createTestInjector(t, "localhost:9000", false, "test-bucket")
	svc := do.MustInvoke[*S3Service](injector)

	tests := []struct {
		input       string
		expected    string
		expectError bool
	}{
		{input: "/images/avatar.png", expected: "images/avatar.png", expectError: false},
		{input: "images/avatar.png", expected: "images/avatar.png", expectError: false},
		{input: "  /docs/file.pdf  ", expected: "docs/file.pdf", expectError: false},
		{input: "", expected: "", expectError: true},
		{input: "   ", expected: "", expectError: true},
		{input: "/", expected: "", expectError: true},
	}

	for _, tc := range tests {
		clean, err := svc.cleanPath(tc.input)
		if tc.expectError {
			if err == nil {
				t.Errorf("cleanPath(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("cleanPath(%q) unexpected error: %v", tc.input, err)
			}
			if clean != tc.expected {
				t.Errorf("cleanPath(%q) = %q, expected %q", tc.input, clean, tc.expected)
			}
		}
	}
}

func TestS3Service_BucketResolution(t *testing.T) {
	// Case 1: Default bucket configured
	injector1 := createTestInjector(t, "localhost:9000", false, "default-bucket")
	svc1 := do.MustInvoke[*S3Service](injector1)

	opts1 := svc1.applyOptions()
	b1, err := svc1.resolveBucket(opts1.Bucket)
	if err != nil || b1 != "default-bucket" {
		t.Fatalf("expected 'default-bucket', got %q, err: %v", b1, err)
	}

	// Override with WithBucket
	opts1Override := svc1.applyOptions(WithBucket("override-bucket"))
	b1Override, err := svc1.resolveBucket(opts1Override.Bucket)
	if err != nil || b1Override != "override-bucket" {
		t.Fatalf("expected 'override-bucket', got %q, err: %v", b1Override, err)
	}

	// Case 2: No default bucket
	injector2 := createTestInjector(t, "localhost:9000", false, "")
	svc2 := do.MustInvoke[*S3Service](injector2)

	opts2 := svc2.applyOptions()
	_, err2 := svc2.resolveBucket(opts2.Bucket)
	if err2 != ErrBucketNotConfigured {
		t.Fatalf("expected ErrBucketNotConfigured, got %v", err2)
	}

	// Providing bucket via option works even without default
	opts2Custom := svc2.applyOptions(WithBucket("custom-bucket"))
	b2Custom, err2Custom := svc2.resolveBucket(opts2Custom.Bucket)
	if err2Custom != nil || b2Custom != "custom-bucket" {
		t.Fatalf("expected 'custom-bucket', got %q, err: %v", b2Custom, err2Custom)
	}
}

func TestS3Service_DetectContentType(t *testing.T) {
	tests := []struct {
		path        string
		data        []byte
		expectedSub string
	}{
		{path: "test.png", data: nil, expectedSub: "image/png"},
		{path: "file.pdf", data: nil, expectedSub: "application/pdf"},
		{path: "doc.json", data: nil, expectedSub: "application/json"},
		{path: "unknown.xyz", data: []byte("<!DOCTYPE html><html><body>Hello</body></html>"), expectedSub: "text/html"},
		{path: "binary.bin", data: []byte{0x00, 0x01, 0x02, 0x03}, expectedSub: "application/octet-stream"},
	}

	for _, tc := range tests {
		ct := detectContentType(tc.path, tc.data)
		if !strings.Contains(ct, tc.expectedSub) {
			t.Errorf("detectContentType(%q) = %q, expected to contain %q", tc.path, ct, tc.expectedSub)
		}
	}
}

func TestS3Service_GetPublicURL(t *testing.T) {
	// HTTP
	injectorHTTP := createTestInjector(t, "minio.local:9000", false, "public-bucket")
	svcHTTP := do.MustInvoke[*S3Service](injectorHTTP)

	u1, err := svcHTTP.GetPublicURL("assets/logo.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected1 := "http://minio.local:9000/public-bucket/assets/logo.png"
	if u1 != expected1 {
		t.Fatalf("expected %q, got %q", expected1, u1)
	}

	// HTTPS + WithBucket override
	injectorHTTPS := createTestInjector(t, "s3.example.com", true, "default-bucket")
	svcHTTPS := do.MustInvoke[*S3Service](injectorHTTPS)

	u2, err := svcHTTPS.GetPublicURL("/files/report.pdf", WithBucket("custom-bucket"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected2 := "https://s3.example.com/custom-bucket/files/report.pdf"
	if u2 != expected2 {
		t.Fatalf("expected %q, got %q", expected2, u2)
	}
}

func TestS3Service_PresignedAndPreviewURLs(t *testing.T) {
	injector := createTestInjector(t, "localhost:9000", false, "media-bucket")
	svc := do.MustInvoke[*S3Service](injector)
	ctx := context.Background()

	// 1. Presigned URL
	presignedURL, err := svc.GetPresignedURL(ctx, "images/photo.jpg", 15*time.Minute)
	if err != nil {
		t.Fatalf("GetPresignedURL failed: %v", err)
	}
	if !strings.Contains(presignedURL, "images/photo.jpg") || !strings.Contains(presignedURL, "X-Amz-Signature") {
		t.Fatalf("GetPresignedURL returned invalid format: %s", presignedURL)
	}

	// 2. Preview URL (checks response-content-disposition=inline)
	previewURL, err := svc.GetPreviewURL(ctx, "docs/manual.pdf", 30*time.Minute)
	if err != nil {
		t.Fatalf("GetPreviewURL failed: %v", err)
	}
	parsedPreview, err := url.Parse(previewURL)
	if err != nil {
		t.Fatalf("failed to parse preview URL: %v", err)
	}
	if parsedPreview.Query().Get("response-content-disposition") != "inline" {
		t.Fatalf("expected response-content-disposition=inline, got %q", parsedPreview.Query().Get("response-content-disposition"))
	}
	if parsedPreview.Query().Get("response-content-type") != "application/pdf" {
		t.Fatalf("expected response-content-type=application/pdf, got %q", parsedPreview.Query().Get("response-content-type"))
	}

	// 3. Download URL (checks response-content-disposition=attachment)
	downloadURL, err := svc.GetDownloadURL(ctx, "docs/manual.pdf", "user_guide.pdf", 30*time.Minute)
	if err != nil {
		t.Fatalf("GetDownloadURL failed: %v", err)
	}
	parsedDownload, err := url.Parse(downloadURL)
	if err != nil {
		t.Fatalf("failed to parse download URL: %v", err)
	}
	expectedDisposition := `attachment; filename="user_guide.pdf"`
	if parsedDownload.Query().Get("response-content-disposition") != expectedDisposition {
		t.Fatalf("expected response-content-disposition=%q, got %q", expectedDisposition, parsedDownload.Query().Get("response-content-disposition"))
	}

	// 4. Invalid expiration
	_, err = svc.GetPresignedURL(ctx, "photo.jpg", 8*24*time.Hour)
	if err != ErrInvalidExpiration {
		t.Fatalf("expected ErrInvalidExpiration, got %v", err)
	}
}

func TestS3Service_MockServerOperations(t *testing.T) {
	// Create mock HTTP server simulating S3 responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Handle bucket location or head bucket
		if r.URL.Query().Has("location") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<LocationConstraint>us-east-1</LocationConstraint>`))
			return
		}
		if path == "/test-bucket" || path == "/test-bucket/" {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodPut:
			if strings.Contains(path, "upload-fail") {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("ETag", `"mocketag123"`)
			w.WriteHeader(http.StatusOK)

		case http.MethodHead:
			if strings.Contains(path, "exists.txt") {
				w.Header().Set("Content-Length", "11")
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("ETag", `"mocketag123"`)
				w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
				w.WriteHeader(http.StatusOK)
				return
			}
			// Object not found
			w.WriteHeader(http.StatusNotFound)

		case http.MethodGet:
			if strings.Contains(path, "exists.txt") {
				w.Header().Set("Content-Length", "11")
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("ETag", `"mocketag123"`)
				w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("hello world"))
				return
			}
			w.WriteHeader(http.StatusNotFound)

		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	endpoint := u.Host

	injector := createTestInjector(t, endpoint, false, "test-bucket")
	svc := do.MustInvoke[*S3Service](injector)
	ctx := context.Background()

	// 1. PutBytes
	putInfo, err := svc.PutBytes(ctx, "data/file.txt", []byte("hello world"), WithContentType("text/plain"))
	if err != nil {
		t.Fatalf("PutBytes failed: %v", err)
	}
	if putInfo.Key != "data/file.txt" || putInfo.Size != 11 {
		t.Fatalf("unexpected put info: %+v", putInfo)
	}

	// 2. Put stream
	streamData := []byte("stream test")
	putStreamInfo, err := svc.Put(ctx, "data/stream.txt", bytes.NewReader(streamData), int64(len(streamData)))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if putStreamInfo.Key != "data/stream.txt" || putStreamInfo.Size != int64(len(streamData)) {
		t.Fatalf("unexpected put stream info: %+v", putStreamInfo)
	}

	// 3. Exists - Existing file
	exists, err := svc.Exists(ctx, "exists.txt")
	if err != nil {
		t.Fatalf("Exists returned error: %v", err)
	}
	if !exists {
		t.Fatal("expected exists to be true")
	}

	// 4. Exists - Non-existing file
	notExists, err := svc.Exists(ctx, "non-existing.txt")
	if err != nil {
		t.Fatalf("Exists for non-existing file returned error: %v", err)
	}
	if notExists {
		t.Fatal("expected exists to be false")
	}

	// 5. Stat - Existing file
	statInfo, err := svc.Stat(ctx, "exists.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if statInfo.Size != 11 || statInfo.Key != "exists.txt" {
		t.Fatalf("unexpected StatInfo: %+v", statInfo)
	}

	// 6. Stat - Non-existing file returns ErrObjectNotFound
	_, err = svc.Stat(ctx, "non-existing.txt")
	if err != ErrObjectNotFound {
		t.Fatalf("expected ErrObjectNotFound, got %v", err)
	}

	// 7. GetBytes
	data, err := svc.GetBytes(ctx, "exists.txt")
	if err != nil {
		t.Fatalf("GetBytes failed: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(data))
	}

	// 8. Delete
	err = svc.Delete(ctx, "exists.txt")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

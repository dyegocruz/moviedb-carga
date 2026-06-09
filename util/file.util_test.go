package util

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func withFatalPanic(t *testing.T) {
	t.Helper()
	old := fileFatalFn
	fileFatalFn = func(v ...interface{}) { panic("fatal") }
	t.Cleanup(func() { fileFatalFn = old })
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestDownloadExportFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir temp dir: %v", err)
	}

	const content = "payload"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sample.json.gz" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	DownloadExportFile(server.URL, "sample")

	filePath := filepath.Join(tmpDir, "sample.json.gz")
	body, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected downloaded file: %v", err)
	}

	if string(body) != content {
		t.Fatalf("unexpected file content: %q", string(body))
	}
}

func TestUnzipAndRemoveFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir temp dir: %v", err)
	}

	const fileBase = "catalog"
	const payload = "json-content"

	gzFile, err := os.Create(fileBase + ".json.gz")
	if err != nil {
		t.Fatalf("failed to create gz file: %v", err)
	}

	gzWriter := gzip.NewWriter(gzFile)
	if _, err := io.WriteString(gzWriter, payload); err != nil {
		t.Fatalf("failed to write gz payload: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("failed to close gz writer: %v", err)
	}
	if err := gzFile.Close(); err != nil {
		t.Fatalf("failed to close gz file: %v", err)
	}

	Unzip(fileBase)

	if _, err := os.Stat(fileBase + ".json.gz"); err == nil {
		t.Fatal("expected gz file to be removed after unzip")
	}

	body, err := os.ReadFile(fileBase + ".json")
	if err != nil {
		t.Fatalf("expected unzipped json file: %v", err)
	}
	if string(body) != payload {
		t.Fatalf("unexpected json payload: %q", string(body))
	}

	RemoveFile(fileBase + ".json")
	if _, err := os.Stat(fileBase + ".json"); err == nil {
		t.Fatal("expected json file to be removed")
	}
}

func TestDownloadExportFile_PanicOnRequestError(t *testing.T) {
	withFatalPanic(t)
	assertPanics(t, func() {
		DownloadExportFile("://invalid-url", "sample")
	})
}

func TestDownloadExportFile_PanicOnCreateError(t *testing.T) {
	withFatalPanic(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("payload"))
	}))
	defer server.Close()

	assertPanics(t, func() {
		// Invalid nested path should fail os.Create when dir doesn't exist.
		DownloadExportFile(server.URL, "missing-dir/sample")
	})
}

func TestUnzip_PanicOnOpenError(t *testing.T) {
	withFatalPanic(t)
	assertPanics(t, func() {
		Unzip("does-not-exist")
	})
}

func TestUnzip_PanicOnInvalidGzip(t *testing.T) {
	withFatalPanic(t)

	tmpDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir temp dir: %v", err)
	}

	if err := os.WriteFile("broken.json.gz", []byte("not-gzip"), 0o600); err != nil {
		t.Fatalf("failed to write broken gz file: %v", err)
	}

	assertPanics(t, func() {
		Unzip("broken")
	})
}

func TestRemoveFile_PanicOnMissingFile(t *testing.T) {
	withFatalPanic(t)
	assertPanics(t, func() {
		RemoveFile("definitely-missing.file")
	})
}

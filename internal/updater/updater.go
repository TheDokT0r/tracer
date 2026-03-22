package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const repoURL = "https://api.github.com/repos/TheDokT0r/tracer/releases/latest"

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Check returns the latest version available, or empty string on error.
func Check() (string, error) {
	resp, err := http.Get(repoURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.TagName, nil
}

// NeedsUpdate returns true if latest is newer than current.
func NeedsUpdate(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	return current != latest && current != "dev"
}

// Update downloads the latest release and replaces the current binary.
func Update(current string) error {
	resp, err := http.Get(repoURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if !NeedsUpdate(current, r.TagName) {
		fmt.Printf("Already up to date (%s)\n", current)
		return nil
	}

	// Find the right asset
	target := fmt.Sprintf("tracer-%s-%s.tar.gz", runtime.GOOS, goArch())
	var downloadURL string
	for _, a := range r.Assets {
		if a.Name == target {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no release found for %s/%s", runtime.GOOS, goArch())
	}

	fmt.Printf("Updating %s -> %s\n", current, r.TagName)

	// Download
	dlResp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer dlResp.Body.Close()

	// Extract binary from tar.gz
	binary, err := extractBinary(dlResp.Body)
	if err != nil {
		return err
	}

	// Replace current binary
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Write to temp file next to current binary, then rename (atomic on same fs)
	tmpPath := execPath + ".tmp"
	if err := os.WriteFile(tmpPath, binary, 0755); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	fmt.Printf("Updated to %s\n", r.TagName)
	return nil
}

func extractBinary(r io.Reader) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == "tracer" {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("tracer binary not found in archive")
}

func goArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}

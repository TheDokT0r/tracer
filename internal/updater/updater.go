package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const repoURL = "https://api.github.com/repos/orkwitzel/tracer/releases/latest"

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Check returns the latest version tag from GitHub.
func Check() (string, error) {
	resp, err := http.Get(repoURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.TagName, nil
}

// NeedsUpdate returns true if latest is strictly newer than current.
func NeedsUpdate(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	if current == "dev" || current == latest {
		return false
	}
	return compareSemver(latest, current) > 0
}

// compareSemver returns >0 if a > b, 0 if equal, <0 if a < b.
func compareSemver(a, b string) int {
	ap := parseSemver(a)
	bp := parseSemver(b)
	for i := 0; i < 3; i++ {
		if ap[i] != bp[i] {
			return ap[i] - bp[i]
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	var parts [3]int
	for i, s := range strings.SplitN(v, ".", 3) {
		n, _ := strconv.Atoi(s)
		parts[i] = n
	}
	return parts
}

// Update downloads the latest release and replaces the current binary.
func Update(current string) error {
	fmt.Println("Checking for updates...")

	resp, err := http.Get(repoURL)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	fmt.Printf("Current version: %s\n", current)
	fmt.Printf("Latest version:  %s\n", r.TagName)

	if !NeedsUpdate(current, r.TagName) {
		fmt.Println("Already up to date.")
		return nil
	}

	// Find the right asset
	target := fmt.Sprintf("tracer-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, a := range r.Assets {
		if a.Name == target {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no release found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading %s...\n", target)

	dlResp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer dlResp.Body.Close()

	// Extract binary from tar.gz
	fmt.Println("Extracting...")
	binary, err := extractBinary(dlResp.Body)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine binary path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve binary path: %w", err)
	}

	fmt.Printf("Installing to %s...\n", execPath)

	// Try writing directly first
	tmpPath := execPath + ".tmp"
	err = os.WriteFile(tmpPath, binary, 0755)
	if err != nil {
		// Permission denied — try with sudo
		fmt.Println("Permission denied, retrying with sudo...")
		err = sudoInstall(binary, execPath)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	} else {
		if err := os.Rename(tmpPath, execPath); err != nil {
			os.Remove(tmpPath)
			// Rename failed — try with sudo
			fmt.Println("Permission denied, retrying with sudo...")
			err = sudoInstall(binary, execPath)
			if err != nil {
				return fmt.Errorf("install failed: %w", err)
			}
		}
	}

	fmt.Printf("Updated to %s\n", r.TagName)
	return nil
}

func sudoInstall(binary []byte, destPath string) error {
	// Write to temp file in a writable location
	tmp, err := os.CreateTemp("", "tracer-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(binary); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	cmd := exec.Command("sudo", "mv", tmpPath, destPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
			const maxBinarySize = 200 * 1024 * 1024 // 200 MB
			return io.ReadAll(io.LimitReader(tr, maxBinarySize))
		}
	}
	return nil, fmt.Errorf("tracer binary not found in archive")
}

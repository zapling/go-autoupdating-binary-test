package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"syscall"
)

const updateServer = "http://localhost:3000"
const version = "v0.0.1"

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Hello, I am %s\n", version)

	reader.ReadString('\n')

	fmt.Print("Checking for updates...")

	latestVersion, downloadPath, err := getLatestVersion()
	if err != nil {
		fmt.Printf("Failed to check version: %v", err)
		os.Exit(1)
	}

	if latestVersion == version {
		fmt.Println("Already the latest version")
		return
	}

	fmt.Println("Downloading latest version...")

	if err := downloadUpdate(downloadPath); err != nil {
		fmt.Printf("Failed to download version: %w", err)
		os.Exit(1)
	}

	fmt.Println("New version downloaded: app2")

	syscall.Exec("upgrade.sh", nil, nil)
}

func getLatestVersion() (string, string, error) {
	req, err := http.NewRequest("GET", updateServer+"/latest", nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to perform request: %w", err)
	}

	defer res.Body.Close()

	responseBody := struct {
		Version string `json:"version"`
		Path    string `json:"path"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
		return "", "", fmt.Errorf("failed to decode body: %w", err)
	}

	return responseBody.Version, responseBody.Path, nil
}

func downloadUpdate(path string) error {
	query := url.Values{}
	query.Add("path", path)

	req, err := http.NewRequest("GET", updateServer+"/file?"+query.Encode(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	if err := extractTarGz(res.Body); err != nil {
		return fmt.Errorf("failed to extract tar gz: %w", err)
	}

	return nil
}

func extractTarGz(gzipStream io.Reader) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	var header *tar.Header
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %w", err)
			}
		case tar.TypeReg:
			outFile, err := os.Create("app2")
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				// outFile.Close error omitted as Copy error is more interesting at this point
				outFile.Close()
				return fmt.Errorf("ExtractTarGz: Copy() failed: %w", err)
			}
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("ExtractTarGz: Close() failed: %w", err)
			}
		default:
			return fmt.Errorf("ExtractTarGz: uknown type: %b in %s", header.Typeflag, header.Name)
		}
	}
	if err != io.EOF {
		return fmt.Errorf("ExtractTarGz: Next() failed: %w", err)
	}
	return nil
}

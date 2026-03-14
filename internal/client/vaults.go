package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Vault struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Client) ListVaults() ([]Vault, error) {
	resp, err := c.request(http.MethodGet, "/vaults", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}

	var vaults []Vault
	if err := json.NewDecoder(resp.Body).Decode(&vaults); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return vaults, nil
}

func (c *Client) CreateVault(name string) (bool, error) {
	resp, err := c.request("POST", "/vaults/"+name, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return true, nil
	}

	if resp.StatusCode == http.StatusConflict {
		return false, nil
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf(
		"create vault: %s (status %d)",
		string(body),
		resp.StatusCode,
	)
}

func WalkVault(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			skip := map[string]bool{
				".potok": true,
				".git":   true,
				// ".obsidian": true,
				// ".trash":    true,
			}

			if skip[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		rel = filepath.ToSlash(rel)
		files = append(files, rel)
		return nil
	})

	return files, err

}

func (c *Client) UploadFile(vault, relPath string, data []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(relPath))
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}
	writer.Close()

	encodedPath := url.PathEscape(relPath)
	encodedPath = strings.ReplaceAll(encodedPath, "%2F", "/")

	resp, err := c.requestMultipart(
		"POST",
		"/vaults/"+vault+"/files/"+encodedPath,
		body,
		writer.FormDataContentType(),
	)
	if err != nil {
		return fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s (status %d)", string(respBody), resp.StatusCode)
	}

	return nil
}

// func (c *Client) UploadVault(vault, root string) error {
// 	files, err := WalkVault(root)
// 	if err != nil {
// 		return fmt.Errorf("walk vault: %w", err)
// 	}

// 	for _, file := range files {
// 		fmt.Printf("Uploading %s...\n", file)
// 		if err := c.UploadFile(vault, file, root); err != nil {
// 			return fmt.Errorf("upload %s: %w", file, err)
// 		}
// 	}

// 	return nil
// }

func (c *Client) ListFiles(vault string) ([]string, error) {
	resp, err := c.request(http.MethodGet, "/vaults/"+vault+"/files", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list files: %s (status %d)", string(body), resp.StatusCode)
	}

	var files []string
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return files, nil
}

func (c *Client) DownloadFile(vault, relPath string) ([]byte, error) {
	encodedPath := url.PathEscape(relPath)
	encodedPath = strings.ReplaceAll(encodedPath, "%2F", "/")

	resp, err := c.request(
		http.MethodGet,
		"/vaults/"+vault+"/files/"+encodedPath,
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download: %s (status %d)", string(body), resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return data, nil
}

func (c *Client) DeleteFile(vault, relPath string) error {
	encodedPath := url.PathEscape(relPath)
	encodedPath = strings.ReplaceAll(encodedPath, "%2F", "/")

	resp, err := c.request(
		http.MethodDelete,
		"/vaults/"+vault+"/files/"+encodedPath,
		nil,
	)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s (status %d)",
			string(body), resp.StatusCode,
		)
	}

	return nil
}

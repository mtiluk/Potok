package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"text/tabwriter"

// 	"github.com/michaeltukdev/Potok/internal/client"
// 	"github.com/michaeltukdev/Potok/internal/config"
// 	"github.com/michaeltukdev/Potok/internal/crypto"
// 	"github.com/michaeltukdev/Potok/internal/prompt"
// 	"github.com/spf13/cobra"
// 	"github.com/zalando/go-keyring"
// )

// var restoreVaultCmd = &cobra.Command{
// 	Use:   "restore-vault",
// 	Short: "Select a Vault to restore!",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		cfg, err := config.MustLoadWithAPIURL()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}

// 		secret, err := keyring.Get("potok", "api-key")
// 		if err != nil {
// 			fmt.Println("Error retrieving API key:", err)
// 			return
// 		}

// 		url := fmt.Sprintf("%s/users/%s/vaults", cfg.APIURL, cfg.Username)
// 		resp, err := client.MakeAuthenticatedRequest(secret, url)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer resp.Body.Close()

// 		if resp.StatusCode != 200 {
// 			fmt.Println("Unauthenticated! Please set or update your API key")
// 			return
// 		}

// 		vaults, err := client.ReadVaultsFromResponse(resp)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		if len(vaults) == 0 {
// 			fmt.Println("No vaults found on the server.")
// 			return
// 		}

// 		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
// 		fmt.Fprintln(w, "ID\tName\tCreated At\tUpdated At")
// 		fmt.Fprintln(w, "--\t----\t----------\t----------")
// 		for _, v := range vaults {
// 			created := v.CreatedAt.Format("2006-01-02 15:04")
// 			updated := v.UpdatedAt.Format("2006-01-02 15:04")
// 			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", v.ID, v.Name, created, updated)
// 		}
// 		w.Flush()

// 		vaultName := prompt.Input("Type the name of the vault you want to restore: ")
// 		destination := prompt.Input("Destination to restore to: ")

// 		filesURL := fmt.Sprintf("%s/users/%s/vaults/%s/files", cfg.APIURL, cfg.Username, vaultName)
// 		req, _ := http.NewRequest("GET", filesURL, nil)
// 		req.Header.Set("Authorization", secret)
// 		resp, err = http.DefaultClient.Do(req)
// 		if err != nil {
// 			fmt.Println("Failed to get file list:", err)
// 			return
// 		}
// 		defer resp.Body.Close()
// 		if resp.StatusCode != http.StatusOK {
// 			fmt.Printf("Failed to list files: %s\n", resp.Status)
// 			return
// 		}
// 		var files []string
// 		if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
// 			fmt.Println("Failed to decode file list:", err)
// 			return
// 		}

// 		vaultKeyringUser := "vault-" + vaultName
// 		vaultPassword, err := keyring.Get("potok", vaultKeyringUser)
// 		if err != nil {
// 			fmt.Printf("Failed to get vault password: %v\n", err)
// 			return
// 		}

// 		for i, relPath := range files {
// 			fileURL := fmt.Sprintf("%s/users/%s/vaults/%s/files/%s", cfg.APIURL, cfg.Username, vaultName, relPath)
// 			req, _ := http.NewRequest("GET", fileURL, nil)
// 			req.Header.Set("Authorization", secret)
// 			resp, err := http.DefaultClient.Do(req)
// 			if err != nil {
// 				fmt.Printf("Failed to download %s: %v\n", relPath, err)
// 				continue
// 			}
// 			encrypted, err := io.ReadAll(resp.Body)

// 			resp.Body.Close()
// 			if err != nil {
// 				fmt.Printf("Failed to read %s: %v\n", relPath, err)
// 				continue
// 			}

// 			decrypted, err := crypto.DecryptBytes(vaultPassword, encrypted)
// 			if err != nil {
// 				fmt.Printf("Failed to decrypt %s: %v\n", relPath, err)
// 				continue
// 			}

// 			absPath := filepath.Join(destination, relPath)
// 			if err := os.MkdirAll(filepath.Dir(absPath), 0700); err != nil {
// 				fmt.Printf("Failed to create directory for %s: %v\n", relPath, err)
// 				continue
// 			}
// 			if err := os.WriteFile(absPath, decrypted, 0644); err != nil {
// 				fmt.Printf("Failed to write %s: %v\n", relPath, err)
// 				continue
// 			}
// 			fmt.Printf("Restored %d/%d: %s\n", i+1, len(files), relPath)
// 		}

// 		fmt.Println("Vault restore complete!")
// 	},
// }

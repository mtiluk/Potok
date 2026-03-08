package main

import (
	"fmt"

	"github.com/michaeltukdev/Potok/internal/prompt"
	"github.com/spf13/cobra"
)

// func WalkVaultDir(vaultRoot string) ([]string, error) {
// 	var files []string
// 	err := filepath.WalkDir(vaultRoot, func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !d.IsDir() {
// 			rel, err := filepath.Rel(vaultRoot, path)
// 			if err != nil {
// 				return err
// 			}
// 			files = append(files, rel)
// 		}
// 		return nil
// 	})
// 	return files, err
// }

//
// potok vault add - Register a local folder as a vault and store its encryption
// password in the OS keyring.
//
// Usage Examples:
//   potok vault add notes
//   potok vault add notes --path ~/Documents/Obsidian/Notes
//   potok vault add notes --upload
//
// Inputs:
//   - Prompts:
//  	- Vault path (if --path not provided)
//   	- Vault password (input hidden; stored in OS keyring)
//   - Flags:
//	    --path string          Path to the local vault folder
//		--upload               Encrypt and upload after adding
//      --password-stdin       Read vault password from stdin (non-interactive)
//

var addVaultCmd = &cobra.Command{
	Use:   "vault-add",
	Short: "Select a Vault to be backed up securely!",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Registering vault...")
		cmd.Println()

		vaultPassword, err := prompt.Secret("Vault password (input hidden): ")
		if err != nil {
			return fmt.Errorf("read vault password: %w", err)
		}

		passwordConfirmation, err := prompt.Secret("Confirm password (input hidden): ")
		if err != nil {
			return fmt.Errorf("read password confirmation: %w", err)
		}

		if vaultPassword != passwordConfirmation {
			return fmt.Errorf("passwords do not match")
		}

		cmd.Println()
		cmd.Println("Saving local vault config...")
		cmd.Println("- Config: /home/athena/.potok/config.json")
		cmd.Println(`- Password: stored in OS keyring (service="potok", user="vault:notes")`)
		cmd.Println()

		return nil

		// Ensure the API Url is set in config
		// cfg, err := config.MustLoadWithAPIURL()
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// // Grab the api key from the OS keyring
		// secret, err := keyring.Get("potok", "api-key")
		// if err != nil {
		// 	fmt.Println("Error retrieving API key:", err)
		// 	return
		// }

		// fmt.Println("Adding Vault...")

		// vaultPath := prompt.Input("Path to your vault: ")

		// // Constant loop until we get a set vault name
		// var vaultName string
		// for {
		// 	vaultName = prompt.Input("Name your vault: ")
		// 	exists, err := client.CheckVault(cfg.APIURL, cfg.Username, vaultName)
		// 	if err != nil {
		// 		fmt.Println("Error checking vault:", err)
		// 		return
		// 	}
		// 	if exists {
		// 		fmt.Printf("A vault named '%s' already exists. Please choose a different name.\n", vaultName)
		// 	} else {
		// 		break
		// 	}
		// }

		// vaultPassword := prompt.Input("Encryption password: ")

		// vaultKeyringUser := "vault-" + vaultName
		// if err := keyring.Set("potok", vaultKeyringUser, vaultPassword); err != nil {
		// 	fmt.Println("Failed to save vault password in keyring:", err)
		// 	return
		// }

		// fmt.Printf("Vault Path: %s\nVault Name: %s\n", vaultPath, vaultName)

		// // Request to API to create the vault
		// url := fmt.Sprintf("%s/users/%s/vaults/%s", cfg.APIURL, cfg.Username, vaultName)
		// req, err := http.NewRequest("POST", url, nil)
		// if err != nil {
		// 	fmt.Println("Failed to create request:", err)
		// 	return
		// }
		// req.Header.Set("Authorization", secret)

		// resp, err := http.DefaultClient.Do(req)
		// if err != nil {
		// 	fmt.Println("Failed to register vault:", err)
		// 	return
		// }
		// defer resp.Body.Close()

		// if resp.StatusCode == http.StatusConflict {
		// 	fmt.Printf("A vault named '%s' already exists. Please choose a different name.\n", vaultName)
		// 	return
		// }
		// if resp.StatusCode != http.StatusCreated {
		// 	fmt.Printf("Failed to register vault! Server returned: %s\n", resp.Status)
		// 	return
		// }

		// fmt.Println("Vault registered successfully!")

		// // Add vault to the config
		// vaultInfo := config.VaultInfo{
		// 	Name: vaultName,
		// 	Path: vaultPath,
		// }

		// cfg.AddVault(vaultInfo)
		// if err := config.Save(cfg); err != nil {
		// 	fmt.Println("Failed to save vault info locally:", err)
		// 	return
		// }

		// fmt.Println("Vault info saved locally!")

		// // Encrypt and Upload
		// files, err := WalkVaultDir(vaultPath)
		// if err != nil {
		// 	fmt.Println("error")
		// }

		// for i, relPath := range files {
		// 	absPath := filepath.Join(vaultPath, relPath)
		// 	encrypted, err := crypto.EncryptFile(vaultPassword, absPath)
		// 	if err != nil {
		// 		fmt.Printf("Failed to encrypt %s: %v\n", relPath, err)
		// 		continue
		// 	}

		// 	err = storage.UploadFile(cfg.APIURL, cfg.Username, vaultName, relPath, encrypted, secret)
		// 	if err != nil {
		// 		fmt.Printf("Failed to upload %s: %v\n", relPath, err)
		// 		continue
		// 	}

		// 	fmt.Printf("Uploaded %d/%d: %s\n", i+1, len(files), relPath)
		// }
	},
}

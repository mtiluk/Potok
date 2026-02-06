package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"

// 	"github.com/michaeltukdev/Potok/internal/client"
// 	"github.com/michaeltukdev/Potok/internal/config"
// 	"github.com/michaeltukdev/Potok/internal/prompt"
// 	"github.com/spf13/cobra"
// 	"github.com/zalando/go-keyring"
// )

// var setApiKeyCmd = &cobra.Command{
// 	Use:   "set-api-key",
// 	Short: "Set your API key",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		service := "potok"
// 		user := "api-key"

// 		cfg, err := config.MustLoadWithAPIURL()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}

// 		apiKey := prompt.Input("Enter your API key: ")

// 		r, err := client.MakeAuthenticatedRequest(apiKey, cfg.APIURL+"/me")
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer r.Body.Close()

// 		if r.StatusCode != 200 {
// 			fmt.Println("Authentication Test Request Failed! Key not modified or saved. Please check your API url and make sure the server is running!")
// 			return
// 		}

// 		// TODO: I really need to improve this but it works for now
// 		var me struct {
// 			Username string `json:"username"`
// 			ID       int    `json:"id"`
// 		}

// 		if err := json.NewDecoder(r.Body).Decode(&me); err != nil {
// 			log.Fatal("Failed to decode /me response:", err)
// 		}

// 		cfg.Username = me.Username
// 		if err := config.Save(cfg); err != nil {
// 			log.Fatal("Failed to save config:", err)
// 		}

// 		fmt.Println("Authentication Test Request Success!")

// 		err = keyring.Set(service, user, apiKey)
// 		if err != nil {
// 			log.Fatal("Failed to save API key:", err)
// 		}

// 		fmt.Printf("Successfully saved your API key and username (%s)!\n", me.Username)
// 	},
// }

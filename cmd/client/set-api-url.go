package main

// import (
// 	"fmt"

// 	"github.com/michaeltukdev/Potok/internal/config"
// 	"github.com/michaeltukdev/Potok/internal/prompt"
// 	"github.com/spf13/cobra"
// )

// var setApiUrlCmd = &cobra.Command{
// 	Use:   "set-api-url",
// 	Short: "Set the Potok server API URL",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		url := prompt.Input("Enter your Potok server URL (e.g., http://localhost:8080): ")

// 		cfg, err := config.Load()
// 		if err != nil {
// 			fmt.Println("Failed to load config:", err)
// 			return
// 		}

// 		cfg.APIURL = url
// 		if err := config.Save(cfg); err != nil {
// 			fmt.Println("Failed to save config:", err)
// 			return
// 		}

// 		fmt.Println("Successfully saved your API URL!")
// 	},
// }

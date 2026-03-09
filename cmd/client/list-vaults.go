package main

import (
	"github.com/spf13/cobra"
)

var listVaultsCmd = &cobra.Command{
	Use:   "list-vaults",
	Short: "List all of your synced vaults on the server",
	Run: func(cmd *cobra.Command, args []string) {
		// cfg, err := config.MustLoadWithAPIURL()
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// secret, err := keyring.Get("potok", "api-key")
		// if err != nil {
		// 	fmt.Println("Error retrieving API key:", err)
		// 	return
		// }

		// url := fmt.Sprintf("%s/users/%s/vaults", cfg.ServerURL, cfg.Username)

		// resp, err := client.MakeAuthenticatedRequest(secret, url)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer resp.Body.Close()

		// if resp.StatusCode != 200 {
		// 	fmt.Println("Unauthenticated! Please set or update your API key")
		// 	return
		// }

		// vaults, err := client.ReadVaultsFromResponse(resp)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// if len(vaults) == 0 {
		// 	fmt.Println("No vaults found on the server.")
		// 	return
		// }

		// w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		// fmt.Fprintln(w, "ID\tName\tCreated At\tUpdated At")
		// fmt.Fprintln(w, "--\t----\t----------\t----------")

		// for _, v := range vaults {
		// 	created := v.CreatedAt.Format("2006-01-02 15:04")
		// 	updated := v.UpdatedAt.Format("2006-01-02 15:04")
		// 	fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", v.ID, v.Name, created, updated)
		// }
		// w.Flush()
	},
}

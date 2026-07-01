package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"codeforge/internal/secrets"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage the secure encrypted credentials store",
}

var secretsSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Encrypt and store a secret credential",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		home, _ := os.UserHomeDir()
		storePath := filepath.Join(home, ".codeforge", "secrets.enc")

		store, err := secrets.LoadStore(storePath)
		if err != nil {
			color.Red("Error: failed to load secrets store: %v", err)
			os.Exit(1)
		}

		err = store.Set(key, value)
		if err != nil {
			color.Red("Error: failed to save secret: %v", err)
			os.Exit(1)
		}

		color.Green("CodeForge ✓  Secret %q set successfully.", key)
	},
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all encrypted credential key names",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		storePath := filepath.Join(home, ".codeforge", "secrets.enc")

		store, err := secrets.LoadStore(storePath)
		if err != nil {
			color.Red("Error: failed to load secrets store: %v", err)
			os.Exit(1)
		}

		keys := store.List()
		if len(keys) == 0 {
			fmt.Println("No secrets registered.")
			return
		}

		fmt.Println("Encrypted Secret Keys:")
		for _, k := range keys {
			fmt.Printf("  - %s\n", k)
		}
	},
}

var secretsDeleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a secret credential from the vault",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		home, _ := os.UserHomeDir()
		storePath := filepath.Join(home, ".codeforge", "secrets.enc")

		store, err := secrets.LoadStore(storePath)
		if err != nil {
			color.Red("Error: failed to load secrets store: %v", err)
			os.Exit(1)
		}

		err = store.Delete(key)
		if err != nil {
			color.Red("Error: failed to delete secret %q: %v", key, err)
			os.Exit(1)
		}

		color.Green("CodeForge ✓  Secret %q deleted successfully.", key)
	},
}

func init() {
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsListCmd)
	secretsCmd.AddCommand(secretsDeleteCmd)
	rootCmd.AddCommand(secretsCmd)
}

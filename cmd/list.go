package cmd

import (
	"github.com/jhonVitor-rs/devsetup/app"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Mostra o status das libs sem abrir a interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.List(cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

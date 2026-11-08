package cmd

import (
	"os"

	"github.com/jhonVitor-rs/devsetup/app"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devsetup",
	Short: "Reinstale seu ambiente de desenvolvimento após formatar o PC",
	Long: `devsetup detecta o seu sistema operacional e permite instalar
linguagens e ferramentas de desenvolvimento por uma única interface.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

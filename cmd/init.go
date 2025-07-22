package cmd

import (
	"fmt"

	"github.com/jackc/portreg/registry"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the registry file",
	Long:  `Initialize a new registry file with default blocked ports for common services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := registry.New(registryPath)
		if err != nil {
			return fmt.Errorf("failed to create registry: %w", err)
		}

		if err := reg.Init(); err != nil {
			return err
		}

		fmt.Printf("Initialized registry at %s\n", registryPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/jackc/portreg/registry"
	"github.com/spf13/cobra"
)

var (
	assignPort        int
	assignPath        string
	assignDescription string
)

var assignCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a port to a project",
	Long: `Assign a port to a project. If no port is specified, automatically assigns
the next available port starting from 3100.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		reg, err := registry.New(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}

		// Use current directory if no path specified
		if assignPath == "" {
			assignPath, _ = os.Getwd()
		}

		if assignPort > 0 {
			// Assign specific port
			err = reg.AssignPort(assignPort, assignDescription, assignPath)
			if err != nil {
				if errors.Is(err, registry.ErrPortAlreadyAssigned) {
					return fmt.Errorf("%w. Use 'portreg list' to see all assignments", err)
				}
				return err
			}
			if assignDescription != "" {
				fmt.Printf("Assigned port %d to %s\n", assignPort, assignDescription)
			} else {
				fmt.Printf("Assigned port %d\n", assignPort)
			}
		} else {
			// Auto-assign next available port
			port, err := reg.AssignNextAvailable(assignDescription, assignPath)
			if err != nil {
				return err
			}
			if assignDescription != "" {
				fmt.Printf("Assigned port %d to %s\n", port, assignDescription)
			} else {
				fmt.Printf("Assigned port %d\n", port)
			}
		}

		return nil
	},
}

func init() {
	assignCmd.Flags().IntVarP(&assignPort, "port", "p", 0, "Specific port to assign")
	assignCmd.Flags().StringVar(&assignPath, "path", "", "Project path (defaults to current directory)")
	assignCmd.Flags().StringVarP(&assignDescription, "description", "d", "", "Description for the port assignment")
	rootCmd.AddCommand(assignCmd)
}
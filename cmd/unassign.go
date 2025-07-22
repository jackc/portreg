package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/portreg/registry"
	"github.com/spf13/cobra"
)

var unassignCmd = &cobra.Command{
	Use:   "unassign <port>",
	Short: "Release a port assignment",
	Long:  `Release a port assignment by port number.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid port number: %s", args[0])
		}

		reg, err := registry.New(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}

		err = reg.UnassignPort(port)
		if err != nil {
			if errors.Is(err, registry.ErrPortNotAssigned) {
				return fmt.Errorf("%w. Use 'portreg list' to see all assignments", err)
			}
			return err
		}

		fmt.Printf("Unassigned port %d\n", port)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(unassignCmd)
}
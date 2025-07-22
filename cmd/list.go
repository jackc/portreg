package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jackc/portreg/registry"
	"github.com/spf13/cobra"
)

var listFormat string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Display all assigned ports",
	Long:  `Display all assigned ports in a table or JSON format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := registry.New(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}

		assignments := reg.ListAssignments()

		if listFormat == "json" {
			// JSON output
			data, err := json.MarshalIndent(assignments, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
		} else {
			// Table output
			if len(assignments) == 0 {
				fmt.Println("No ports assigned")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PORT\tDESCRIPTION\tPATH")
			fmt.Fprintln(w, "----\t-----------\t----")

			for _, a := range assignments {
				path := a.Path
				if path == "" {
					path = "-"
				}
				fmt.Fprintf(w, "%d\t%s\t%s\n", a.Port, a.Description, path)
			}

			w.Flush()
		}

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format (table or json)")
	rootCmd.AddCommand(listCmd)
}
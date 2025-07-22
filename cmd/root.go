package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var registryPath string

var rootCmd = &cobra.Command{
	Use:   "portreg",
	Short: "A port registry tool to manage port assignments",
	Long: `portreg helps developers manage port assignments across multiple projects
to avoid conflicts. It uses static port assignment stored in a JSON registry file.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	defaultPath := filepath.Join(os.Getenv("HOME"), ".portreg.json")
	rootCmd.PersistentFlags().StringVarP(&registryPath, "registry", "r", defaultPath, "Path to registry file")
}
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	formatFlag   string
	allFlag      bool
	devPathsFlag bool
	verboseFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "ddev-explain",
	Short: "Summarize DDEV project configuration",
	Long:  `A CLI tool that analyzes DDEV projects and summarizes their configuration with focus on development directories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("ddev-explain v0.1.0")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&formatFlag, "format", "f", "text", "Output format: text, json, markdown")
	rootCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Show all known DDEV projects")
	rootCmd.Flags().BoolVar(&devPathsFlag, "dev-paths", false, "Show only development paths")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show additional details")
}

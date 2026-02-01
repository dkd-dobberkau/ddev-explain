package cmd

import (
	"fmt"
	"os"

	"github.com/ochorocho/ddev-explain/internal/ddev"
	"github.com/ochorocho/ddev-explain/internal/detector"
	"github.com/ochorocho/ddev-explain/internal/finder"
	"github.com/ochorocho/ddev-explain/internal/output"
	"github.com/spf13/cobra"
)

var (
	formatFlag     string
	allFlag        bool
	devPathsFlag   bool
	verboseFlag    bool
	installCmdFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "ddev-explain",
	Short: "Summarize DDEV project configuration",
	Long:  `A CLI tool that analyzes DDEV projects and summarizes their configuration with focus on development directories.`,
	RunE:  runExplain,
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
	rootCmd.Flags().BoolVar(&installCmdFlag, "install-command", false, "Install as DDEV custom command")
}

func runExplain(cmd *cobra.Command, args []string) error {
	if installCmdFlag {
		return installDDEVCommand()
	}

	var projectPaths []string

	if allFlag {
		paths, err := finder.FindAllProjects()
		if err != nil {
			return fmt.Errorf("failed to find projects: %w", err)
		}
		projectPaths = paths
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		projectPath, err := finder.FindProjectUpward(cwd)
		if err != nil {
			return fmt.Errorf("no DDEV project found in %s or parent directories", cwd)
		}
		projectPaths = []string{projectPath}
	}

	for _, projectPath := range projectPaths {
		project, err := ddev.ParseConfig(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", projectPath, err)
			continue
		}

		// Detect dev paths
		devPaths, err := detector.DetectDevPaths(projectPath)
		if err == nil {
			project.DevPaths = devPaths
		}

		// Format output
		var formatter output.Formatter
		switch formatFlag {
		case "json":
			formatter = output.NewJSONFormatter()
		case "markdown":
			formatter = output.NewMarkdownFormatter(verboseFlag)
		default:
			formatter = output.NewTextFormatter(verboseFlag)
		}

		out, err := formatter.Format(project)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		fmt.Println(out)
	}

	return nil
}

func installDDEVCommand() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cmdDir := fmt.Sprintf("%s/.ddev/commands/host", homeDir)
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}

	cmdPath := fmt.Sprintf("%s/explain", cmdDir)
	cmdContent := `#!/bin/bash
## Description: Summarize DDEV project configuration
## Usage: explain [flags]
## Example: ddev explain --format=json

ddev-explain "$@"
`

	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0755); err != nil {
		return err
	}

	fmt.Printf("DDEV command installed: %s\n", cmdPath)
	fmt.Println("You can now use: ddev explain")
	return nil
}

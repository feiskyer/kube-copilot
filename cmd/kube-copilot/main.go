package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// global flags
	model         string
	maxTokens     int
	countTokens   bool
	verbose       bool
	maxIterations int

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "kube-copilot",
		Short: "Kubernetes Copilot powered by OpenAI",
	}
)

// init initializes the command line flags
func init() {
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "gpt-4", "OpenAI model to use")
	rootCmd.PersistentFlags().IntVarP(&maxTokens, "max-tokens", "t", 2048, "Max tokens for the GPT model")
	rootCmd.PersistentFlags().BoolVarP(&countTokens, "count-tokens", "c", false, "Print tokens count")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().IntVarP(&maxIterations, "max-iterations", "x", 10, "Max iterations for the agent running")

	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(diagnoseCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(executeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

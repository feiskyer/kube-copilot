/*
Copyright 2023 - Present, Pengfei Ni

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
		Use:     "kube-copilot",
		Version: VERSION,
		Short:   "Kubernetes Copilot powered by AI",
	}
)

// init initializes the command line flags
func init() {
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "gpt-4o", "AI model to use")
	rootCmd.PersistentFlags().IntVarP(&maxTokens, "max-tokens", "t", 2048, "Max tokens for the AI model")
	rootCmd.PersistentFlags().BoolVarP(&countTokens, "count-tokens", "c", false, "Print tokens count")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().IntVarP(&maxIterations, "max-iterations", "x", 30, "Max iterations for the agent running")

	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(diagnoseCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(executeCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

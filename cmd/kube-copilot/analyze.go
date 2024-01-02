package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/feiskyer/kube-copilot/pkg/assistants"
	"github.com/feiskyer/kube-copilot/pkg/kubernetes"
	"github.com/feiskyer/kube-copilot/pkg/llms"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const analysisSystemPrompt = `You are an expert in Kubernetes and cloud-native technologies. Your task is to use a Chain of Thought diagnostic method to transform technical analysis into explanations that are easy to understand for non-technical users. Approach this task as if you were solving a mystery or fixing a common household appliance, making each step of the process relatable and clear. Here's how to proceed:

- View each YAML data point as a crucial clue in unravelling a mystery. Lead your audience through your thought process as you uncover and identify the issue, akin to a detective piecing together a puzzle.
- When developing solutions, imagine guiding a friend through a simple DIY repair. Each step should be straightforward, using everyday language and comparisons for clarity.
- Documenting Findings and Actions in the following formats:

  - **Issue 1**:
    - **Analysis**: Describe the symptoms of Issue 1, using everyday analogies to demystify technical details.
    - **Solution**: Break down the solution into easy-to-follow steps, each described in a way that's easy for a layperson to understand.
  - **Issue 2**:
    - **Analysis**: Explain your process in discovering Issue 2, avoiding complex technical terms.
    - **Solution**: Present a clear resolution, explaining each step in a manner that's comprehensible to a non-expert.
  - More issues and solutions as needed.

Remember, your goal is to make technical information accessible and engaging to those without a background in Kubernetes or cloud-native technologies."
`

var analysisName string
var analysisNamespace string
var analysisResource string

func init() {
	analyzeCmd.PersistentFlags().StringVarP(&analysisName, "name", "", "", "Resource name")
	analyzeCmd.PersistentFlags().StringVarP(&analysisNamespace, "namespace", "n", "default", "Resource namespace")
	analyzeCmd.PersistentFlags().StringVarP(&analysisResource, "resource", "r", "pod", "Resource type")
	analyzeCmd.MarkFlagRequired("name")
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze issues for a given resource",
	Run: func(cmd *cobra.Command, args []string) {
		if analysisName == "" && len(args) > 0 {
			analysisName = args[0]
		}
		if analysisName == "" {
			fmt.Println("Please provide a resource name")
			return
		}

		fmt.Printf("Analysing %s %s/%s\n", analysisResource, analysisNamespace, analysisName)

		manifests, err := kubernetes.GetYaml(analysisResource, analysisName, analysisNamespace)
		if err != nil {
			color.Red(err.Error())
			return
		}

		manifests = llms.ConstrictPrompt(manifests, model, maxTokens)
		if verbose {
			color.Cyan("Got manifests for %s/%s:\n%s\n\n", analysisNamespace, analysisName, manifests)
		}

		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: analysisSystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("**Task**: Analyze and explain issues for the following kubernetes manifests:\n%s", manifests),
			},
		}
		response, _, err := assistants.Assistant(model, messages, maxTokens, countTokens, verbose, maxIterations)
		if err != nil {
			color.Red(err.Error())
			return
		}
		fmt.Println(response)
	},
}

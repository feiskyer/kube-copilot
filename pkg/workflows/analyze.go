package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/feiskyer/swarm-go"
)

const analysisPrompt = `As an expert on Kubernetes, your task is analyzing the given Kubernetes manifests, figure out the issues and provide solutions in a human-readable format.
For each identified issue, document the analysis and solution in everyday language, employing simple analogies to clarify technical points.

# Steps

1. **Identify Clues**: Treat each piece of YAML configuration data like a clue in a mystery. Explain how it helps to understand the issue, similar to a detective piecing together a case.
2. **Analysis with Analogies**: Translate your technical findings into relatable scenarios. Use everyday analogies to explain concepts, avoiding complex jargon. This makes episodes like 'pod failures' or 'service disruptions' simple to grasp.
3. **Solution as a DIY Guide**: Offer a step-by-step solution akin to guiding someone through a household fix-up. Instructions should be straightforward, logical, and accessible.
4. **Document Findings**:
   - Separate analysis and solution clearly for each issue, detailing them in non-technical language.

# Output Format

Provide the output in structured markdown, using clear and concise language.

# Examples

## 1. <title of the issue or potential problem>

- **Findings**: The YAML configuration doesn't specify the memory limit for the pod.
- **How to resolve**: Set memory limit in Pod spec.

## 2. HIGH Severity: CVE-2024-10963

- **Findings**: The Pod is running with CVE pam: Improper Hostname Interpretation in pam_access Leads to Access Control Bypass.
- **How to resolve**: Update package libpam-modules to fixed version (>=1.5.3) in the image. (leave the version number to empty if you don't know it)

# Notes

- Keep your language concise and simple.
- Ensure key points are included, e.g. CVE number, error code, versions.
- Relatable analogies should help in visualizing the problem and solution.
- Ensure explanations are self-contained, enough for newcomers without previous technical exposure to understand.
`

// AnalysisFlow runs a workflow to analyze Kubernetes issues and provide solutions in a human-readable format.
func AnalysisFlow(model string, manifest string, verbose bool) (string, error) {
	analysisWorkflow := &swarm.Workflow{
		Name:     "analysis-workflow",
		Model:    model,
		MaxTurns: 30,
		Verbose:  verbose,
		System:   "You are an expert on Kubernetes helping user to analyze issues and provide solutions.",
		Steps: []swarm.WorkflowStep{
			{
				Name:         "analyze",
				Instructions: analysisPrompt,
				Inputs: map[string]interface{}{
					"k8s_manifest": manifest,
				},
				Functions: []swarm.AgentFunction{kubectlFunc},
			},
		},
	}

	// Create OpenAI client
	client, err := NewSwarm()
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Initialize and run workflow
	analysisWorkflow.Initialize()
	result, _, err := analysisWorkflow.Run(context.Background(), client)
	if err != nil {
		return "", err
	}

	return result, nil
}

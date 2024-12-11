package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/feiskyer/swarm-go"
)

const generatePrompt = `As a skilled technical specialist in Kubernetes and cloud-native technologies, your task is to create Kubernetes YAML manifests by following these detailed steps:

1. Review the instructions provided to generate Kubernetes YAML manifests. Ensure that these manifests adhere to current security protocols and best practices. If an instruction lacks a specific image, choose the most commonly used one from reputable sources.
2. Utilize your expertise to scrutinize the YAML manifests. Conduct a thorough step-by-step analysis to identify any issues. Resolve these issues, ensuring the YAML manifests are accurate and secure.
3. After fixing and verifying the manifests, compile them in their raw form. For multiple YAML files, use '---' as a separator.

# Steps

1. **Understand the Instructions:**
   - Evaluate the intended use and environment for each manifest as per instructions provided.

2. **Security and Best Practices Assessment:**
   - Assess the security aspects of each component, ensuring alignment with current standards and best practices.
   - Perform a comprehensive analysis of the YAML structure and configurations.

3. **Document and Address Discrepancies:**
   - Document and justify any discrepancies or issues you find, in a sequential manner.
   - Implement robust solutions that enhance the manifests' performance and security, utilizing best practices and recommended images.

4. **Finalize the YAML Manifests:**
   - Ensure the final manifests are syntactically correct, properly formatted, and deployment-ready.

# Output Format

- Present only the final YAML manifests in raw format, separated by "---" for multiple files.
- Exclude any comments or additional annotations within the YAML files.

Your expertise ensures these manifests are not only functional but also compliant with the highest standards in Kubernetes and cloud-native technologies.`

// GeneratorFlow runs a workflow to generate Kubernetes YAML manifests based on the provided instructions.
func GeneratorFlow(model string, instructions string, verbose bool) (string, error) {
	generatorWorkflow := &swarm.Workflow{
		Name:     "generator-workflow",
		Model:    model,
		MaxTurns: 30,
		Verbose:  verbose,
		System:   "You are an expert on Kubernetes helping user to generate Kubernetes YAML manifests.",
		Steps: []swarm.WorkflowStep{
			{
				Name:         "generator",
				Instructions: generatePrompt,
				Inputs: map[string]interface{}{
					"instructions": instructions,
				},
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
	generatorWorkflow.Initialize()
	result, _, err := generatorWorkflow.Run(context.Background(), client)
	if err != nil {
		return "", err
	}

	return result, nil
}

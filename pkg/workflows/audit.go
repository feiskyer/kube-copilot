package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/feiskyer/swarm-go"
)

const auditPrompt = `Conduct a structured security audit of a Kubernetes environment using a Chain of Thought (CoT) approach, ensuring each technical step is clearly connected to solutions with easy-to-understand explanations.

## Plan of Action

**1. Security Auditing:**
   - **Retrieve Pod Configuration:**
      - Use "kubectl get -n {namespace} pod {pod} -o yaml" to obtain pod YAML configuration.
      - **Explain YAML:**
        - Breakdown what YAML is and its importance in understanding a pod's security posture, using analogies for clarity.

   - **Analyze YAML for Misconfigurations:**
      - Look for common security misconfigurations or risky settings within the YAML.
      - Connect issues to relatable concepts for non-technical users (e.g., likening insecure settings to an unlocked door).

**2. Vulnerability Scanning:**
   - **Extract and Scan Image:**
      - Extract the container image from the YAML configuration obtained during last step.
      - Perform a scan using "trivy image <image>".
      - Summerize Vulnerability Scans results with CVE numbers, severity, and descriptions.

**3. Issue Identification and Solution Formulation:**
   - Document each issue clearly and concisely.
   - Provide the recommendations to fix each issue.

## Provide the output in structured markdown, using clear and concise language.

Example output:

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

// AuditFlow conducts a structured security audit of a Kubernetes Pod.
func AuditFlow(model string, namespace string, name string, verbose bool) (string, error) {
	auditWorkflow := &swarm.Workflow{
		Name:     "audit-workflow",
		Model:    model,
		MaxTurns: 30,
		Verbose:  verbose,
		System:   "You are an expert on Kubernetes helping user to audit the security issues for a given Pod.",
		Steps: []swarm.WorkflowStep{
			{
				Name:         "audit",
				Instructions: auditPrompt,
				Inputs: map[string]interface{}{
					"pod_namespace": namespace,
					"pod_name":      name,
				},
				Functions: []swarm.AgentFunction{trivyFunc, kubectlFunc},
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
	auditWorkflow.Initialize()
	result, _, err := auditWorkflow.Run(context.Background(), client)
	if err != nil {
		return "", err
	}

	return result, nil
}

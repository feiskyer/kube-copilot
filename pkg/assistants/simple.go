package assistants

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/feiskyer/kube-copilot/pkg/llms"
	"github.com/feiskyer/kube-copilot/pkg/tools"
	"github.com/sashabaranov/go-openai"
)

const (
	defaultMaxIterations = 10
)

// Assistant is the simplest AI assistant.
func Assistant(model string, prompts []openai.ChatCompletionMessage, maxTokens int, countTokens bool, verbose bool, maxIterations int) (result string, chatHistory []openai.ChatCompletionMessage, err error) {
	chatHistory = prompts
	if len(prompts) == 0 {
		return "", nil, fmt.Errorf("prompts cannot be empty")
	}

	client, err := llms.NewOpenAIClient()
	if err != nil {
		return "", nil, fmt.Errorf("unable to get OpenAI client: %v", err)
	}

	defer func() {
		if countTokens {
			count := llms.NumTokensFromMessages(chatHistory, model)
			color.Green("Total tokens: %d\n\n", count)
		}
	}()

	if verbose {
		color.Blue("Iteration 1): chatting with LLM\n")
	}

	resp, err := client.Chat(model, maxTokens, chatHistory)
	if err != nil {
		return "", chatHistory, fmt.Errorf("chat completion error: %v", err)
	}

	chatHistory = append(chatHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: string(resp),
	})

	if verbose {
		color.Cyan("Initial response from LLM:\n%s\n\n", resp)
	}

	var toolPrompt tools.ToolPrompt
	if err = json.Unmarshal([]byte(resp), &toolPrompt); err != nil {
		if verbose {
			color.Cyan("Unable to parse tool from prompt, assuming got final answer.\n\n", resp)
		}
		return resp, chatHistory, nil
	}

	iterations := 0
	if maxIterations <= 0 {
		maxIterations = defaultMaxIterations
	}
	for {
		iterations++

		if verbose {
			color.Cyan("Thought: %s\n\n", toolPrompt.Thought)
		}

		if iterations > maxIterations {
			color.Red("Max iterations reached")
			return toolPrompt.FinalAnswer, chatHistory, nil
		}

		if toolPrompt.FinalAnswer != "" {
			if verbose {
				color.Cyan("Final answer: %s\n\n", toolPrompt.FinalAnswer)
			}
			return toolPrompt.FinalAnswer, chatHistory, nil
		}

		if toolPrompt.Action.Name != "" {
			if verbose {
				color.Blue("Iteration %d): executing tool %s\n", iterations, toolPrompt.Action.Name)
				color.Cyan("Invoking %s tool with inputs: \n============\n%s\n============\n\n", toolPrompt.Action.Name, toolPrompt.Action.Input)
			}
			ret, err := tools.CopilotTools[toolPrompt.Action.Name](toolPrompt.Action.Input)
			observation := strings.TrimSpace(ret)
			if err != nil {
				observation = fmt.Sprintf("Tool %s failed with error %s. Considering refine the inputs for the tool.", toolPrompt.Action.Name, ret)
			}
			if verbose {
				color.Cyan("Observation: %s\n\n", observation)
			}

			// Constrict the prompt to the max tokens allowed by the model.
			// This is required because the tool may have generated a long output.
			observation = llms.ConstrictPrompt(observation, model, 1024)
			toolPrompt.Observation = observation
			assistantMessage, _ := json.Marshal(toolPrompt)
			chatHistory = append(chatHistory, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: string(assistantMessage),
			})
			// Constrict the chat history to the max tokens allowed by the model.
			// This is required because the chat history may have grown too large.
			chatHistory = llms.ConstrictMessages(chatHistory, model, maxTokens)

			// Start next iteration of LLM chat.
			if verbose {
				color.Blue("Iteration %d): chatting with LLM\n", iterations)
			}

			resp, err := client.Chat(model, maxTokens, chatHistory)
			if err != nil {
				return "", chatHistory, fmt.Errorf("chat completion error: %v", err)
			}

			chatHistory = append(chatHistory, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: string(resp),
			})
			if verbose {
				color.Cyan("Intermediate response from LLM: %s\n\n", resp)
			}

			// extract the tool prompt from the LLM response.
			if err = json.Unmarshal([]byte(resp), &toolPrompt); err != nil {
				if verbose {
					color.Cyan("Unable to parse tools from LLM, summarizing the final answer.\n\n")
				}

				chatHistory = append(chatHistory, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "Summarize all the chat history and respond to original question with final answer",
				})

				resp, err = client.Chat(model, maxTokens, chatHistory)
				if err != nil {
					return "", chatHistory, fmt.Errorf("chat completion error: %v", err)
				}

				return resp, chatHistory, nil
			}
		}
	}
}

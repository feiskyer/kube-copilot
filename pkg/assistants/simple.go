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

// Package assistants provides a simple AI assistant using OpenAI's GPT models.
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

// ToolPrompt is the JSON format for the prompt.
type ToolPrompt struct {
	Question string `json:"question"`
	Thought  string `json:"thought,omitempty"`
	Action   struct {
		Name  string `json:"name"`
		Input string `json:"input"`
	} `json:"action,omitempty"`
	Observation string `json:"observation,omitempty"`
	FinalAnswer string `json:"final_answer,omitempty"`
}

// Assistant is the simplest AI assistant.
// Deprecated: Use ReActFlow instead.
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

	var toolPrompt ToolPrompt
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
				color.Cyan("Final answer: %v\n\n", toolPrompt.FinalAnswer)
			}
			return toolPrompt.FinalAnswer, chatHistory, nil
		}

		if toolPrompt.Action.Name != "" {
			var observation string
			if verbose {
				color.Blue("Iteration %d): executing tool %s\n", iterations, toolPrompt.Action.Name)
				color.Cyan("Invoking %s tool with inputs: \n============\n%s\n============\n\n", toolPrompt.Action.Name, toolPrompt.Action.Input)
			}
			if toolFunc, ok := tools.CopilotTools[toolPrompt.Action.Name]; ok {
				ret, err := toolFunc(toolPrompt.Action.Input)
				observation = strings.TrimSpace(ret)
				if err != nil {
					observation = fmt.Sprintf("Tool %s failed with error %s. Considering refine the inputs for the tool.", toolPrompt.Action.Name, ret)
				}
			} else {
				observation = fmt.Sprintf("Tool %s is not available. Considering switch to other supported tools.", toolPrompt.Action.Name)
			}
			if verbose {
				color.Cyan("Observation: %s\n\n", observation)
			}

			// Constrict the prompt to the max tokens allowed by the model.
			// This is required because the tool may have generated a long output.
			observation = llms.ConstrictPrompt(observation, model)
			toolPrompt.Observation = observation
			assistantMessage, _ := json.Marshal(toolPrompt)
			chatHistory = append(chatHistory, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: string(assistantMessage),
			})
			// Constrict the chat history to the max tokens allowed by the model.
			// This is required because the chat history may have grown too large.
			chatHistory = llms.ConstrictMessages(chatHistory, model)

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
					color.Cyan("Unable to parse tools from LLM (%s), summarizing the final answer.\n\n", err.Error())
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

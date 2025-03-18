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

package llms

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

var tokenLimitsPerModel = map[string]int{
	"code-davinci-002":       4096,
	"gpt-3.5-turbo-0301":     4096,
	"gpt-3.5-turbo-0613":     4096,
	"gpt-3.5-turbo-1106":     16385,
	"gpt-3.5-turbo-16k-0613": 16385,
	"gpt-3.5-turbo-16k":      16385,
	"gpt-3.5-turbo-instruct": 4096,
	"gpt-3.5-turbo":          4096,
	"gpt-4-0314":             8192,
	"gpt-4-0613":             8192,
	"gpt-4-1106-preview":     128000,
	"gpt-4-32k-0314":         32768,
	"gpt-4-32k-0613":         32768,
	"gpt-4-32k":              32768,
	"gpt-4-vision-preview":   128000,
	"gpt-4":                  8192,
	"gpt-4-turbo":            128000,
	"text-davinci-002":       4096,
	"text-davinci-003":       4096,
	"gpt-4o":                 128000,
	"gpt-4o-mini":            128000,
	"o1-mini":                128000,
	"o3-mini":                200000,
	"o1":                     200000,
}

// GetTokenLimits returns the maximum number of tokens for the given model.
func GetTokenLimits(model string) int {
	model = strings.ToLower(model)
	if maxTokens, ok := tokenLimitsPerModel[model]; ok {
		return maxTokens
	}

	return 8192
}

// NumTokensFromMessages returns the number of tokens in the given messages.
// OpenAI Cookbook: https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
func NumTokensFromMessages(messages []openai.ChatCompletionMessage, model string) (numTokens int) {
	encodingModel := model
	if model == "o1-mini" || model == "o3-mini" || model == "o1" || model == "o3" {
		encodingModel = "gpt-4o"
	}
	tkm, err := tiktoken.EncodingForModel(encodingModel)
	if err != nil {
		err = fmt.Errorf("encoding for model: %v", err)
		log.Println(err)
		return
	}

	tokensPerMessage := 3
	tokensPerName := 1
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1   // if there's a name, the role is omitted
	}

	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(message.Content, nil, nil))
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		if message.Name != "" {
			numTokens += len(tkm.Encode(message.Name, nil, nil))
			numTokens += tokensPerName
		}
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens
}

// ConstrictMessages returns the messages that fit within the token limit.
func ConstrictMessages(messages []openai.ChatCompletionMessage, model string) []openai.ChatCompletionMessage {
	tokenLimits := GetTokenLimits(model)

	for {
		numTokens := NumTokensFromMessages(messages, model)
		if numTokens <= tokenLimits {
			return messages
		}

		// If no messages, return empty
		if len(messages) == 0 {
			return messages
		}

		// If only one message or we can't reduce further
		if len(messages) <= 1 {
			return messages
		}

		// When over the limit, try keeping only the system prompt (first message)
		// and the most recent message
		if len(messages) > 2 {
			// Try with just system and last message
			systemAndLatest := []openai.ChatCompletionMessage{
				messages[0],
				messages[len(messages)-1],
			}
			messages = systemAndLatest
		} else {
			// We have exactly 2 messages and they're still over the limit
			// Keep only the first message (usually system prompt)
			messages = messages[:1]
		}
	}
}

// ConstrictPrompt returns the prompt that fits within the token limit.
func ConstrictPrompt(prompt string, model string) string {
	tokenLimits := GetTokenLimits(model)

	for {
		numTokens := NumTokensFromMessages([]openai.ChatCompletionMessage{{Content: prompt}}, model)
		if numTokens < tokenLimits {
			return prompt
		}

		// Remove the first third percent lines
		lines := strings.Split(prompt, "\n")
		lines = lines[int64(math.Ceil(float64(len(lines))/3)):]
		prompt = strings.Join(lines, "\n")

		if strings.TrimSpace(prompt) == "" {
			return ""
		}
	}
}

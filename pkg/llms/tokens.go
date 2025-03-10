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

	return 4096
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

	var tokensPerMessage, tokensPerName int
	switch model {
	case "gpt-3.5-turbo-0613",
		"gpt-3.5-turbo-16k-0613",
		"gpt-4-0314",
		"gpt-4-32k-0314",
		"gpt-4-0613",
		"gpt-4-32k-0613",
		"gpt-4o",
		"gpt-4o-mini",
		"o1-mini",
		"o3-mini",
		"o1":
		tokensPerMessage = 3
		tokensPerName = 1
	case "gpt-3.5-turbo-0301":
		tokensPerMessage = 4 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1   // if there's a name, the role is omitted
	default:
		if strings.Contains(model, "gpt-3.5-turbo") {
			return NumTokensFromMessages(messages, "gpt-3.5-turbo-0613")
		} else if strings.Contains(model, "gpt-4") {
			return NumTokensFromMessages(messages, "gpt-4-0613")
		} else {
			err = fmt.Errorf("num_tokens_from_messages() is not implemented for model %s. See https://github.com/openai/openai-python/blob/main/chatml.md for information on how messages are converted to tokens", model)
			log.Println(err)
			return
		}
	}

	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(message.Content, nil, nil))
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		numTokens += len(tkm.Encode(message.Name, nil, nil))
		if message.Name != "" {
			numTokens += tokensPerName
		}
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens
}

// ConstrictMessages returns the messages that fit within the token limit.
func ConstrictMessages(messages []openai.ChatCompletionMessage, model string, maxTokens int) []openai.ChatCompletionMessage {
	tokenLimits := GetTokenLimits(model)
	if maxTokens >= tokenLimits {
		return nil
	}

	for {
		numTokens := NumTokensFromMessages(messages, model)
		if numTokens+maxTokens < tokenLimits {
			return messages
		}

		// Remove the oldest message (keep the first one as it is usually the system prompt)
		messages = append(messages[:1], messages[2:]...)
	}
}

// ConstrictPrompt returns the prompt that fits within the token limit.
func ConstrictPrompt(prompt string, model string, tokenLimits int) string {
	for {
		numTokens := NumTokensFromMessages([]openai.ChatCompletionMessage{{Content: prompt}}, model)
		if numTokens < tokenLimits {
			return prompt
		}

		// Remove the first thrid percent lines
		lines := strings.Split(prompt, "\n")
		lines = lines[int64(math.Ceil(float64(len(lines))/3)):]
		prompt = strings.Join(lines, "\n")

		if strings.TrimSpace(prompt) == "" {
			return ""
		}
	}
}

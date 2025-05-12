/*
Copyright 2025 - Present, Pengfei Ni

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
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPConfig is the configuration for the MCP server.
type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer is the configuration for a single MCP server.
type MCPServer struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	URL     string            `json:"url,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

// MCPTool is a tool that uses the MCP protocol.
type MCPTool struct {
	name        string
	description string
	inputSchema string
	toolFunc    func(input string) (string, error)
}

// Description returns the description of the tool.
func (t MCPTool) Description() string {
	return t.description
}

// InputSchema returns the input schema for the tool.
func (t MCPTool) InputSchema() string {
	return t.inputSchema
}

// ToolFunc is the function that will be called to execute the tool.
func (t MCPTool) ToolFunc(input string) (string, error) {
	return t.toolFunc(input)
}

// GetMCPTools returns the MCP tools.
func GetMCPTools(configFile string, verbose bool) (map[string]Tool, map[string]client.MCPClient, error) {
	var config MCPConfig
	mcpTools := make(map[string]Tool)
	mcpClients := make(map[string]client.MCPClient)

	// Read the config file
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the config file
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Create a client for each MCP server
	for name, server := range config.MCPServers {
		if verbose {
			color.Green("Creating client for %s", name)
		}

		var c client.MCPClient
		envs := make([]string, 0)
		for _, env := range os.Environ() {
			envs = append(envs, env)
		}
		for k, v := range server.Env {
			envs = append(envs, fmt.Sprintf("%s=%s", k, v))
		}
		if server.Command != "" || server.Type == "stdio" {
			c, err = client.NewStdioMCPClient(
				server.Command,
				envs,
				server.Args...,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create client for %s: %v", name, err)
			}
		} else if server.URL != "" {
			if strings.Contains(server.URL, "sse") || strings.EqualFold(server.Type, "sse") {
				c, err = client.NewSSEMCPClient(server.URL, transport.WithHeaders(server.Headers), transport.WithHTTPClient(&http.Client{Timeout: time.Duration(server.Timeout) * time.Second}))
			} else {
				c, err = client.NewStreamableHttpClient(server.URL, transport.WithHTTPHeaders(server.Headers), transport.WithHTTPTimeout(time.Duration(server.Timeout)*time.Second))
			}
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create client for %s: %v", name, err)
			}
		} else {
			return nil, nil, fmt.Errorf("no command or URL specified for %s", name)
		}

		tools, err := createMCPTools(c, name, verbose)
		if err != nil {
			c.Close()
			return nil, nil, fmt.Errorf("failed to create client for %s: %v", name, err)
		}

		mcpClients[name] = c
		for toolName, tool := range tools {
			mcpTools[toolName] = tool
		}
	}

	return mcpTools, mcpClients, nil
}

func createMCPTools(c client.MCPClient, name string, verbose bool) (map[string]Tool, error) {
	if verbose {
		color.Green("Initializing client for %s", name)
	}
	mcpTools := make(map[string]Tool)

	// Initialize the client
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	_, err := c.Initialize(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize: %v", err)
	}

	// List tools
	if verbose {
		color.Green("Listing tools for %s", name)
	}
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := c.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %v", err)
	}

	for _, tool := range tools.Tools {
		toolName := fmt.Sprintf("%s_%s", name, tool.Name)
		inputSchema, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input schema: %v", err)
		}

		mcpTools[toolName] = MCPTool{
			name:        toolName,
			description: tool.Description,
			inputSchema: fmt.Sprintf("JSON Schema: %s", inputSchema),
			toolFunc: func(input string) (string, error) {
				callRequest := mcp.CallToolRequest{}
				callRequest.Params.Name = tool.Name

				var args map[string]interface{}
				err := json.Unmarshal([]byte(input), &args)
				if err != nil {
					if len(tool.InputSchema.Required) > 0 {
						args[tool.InputSchema.Required[0]] = input
					} else {
						args["input"] = input
					}
				}
				callRequest.Params.Arguments = args

				callResult, err := c.CallTool(context.Background(), callRequest)
				if err != nil {
					return "", fmt.Errorf("failed to call tool %s: %v", tool.Name, err)
				}
				// color.Green("Got tool %s result: %q", toolName, callResult.Content)

				var contentStrings []string
				for _, content := range callResult.Content {
					if textContent, ok := content.(mcp.TextContent); ok {
						contentStrings = append(contentStrings, textContent.Text)
					}
				}
				return strings.Join(contentStrings, "\n"), nil
			},
		}
	}
	return mcpTools, nil
}

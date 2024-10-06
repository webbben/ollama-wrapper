package ollamawrapper

import (
	"context"
	"os/exec"

	"github.com/ollama/ollama/api"
)

/*
Documentation

API: https://github.com/ollama/ollama/blob/main/docs/api.md

Parameters: https://github.com/ollama/ollama/blob/main/docs/modelfile.md#valid-parameters-and-values
*/

const (
	model = "llama3"
)

// starts the ollama server, and returns its Cmd reference so the process can be managed later
func StartServer() (*exec.Cmd, error) {
	cmd := exec.Command("ollama", "serve")
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// stops the ollama server, killing the process
func StopServer(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}

func GetClient() (*api.Client, error) {
	return api.ClientFromEnvironment()
}

// Call the Chat Completion API. Meant for conversations where past message context is needed.
func ChatCompletion(client *api.Client, messages []api.Message) ([]api.Message, error) {
	stream := false
	ctx := context.Background()
	req := &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   &stream,
	}

	err := client.Chat(ctx, req, func(cr api.ChatResponse) error {
		messages = append(messages, cr.Message)
		return nil
	})
	if err != nil {
		return []api.Message{}, err
	}
	return messages, nil
}

// Generate a completion using custom options. Below are some common options, but find more information about options params here:
//
// https://github.com/ollama/ollama/blob/main/docs/modelfile.md#valid-parameters-and-values
//
// "temperature": float (default: 0.8) - increasing this will make the model answer more creatively
func GenerateCompletionWithOpts(client *api.Client, systemPrompt string, prompt string, opts map[string]interface{}) (string, error) {
	stream := false
	req := &api.GenerateRequest{
		Model:   model,
		Prompt:  prompt,
		System:  systemPrompt,
		Stream:  &stream,
		Options: opts,
	}
	return generateCompletion(client, req)
}

// Generates a completion using the given system prompt to set the context and AI behavior/personality, and based on the given prompt.
//
// Use ChatCompletion for conversations and memory based generation.
//
// Use GenerateCompletionWithOpts to customize options such as temperature.
func GenerateCompletion(client *api.Client, systemPrompt string, prompt string) (string, error) {
	stream := false
	req := &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		System: systemPrompt,
		Stream: &stream,
	}
	return generateCompletion(client, req)
}

func generateCompletion(client *api.Client, req *api.GenerateRequest) (string, error) {
	ctx := context.Background()
	output := ""
	err := client.Generate(ctx, req, func(gr api.GenerateResponse) error {
		output = gr.Response
		return nil
	})
	if err != nil {
		return "", err
	}
	return output, nil
}

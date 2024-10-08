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
	responseFunc := func(cr api.ChatResponse) error {
		messages = append(messages, cr.Message)
		return nil
	}
	err := chatCompletion(client, messages, false, responseFunc)
	return messages, err
}

func ChatCompletionStream(client *api.Client, messages []api.Message, responseFunc func(cr api.ChatResponse) error) ([]api.Message, error) {
	var nextMessage *api.Message = nil
	respFunc := func(cr api.ChatResponse) error {
		if nextMessage == nil {
			nextMessage = &cr.Message // initialize message content based on first stream message
		} else {
			nextMessage.Content += cr.Message.Content // combine the incoming stream to get the full next message
		}
		return responseFunc(cr)
	}

	err := chatCompletion(client, messages, true, respFunc)
	messages = append(messages, *nextMessage)
	return messages, err
}

func chatCompletion(client *api.Client, messages []api.Message, stream bool, responseFunc func(cr api.ChatResponse) error) error {
	ctx := context.Background()
	req := &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   &stream,
	}
	return client.Chat(ctx, req, responseFunc)
}

// Generate a completion using custom options. Below are some common options, but find more information about options params here:
//
// https://github.com/ollama/ollama/blob/main/docs/modelfile.md#valid-parameters-and-values
//
// "temperature": float (default: 0.8) - increasing this will make the model answer more creatively
func GenerateCompletionWithOpts(client *api.Client, systemPrompt string, prompt string, opts map[string]interface{}) (string, error) {
	response := ""
	responseFunc := func(gr api.GenerateResponse) error {
		response += gr.Response
		return nil
	}
	err := generateCompletion(client, prompt, systemPrompt, false, responseFunc, opts)
	return response, err
}

// Generates a completion using the given system prompt to set the context and AI behavior/personality, and based on the given prompt.
//
// Use ChatCompletion for conversations and memory based generation.
//
// Use GenerateCompletionWithOpts to customize options such as temperature.
func GenerateCompletion(client *api.Client, systemPrompt string, prompt string) (string, error) {
	response := ""
	responseFunc := func(gr api.GenerateResponse) error {
		response += gr.Response
		return nil
	}
	err := generateCompletion(client, prompt, systemPrompt, false, responseFunc, nil)
	return response, err
}

func GenerateCompletionStream(client *api.Client, systemPrompt string, prompt string, responseFunc func(gr api.GenerateResponse) error) (string, error) {
	response := ""
	respFunc := func(cr api.GenerateResponse) error {
		response += cr.Response
		return responseFunc(cr)
	}

	err := generateCompletion(client, prompt, systemPrompt, true, respFunc, nil)
	return response, err
}

func generateCompletion(client *api.Client, prompt, sysPrompt string, stream bool, responseFunc func(gr api.GenerateResponse) error, opts map[string]interface{}) error {
	ctx := context.Background()
	req := &api.GenerateRequest{
		Model:   model,
		Prompt:  prompt,
		System:  sysPrompt,
		Stream:  &stream,
		Options: opts,
	}
	return client.Generate(ctx, req, responseFunc)
}

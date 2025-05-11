package ollamawrapper

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/ollama/ollama/api"
	"github.com/shirou/gopsutil/v3/process"
)

/*
Documentation

API: https://github.com/ollama/ollama/blob/main/docs/api.md

Parameters: https://github.com/ollama/ollama/blob/main/docs/modelfile.md#valid-parameters-and-values
*/

var (
	model = "llama3"
)

// SetModel lets you choose a specific model to use. By default, llama3 is used.
//
// Note: Ensure the model has been pulled already before, or else an error may occur.
func SetModel(newModel string) {
	model = newModel
}

// EnsureModelIsPulled ensures that the given model is available locally, and if not, triggers a pull request
func EnsureModelIsPulled(model string, stream bool, fn func(PullRequestProgress)) error {
	models, err := GetModels()
	if err != nil {
		return errors.Join(errors.New("EnsureModelIsPulled: failed to get local models;"), err)
	}

	for _, m := range models {
		// model already exists; continue without pulling
		if m.Name == model {
			return nil
		}
	}

	// model not found; trigger pull request
	return PullModel(model, stream, fn)
}

// starts the ollama server (or finds its process if already running), and returns its PID
func StartServer() (int32, error) {
	pid, err := GetOllamaPID()
	if err != nil {
		return -1, errors.Join(errors.New("StartServer: failed to get ollama PID;"), err)
	}
	if pid != -1 {
		return pid, nil
	}

	// ollama isn't running yet, so start it ourselves
	cmd := exec.Command("ollama", "serve")

	err = cmd.Start()
	if err != nil {
		return -1, err
	}

	if cmd.Process == nil {
		return -1, errors.New("StartServer: failed to get process after running start command")
	}

	return int32(cmd.Process.Pid), nil
}

func GetClient() (*api.Client, error) {
	return api.ClientFromEnvironment()
}

func GetModel() string {
	return model
}

func GetVersion() (string, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}
	return client.Version(context.Background())
}

// GetOllamaPID gets the PID of the first process it finds that has the name "ollama".
//
// If no process is found, returns -1.
func GetOllamaPID() (int32, error) {
	processes, err := process.Processes()
	if err != nil {
		return -1, errors.Join(errors.New("GetOllamaPID: failed to get processes;"), err)
	}

	for _, process := range processes {
		name, err := process.Name()
		if err != nil {
			continue
		}
		if name == "ollama" {
			return process.Pid, nil
		}
	}

	return -1, nil
}

type PullRequestProgress struct {
	Status    string
	Digest    string
	Total     int64
	Completed int64
}

func PullModel(modelName string, stream bool, fn func(PullRequestProgress)) error {
	client, err := GetClient()
	if err != nil {
		return errors.Join(errors.New("PullModel: failed to get client;"), err)
	}

	ctx := context.Background()

	req := api.PullRequest{
		Model:  modelName,
		Stream: &stream,
	}

	// since applications using this module as a dependency can't use ollama's base package and types directly,
	// I made wrappers around this functionality so that it can be passed down to the consuming application.
	var progressFn api.PullProgressFunc = func(pr api.ProgressResponse) error {
		fn(PullRequestProgress{
			Status:    pr.Status,
			Digest:    pr.Digest,
			Total:     pr.Total,
			Completed: pr.Completed,
		})
		return nil
	}

	return client.Pull(ctx, &req, progressFn)
}

type Model struct {
	Name   string
	Digest string
	Size   int64
}

// GetModels gets all models that are available locally
func GetModels() ([]Model, error) {
	client, err := GetClient()
	if err != nil {
		return nil, errors.Join(errors.New("GetModels: failed to get client;"), err)
	}

	resp, err := client.List(context.Background())
	if err != nil {
		return nil, errors.Join(errors.New("GetModels: error getting model list;"), err)
	}
	if resp == nil {
		return nil, errors.New("GetModels: model list response is nil")
	}

	models := make([]Model, len(resp.Models))
	for i, m := range resp.Models {
		models[i] = Model{
			Name:   m.Model,
			Digest: m.Digest,
			Size:   m.Size,
		}
	}
	return models, nil
}

// Call the Chat Completion API. Meant for conversations where past message context is needed.
func ChatCompletion(client *api.Client, chatMessages []Message) ([]Message, error) {
	messages := convertWrapperMessages(chatMessages)
	responseFunc := func(cr api.ChatResponse) error {
		messages = append(messages, cr.Message)
		return nil
	}
	err := chatCompletion(client, messages, false, responseFunc)

	return wrapMessages(messages), err
}

var ErrMessageEmpty error = errors.New("error generating completion: empty response")

func ChatCompletionStream(client *api.Client, chatMessages []Message, responseFunc func(cr ChatResponse) error) ([]Message, error) {
	messages := convertWrapperMessages(chatMessages)
	var nextMessage *api.Message = nil
	respFunc := func(cr api.ChatResponse) error {
		if nextMessage == nil {
			nextMessage = &cr.Message // initialize message content based on first stream message
		} else {
			nextMessage.Content += cr.Message.Content // combine the incoming stream to get the full next message
		}
		return responseFunc(wrapChatResponse(cr))
	}

	err := chatCompletion(client, messages, true, respFunc)
	if err != nil {
		return []Message{}, errors.Join(errors.New("ChatCompletionStream: error generating completion;"), err)
	}
	if nextMessage == nil {
		return []Message{}, ErrMessageEmpty
	}

	messages = append(messages, *nextMessage)
	return wrapMessages(messages), err
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

// Generate a completion using custom options, and specifying a specific json output format.
//
// Below are some common options, but find more information about options params here:
//
// https://github.com/ollama/ollama/blob/main/docs/modelfile.md#valid-parameters-and-values
//
// "temperature": float (default: 0.8) - increasing this will make the model answer more creatively
func GenerateCompletionOptsFormat(client *api.Client, systemPrompt string, prompt string, opts map[string]interface{}, format json.RawMessage) (string, error) {
	response := ""
	responseFunc := func(gr api.GenerateResponse) error {
		response += gr.Response
		return nil
	}
	err := generateCompletionFormatted(client, prompt, systemPrompt, false, responseFunc, opts, format)
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

func GenerateCompletionStream(client *api.Client, systemPrompt string, prompt string, responseFunc func(gr GenerateResponse) error) (string, error) {
	response := ""
	respFunc := func(gr api.GenerateResponse) error {
		response += gr.Response
		return responseFunc(wrapGenerateResponse(gr))
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

func generateCompletionFormatted(client *api.Client, prompt, sysPrompt string, stream bool, responseFunc func(gr api.GenerateResponse) error, opts map[string]interface{}, format json.RawMessage) error {
	ctx := context.Background()
	req := &api.GenerateRequest{
		Model:   model,
		Prompt:  prompt,
		System:  sysPrompt,
		Stream:  &stream,
		Options: opts,
		Format:  format,
	}
	return client.Generate(ctx, req, responseFunc)
}

package ollamawrapper

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/ollama/ollama/api"
)

// run in terminal:
// go test -run ^TestChatCompletion$ github.com/webbben/ollama-wrapper
func TestChatCompletion(t *testing.T) {
	cmd, err := StartServer()
	if err != nil {
		log.Fatal("failed to start ollama server:", err)
	}
	defer StopServer(cmd)

	client, err := GetClient()
	if err != nil {
		log.Fatal("failed to get ollama client:", err)
	}

	// test the ChatCompletion function
	messages := []api.Message{
		{
			Role:    "user",
			Content: "Say the word 'llama' to me, and nothing else.",
		},
	}
	res, err := ChatCompletion(client, messages)
	if err != nil {
		log.Fatal("failed to get chat completion:", err)
	}
	if len(res) < 2 {
		log.Fatal("chat completion returned no messages")
	}
	if !strings.Contains(strings.ToLower(res[1].Content), "llama") {
		log.Fatalf("chat completion returned unexpected message: %s", res[1].Content)
	}
}

func TestChatCompletionStream(t *testing.T) {
	cmd, err := StartServer()
	if err != nil {
		log.Fatal("failed to start ollama server:", err)
	}
	defer StopServer(cmd)

	client, err := GetClient()
	if err != nil {
		log.Fatal("failed to get ollama client:", err)
	}

	// test the ChatCompletionStream function
	messages := []api.Message{
		{
			Role:    "user",
			Content: "Say the word 'llama' to me, and nothing else.",
		},
	}
	res, err := ChatCompletionStream(client, messages, func(cr api.ChatResponse) error {
		fmt.Print(cr.Message.Content)
		return nil
	})
	if err != nil {
		log.Fatal("failed to get chat completion stream:", err)
	}
	if len(res) < 2 {
		log.Fatal("chat completion returned no messages")
	}
	if !strings.Contains(strings.ToLower(res[1].Content), "llama") {
		log.Fatalf("chat completion returned unexpected message: %s", res[1].Content)
	}
}

func TestGenerateCompletion(t *testing.T) {
	cmd, err := StartServer()
	if err != nil {
		log.Fatal("failed to start ollama server:", err)
	}
	defer StopServer(cmd)

	client, err := GetClient()
	if err != nil {
		log.Fatal("failed to get ollama client:", err)
	}

	// test the GenerateCompletion function
	sysPrompt := "repeat after me: "
	prompt := "llama"
	res, err := GenerateCompletion(client, sysPrompt, prompt)
	if err != nil {
		log.Fatal("failed to generate completion:", err)
	}
	if res == "" {
		log.Fatal("generated completion is empty")
	}
	if !strings.Contains(strings.ToLower(res), "llama") {
		log.Fatalf("generated completion returned unexpected message: %s", res)
	}
}

func TestGenerateCompletionStream(t *testing.T) {
	cmd, err := StartServer()
	if err != nil {
		log.Fatal("failed to start ollama server:", err)
	}
	defer StopServer(cmd)

	client, err := GetClient()
	if err != nil {
		log.Fatal("failed to get ollama client:", err)
	}

	// test the GenerateCompletionStream function
	sysPrompt := "repeat after me: "
	prompt := "llama"
	res, err := GenerateCompletionStream(client, sysPrompt, prompt, func(gr api.GenerateResponse) error {
		fmt.Print(gr.Response)
		return nil
	})
	if err != nil {
		log.Fatal("failed to generate completion stream:", err)
	}
	if res == "" {
		log.Fatal("generated completion is empty")
	}
	if !strings.Contains(strings.ToLower(res), "llama") {
		log.Fatalf("generated completion returned unexpected message: %s", res)
	}
}

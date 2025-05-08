package ollamawrapper

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"testing"
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
	messages := []Message{
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
	messages := []Message{
		{
			Role:    "user",
			Content: "Say the word 'llama' to me, and nothing else.",
		},
	}
	res, err := ChatCompletionStream(client, messages, func(cr ChatResponse) error {
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
	res, err := GenerateCompletionStream(client, sysPrompt, prompt, func(gr GenerateResponse) error {
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

type genCompJsonFormat struct {
	Name       string   `json:"name"`
	Population int      `json:"population"`
	Languages  []string `json:"languages"`
}

func TestGenerateCompletionFormat(t *testing.T) {
	cmd, err := StartServer()
	if err != nil {
		log.Fatal("failed to start ollama server:", err)
	}
	defer StopServer(cmd)

	client, err := GetClient()
	if err != nil {
		log.Fatal("failed to get ollama client:", err)
	}

	// test the GenerateCompletionFormat function
	sysPrompt := "tell me information about the given country. for languages, just give me the top 3 if there are multiple."
	prompt := "United States"
	jsonFormat := json.RawMessage(`{
	"type": "object",
	"properties": {
		"name": {
			"type": "string"
		},
		"population": {
			"type": "number"
		},
		"languages": {
			"type": "array",
			"items": {
				"type": "string"
			}
		}
	},
	"required": ["name", "population", "languages"]
}`)

	res, err := GenerateCompletionOptsFormat(client, sysPrompt, prompt, nil, jsonFormat)
	if err != nil {
		log.Fatal("failed to generate completion stream:", err)
	}
	if res == "" {
		log.Fatal("generated completion is empty")
	}
	log.Println(res)
	var result genCompJsonFormat
	err = json.Unmarshal([]byte(res), &result)
	if err != nil {
		log.Fatal("failed to unmarshal result; is output malformed or incorrect?", err)
	}
	if !strings.Contains(strings.ToLower(result.Name), "united states") {
		log.Fatalf("Incorrect: expected '%s', got '%s'", "united states", result.Name)
	}
	if !(slices.Contains(result.Languages, "English") || slices.Contains(result.Languages, "english")) {
		log.Fatalf("Incorrect: expected 'english' to be in languages. got: %v", result.Languages)
	}
}

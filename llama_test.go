package ollamawrapper

import (
	"log"
	"strings"
	"testing"

	"github.com/ollama/ollama/api"
)

// run in terminal:
// go test -run ^TestLlamaWrapper$ github.com/webbben/ollama-wrapper
func TestLlamaWrapper(t *testing.T) {
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
	if len(res) == 0 {
		log.Fatal("chat completion returned no messages")
	}
	if !strings.Contains(res[0].Content, "llama") {
		log.Fatalf("chat completion returned unexpected message: %s", res[0].Content)
	}
}

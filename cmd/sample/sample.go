package main

import (
	"bufio"
	"fmt"
	"os"

	llama "github.com/webbben/ollama-wrapper"
)

func main() {
	// start server and close it upon exit
	cmd, err := llama.StartServer()
	if err != nil {
		panic(err)
	}
	defer llama.StopServer(cmd)

	// get client
	client, err := llama.GetClient()
	if err != nil {
		panic(err)
	}

	// start a chat session with streaming
	messages := make([]llama.Message, 0)
	reader := bufio.NewReader(os.Stdin)
	for {
		// get user input
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		input = input[:len(input)-1] // remove newline character

		messages = append(messages, llama.Message{
			Role:    "user",
			Content: input,
		})
		res, err := llama.ChatCompletionStream(client, messages, func(cr llama.ChatResponse) error {
			fmt.Print(cr.Message.Content)
			if cr.Done {
				fmt.Println()
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
		messages = res
	}
}

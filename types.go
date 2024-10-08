package ollamawrapper

import (
	"time"

	"github.com/ollama/ollama/api"
)

// ChatMessage is a wrapper for the api.Message struct, which is used for chat conversations.
//
// created this so we can expose the api.Message type to code that uses this package.
type Message struct {
	Role, Content string
}

func (cm *Message) ToOriginal() api.Message {
	return api.Message{
		Role:    cm.Role,
		Content: cm.Content,
	}
}

func wrapMessages(messages []api.Message) []Message {
	var wrappedMessages []Message
	for _, m := range messages {
		wrappedMessages = append(wrappedMessages, Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return wrappedMessages
}

func convertWrapperMessages(messages []Message) []api.Message {
	var convertedMessages []api.Message
	for _, m := range messages {
		convertedMessages = append(convertedMessages, m.ToOriginal())
	}
	return convertedMessages
}

// ChatResponse is a wrapper for the api.ChatResponse struct, which is used for chat conversations.
type ChatResponse struct {
	Model      string
	CreatedAt  time.Time
	Message    Message
	DoneReason string
	Done       bool
}

func wrapChatResponse(cr api.ChatResponse) ChatResponse {
	return ChatResponse{
		Model:      cr.Model,
		CreatedAt:  cr.CreatedAt,
		Message:    Message{Role: cr.Message.Role, Content: cr.Message.Content},
		DoneReason: cr.DoneReason,
		Done:       cr.Done,
	}
}

// GenerateResponse is a wrapper for the api.GenerateResponse struct, which is used for generating completions.
type GenerateResponse struct {
	// Model is the model name that generated the response.
	Model string

	// CreatedAt is the timestamp of the response.
	CreatedAt time.Time

	// Response is the textual response itself.
	Response string

	// Done specifies if the response is complete.
	Done bool

	// DoneReason is the reason the model stopped generating text.
	DoneReason string

	// Context is an encoding of the conversation used in this response; this
	// can be sent in the next request to keep a conversational memory.
	Context []int
}

func wrapGenerateResponse(gr api.GenerateResponse) GenerateResponse {
	return GenerateResponse{
		Model:      gr.Model,
		CreatedAt:  gr.CreatedAt,
		Response:   gr.Response,
		Done:       gr.Done,
		DoneReason: gr.DoneReason,
		Context:    gr.Context,
	}
}

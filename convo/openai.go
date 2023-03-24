package convo

import (
	"context"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

func askAI(client *openai.Client, model string, msgs []openai.ChatCompletionMessage) (string, openai.Usage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := client.CreateChatCompletion(
		ctx, openai.ChatCompletionRequest{
			Model:    model,
			Messages: msgs,
		},
	)
	if err != nil {
		return "", openai.Usage{}, err
	}
	return resp.Choices[0].Message.Content, resp.Usage, nil
}

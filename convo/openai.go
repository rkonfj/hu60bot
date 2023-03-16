package convo

import (
	"context"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

func askAI(client *openai.Client, model string, msgs []openai.ChatCompletionMessage) (string, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: msgs,
		},
	)

	if err != nil {
		if e, ok := err.(*openai.RequestError); ok && e.StatusCode == 429 {
			resp, err = client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:    openai.GPT3Ada,
					Messages: msgs,
				},
			)
			if err != nil {
				time.Sleep(2 * time.Second)
				resp, err = client.CreateChatCompletion(
					context.Background(),
					openai.ChatCompletionRequest{
						Model:    model,
						Messages: msgs,
					},
				)
			}
		}
	}
	if err != nil {
		return "", err
	}
	logrus.Debugf("openai usage: %d/%d/%d", resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	return resp.Choices[0].Message.Content, nil
}

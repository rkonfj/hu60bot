package convo

import (
	"context"

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
		return "", err
	}
	logrus.Debugf("openai usage: %d/%d/%d", resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	return resp.Choices[0].Message.Content, nil
}

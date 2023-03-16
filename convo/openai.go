package convo

import (
	"context"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

func askAI(client *openai.Client, model string, msgs []openai.ChatCompletionMessage) (string, openai.Usage, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: msgs,
		},
	)

	if err != nil {
		logrus.Warn("first request openai error: ", err.Error())
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
		return "", openai.Usage{}, err
	}
	return resp.Choices[0].Message.Content, resp.Usage, nil
}

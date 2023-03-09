package hu60

import (
	"context"
	"testing"
)

func TestReply(t *testing.T) {
	client := NewClient("https://hu60.cn")

	resp, err := client.Login(context.Background(), LoginRequest{Username: "hu60bot", Password: "<pass>"})

	if err != nil {
		t.Error(err)
	}

	if resp.Sid == "" {
		t.Error("sid is nil")
	}

	topicid := 104622
	topicResp, err := client.GetTopic(context.Background(), topicid, resp.Sid)

	if err != nil {
		t.Error(err)
	}

	_, err = client.ReplyTopic(context.Background(), resp.Sid, ReplyTopicRequest{Token: topicResp.Token, Content: "@#22780，via hu60bot unit test", TopicID: topicid})
	if err != nil {
		t.Error(err)
	}
}

package llm

import (
	"context"
	"fmt"
	"io"

	"gpilot/internal/infra/config"
	"gpilot/internal/infra/logger"

	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
	cfg    config.LLMConfig
}

func NewClient(cfg config.LLMConfig) *Client {
	clientCfg := openai.DefaultConfig(cfg.APIKey)
	clientCfg.BaseURL = cfg.BaseURL

	return &Client{
		client: openai.NewClientWithConfig(clientCfg),
		model:  cfg.Model,
		cfg:    cfg,
	}
}

// Chat sends a non-streaming request and returns the full response
func (c *Client) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, int, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		MaxTokens:   c.cfg.MaxTokens,
		Temperature: c.cfg.Temperature,
	})
	if err != nil {
		return "", 0, fmt.Errorf("llm chat: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("llm: empty response")
	}

	tokens := resp.Usage.TotalTokens
	return resp.Choices[0].Message.Content, tokens, nil
}

// ChatStream sends a streaming request and calls streamFn for each chunk
func (c *Client) ChatStream(ctx context.Context, systemPrompt, userPrompt string, streamFn func(chunk string)) (string, int, error) {
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		MaxTokens:   c.cfg.MaxTokens,
		Temperature: c.cfg.Temperature,
		Stream:      true,
	})
	if err != nil {
		return "", 0, fmt.Errorf("llm stream: %w", err)
	}
	defer stream.Close()

	var fullResponse string
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.L.Errorw("llm stream recv error", "error", err)
			break
		}

		if len(resp.Choices) > 0 {
			chunk := resp.Choices[0].Delta.Content
			fullResponse += chunk
			if streamFn != nil {
				streamFn(chunk)
			}
		}
	}

	// Estimate tokens (streaming doesn't return usage)
	estimatedTokens := len(fullResponse) / 4
	return fullResponse, estimatedTokens, nil
}

// ModelName returns the configured model name
func (c *Client) ModelName() string {
	return c.model
}

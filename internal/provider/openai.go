package provider

import (
    "context"
    "fmt"
    "io"

    openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface using OpenAI's Chat Completions API
type OpenAIProvider struct {
    client    *openai.Client
    modelName string
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(apiKey, modelName string) (*OpenAIProvider, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("openai apiKey is required")
    }
    if modelName == "" {
        // choose a modern default model
        modelName = "gpt-4o-mini"
    }

    client := openai.NewClient(apiKey)
    return &OpenAIProvider{
        client:    client,
        modelName: modelName,
    }, nil
}

// Generate generates a completion for the given prompt
func (o *OpenAIProvider) Generate(ctx context.Context, prompt string) (string, error) {
    req := openai.ChatCompletionRequest{
        Model: o.modelName,
        Messages: []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleUser, Content: prompt},
        },
    }

    resp, err := o.client.CreateChatCompletion(ctx, req)
    if err != nil {
        return "", fmt.Errorf("openai completion error: %w", err)
    }
    if len(resp.Choices) == 0 {
        return "", fmt.Errorf("openai returned no choices")
    }
    return resp.Choices[0].Message.Content, nil
}

// StreamGenerate generates a streaming completion for the given prompt
func (o *OpenAIProvider) StreamGenerate(ctx context.Context, prompt string) (io.ReadCloser, error) {
    req := openai.ChatCompletionRequest{
        Model: o.modelName,
        Messages: []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleUser, Content: prompt},
        },
        Stream: true,
    }

    stream, err := o.client.CreateChatCompletionStream(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("openai start stream error: %w", err)
    }

    pr, pw := io.Pipe()

    go func() {
        defer stream.Close()
        defer pw.Close()
        for {
            resp, err := stream.Recv()
            if err == io.EOF {
                return
            }
            if err != nil {
                _ = pw.CloseWithError(fmt.Errorf("openai stream recv error: %w", err))
                return
            }
            for _, choice := range resp.Choices {
                if delta := choice.Delta.Content; delta != "" {
                    if _, werr := pw.Write([]byte(delta)); werr != nil {
                        _ = pw.CloseWithError(werr)
                        return
                    }
                }
            }
        }
    }()

    return pr, nil
}

// Close closes any underlying resources (no-op for the OpenAI client)
func (o *OpenAIProvider) Close() error { return nil }

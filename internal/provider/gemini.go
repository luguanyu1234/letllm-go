package provider

import (
	"context"
	"fmt"
	"io"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Generate generates a completion for the given prompt
func (g *GeminiProvider) Generate(ctx context.Context, prompt string) (string, error) {
	model := g.client.GenerativeModel(g.modelName)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	result := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	return result, nil
}

// StreamGenerate generates a streaming completion for the given prompt
func (g *GeminiProvider) StreamGenerate(ctx context.Context, prompt string) (io.ReadCloser, error) {
	model := g.client.GenerativeModel(g.modelName)

	iter := model.GenerateContentStream(ctx, genai.Text(prompt))

	// Create a pipe for streaming the response
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		for {
			resp, err := iter.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				pw.CloseWithError(fmt.Errorf("stream error: %w", err))
				return
			}

			for _, cand := range resp.Candidates {
				for _, part := range cand.Content.Parts {
					if text, ok := part.(genai.Text); ok {
						if _, err := pw.Write([]byte(text)); err != nil {
							pw.CloseWithError(err)
							return
						}
					}
				}
			}
		}
	}()

	return pr, nil
}

// Close closes the Gemini client
func (g *GeminiProvider) Close() error {
	return g.client.Close()
}

type GeminiProvider struct {
	client    *genai.Client
	modelName string
}

// NewGeminiProvider creates a new Gemini provider instance
func NewGeminiProvider(apiKey, modelName string) (*GeminiProvider, error) {
	if modelName == "" {
		modelName = "gemini-pro"
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client:    client,
		modelName: modelName,
	}, nil
}

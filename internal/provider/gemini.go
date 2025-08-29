package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface using Google's Gemini API
type GeminiProvider struct {
	client       *genai.Client
	modelName    string
	capabilities ProviderCapabilities
}

// NewGeminiProvider creates a new Gemini provider instance
func NewGeminiProvider(apiKey, baseURL, modelName string) (*GeminiProvider, error) {
	if modelName == "" {
		modelName = "gemini-pro"
	}

	ctx := context.Background()

	// Create client with optional custom base URL
	var client *genai.Client
	var err error
	if baseURL != "" {
		// For Gemini, we need to set the endpoint through the option
		client, err = genai.NewClient(ctx, option.WithAPIKey(apiKey), option.WithEndpoint(baseURL))
	} else {
		client, err = genai.NewClient(ctx, option.WithAPIKey(apiKey))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Define Gemini capabilities
	capabilities := ProviderCapabilities{
		SupportsStreaming:   true,
		SupportsFunctions:   true,
		SupportsSystemRole:  false, // Gemini doesn't have explicit system role
		MaxTokens:           2048,
		MaxContextLength:    32768, // For Gemini Pro
		SupportedModels:     []string{"gemini-pro", "gemini-pro-vision", "gemini-1.5-pro", "gemini-1.5-flash"},
		SupportedParameters: []string{"temperature", "top_p", "max_tokens", "stream"},
	}

	return &GeminiProvider{
		client:       client,
		modelName:    modelName,
		capabilities: capabilities,
	}, nil
}

// Generate generates a completion for the given request
func (g *GeminiProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if err := ValidateStandardRequest(req.StandardRequest); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	model := g.client.GenerativeModel(req.Model)

	// Configure model parameters
	if req.Temperature != nil {
		temp := float32(*req.Temperature)
		model.Temperature = &temp
	}

	if req.TopP != nil {
		topP := float32(*req.TopP)
		model.TopP = &topP
	}

	if req.MaxTokens != nil {
		maxTokens := int32(*req.MaxTokens)
		model.MaxOutputTokens = &maxTokens
	}

	// Convert messages to Gemini format
	parts, err := g.convertMessagesToParts(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	standardResp, err := g.transformResponse(resp, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to transform response: %w", err)
	}

	return &GenerateResponse{
		StandardResponse: standardResp,
	}, nil
}

// StreamGenerate generates a streaming completion for the given request
func (g *GeminiProvider) StreamGenerate(ctx context.Context, req *GenerateRequest) (io.ReadCloser, error) {
	if err := ValidateStandardRequest(req.StandardRequest); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	model := g.client.GenerativeModel(req.Model)

	// Configure model parameters
	if req.Temperature != nil {
		temp := float32(*req.Temperature)
		model.Temperature = &temp
	}

	if req.TopP != nil {
		topP := float32(*req.TopP)
		model.TopP = &topP
	}

	if req.MaxTokens != nil {
		maxTokens := int32(*req.MaxTokens)
		model.MaxOutputTokens = &maxTokens
	}

	// Convert messages to Gemini format
	parts, err := g.convertMessagesToParts(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	iter := model.GenerateContentStream(ctx, parts...)

	// Create a pipe for streaming the response
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		chunkID := fmt.Sprintf("chatcmpl-%d", time.Now().Unix())
		chunkIndex := 0

		for {
			resp, err := iter.Next()
			if err == io.EOF {
				// Send final chunk
				finalChunk := CreateStreamChunk(chunkID, req.Model, []Choice{}, true)
				if chunkData, marshalErr := json.Marshal(finalChunk); marshalErr == nil {
					pw.Write(append(chunkData, '\n'))
				}
				break
			}
			if err != nil {
				pw.CloseWithError(fmt.Errorf("stream error: %w", err))
				return
			}

			// Transform streaming response to standard format
			chunk, err := g.transformStreamChunk(resp, chunkID, req.Model, chunkIndex)
			if err != nil {
				pw.CloseWithError(fmt.Errorf("failed to transform stream chunk: %w", err))
				return
			}

			// Write chunk as JSON
			chunkData, err := json.Marshal(chunk)
			if err != nil {
				pw.CloseWithError(fmt.Errorf("failed to marshal chunk: %w", err))
				return
			}

			if _, werr := pw.Write(append(chunkData, '\n')); werr != nil {
				pw.CloseWithError(werr)
				return
			}

			chunkIndex++
		}
	}()

	return pr, nil
}

// GetCapabilities returns the capabilities of the Gemini provider
func (g *GeminiProvider) GetCapabilities() ProviderCapabilities {
	return g.capabilities
}

// GetInfo returns information about the Gemini provider
func (g *GeminiProvider) GetInfo() ProviderInfo {
	return ProviderInfo{
		Name:         "gemini",
		Version:      "1.0.0",
		Capabilities: g.capabilities,
		Status:       "active",
		LastUpdated:  time.Now(),
	}
}

// Close closes the Gemini client
func (g *GeminiProvider) Close() error {
	return g.client.Close()
}

// convertMessagesToParts converts standard messages to Gemini parts
func (g *GeminiProvider) convertMessagesToParts(messages []Message) ([]genai.Part, error) {
	var parts []genai.Part

	// Gemini doesn't support system messages directly, so we'll prepend system messages to the first user message
	var systemContent strings.Builder
	var conversationParts []genai.Part

	for _, msg := range messages {
		switch msg.Role {
		case RoleSystem:
			if systemContent.Len() > 0 {
				systemContent.WriteString("\n")
			}
			systemContent.WriteString(msg.Content)
		case RoleUser:
			content := msg.Content
			if systemContent.Len() > 0 {
				content = systemContent.String() + "\n\n" + content
				systemContent.Reset() // Only prepend to first user message
			}
			conversationParts = append(conversationParts, genai.Text(content))
		case RoleAssistant:
			conversationParts = append(conversationParts, genai.Text(msg.Content))
		default:
			return nil, fmt.Errorf("unsupported message role: %s", msg.Role)
		}
	}

	parts = append(parts, conversationParts...)
	return parts, nil
}

// transformResponse converts a Gemini response to StandardResponse
func (g *GeminiProvider) transformResponse(resp *genai.GenerateContentResponse, model string) (*StandardResponse, error) {
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	choices := make([]Choice, len(resp.Candidates))

	for i, candidate := range resp.Candidates {
		var content strings.Builder

		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					content.WriteString(string(text))
				}
			}
		}

		msg := &Message{
			Role:    RoleAssistant,
			Content: content.String(),
		}

		var finishReason *string
		if candidate.FinishReason != 0 {
			reason := g.mapFinishReason(candidate.FinishReason)
			finishReason = &reason
		}

		choices[i] = Choice{
			Index:        i,
			Message:      msg,
			FinishReason: finishReason,
		}
	}

	// Gemini doesn't provide detailed usage info in the same way
	usage := Usage{
		PromptTokens:     0, // Not available from Gemini
		CompletionTokens: 0, // Not available from Gemini
		TotalTokens:      0, // Not available from Gemini
	}

	responseID := fmt.Sprintf("chatcmpl-%d", time.Now().Unix())
	return CreateStandardResponse(responseID, model, choices, usage), nil
}

// transformStreamChunk converts a Gemini stream response to StreamChunk
func (g *GeminiProvider) transformStreamChunk(resp *genai.GenerateContentResponse, chunkID, model string, index int) (*StreamChunk, error) {
	choices := make([]Choice, 0, len(resp.Candidates))

	for i, candidate := range resp.Candidates {
		var content strings.Builder

		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					content.WriteString(string(text))
				}
			}
		}

		delta := &Message{
			Role:    RoleAssistant,
			Content: content.String(),
		}

		var finishReason *string
		if candidate.FinishReason != 0 {
			reason := g.mapFinishReason(candidate.FinishReason)
			finishReason = &reason
		}

		choices = append(choices, Choice{
			Index:        i,
			Delta:        delta,
			FinishReason: finishReason,
		})
	}

	done := len(resp.Candidates) > 0 && resp.Candidates[0].FinishReason != 0

	return CreateStreamChunk(chunkID, model, choices, done), nil
}

// mapFinishReason maps Gemini finish reasons to standard format
func (g *GeminiProvider) mapFinishReason(reason genai.FinishReason) string {
	switch reason {
	case genai.FinishReasonStop:
		return FinishReasonStop
	case genai.FinishReasonMaxTokens:
		return FinishReasonLength
	case genai.FinishReasonSafety:
		return FinishReasonContentFilter
	default:
		return "unknown"
	}
}

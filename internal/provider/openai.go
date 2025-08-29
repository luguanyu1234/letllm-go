package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface using OpenAI's Chat Completions API
type OpenAIProvider struct {
	client       *openai.Client
	modelName    string
	capabilities ProviderCapabilities
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(apiKey, baseURL, modelName string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai apiKey is required")
	}
	if modelName == "" {
		// choose a modern default model
		modelName = "gpt-4o-mini"
	}

	// Create client with optional custom base URL
	var client *openai.Client
	if baseURL != "" {
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = baseURL
		client = openai.NewClientWithConfig(config)
	} else {
		client = openai.NewClient(apiKey)
	}

	// Define OpenAI capabilities
	capabilities := ProviderCapabilities{
		SupportsStreaming:   true,
		SupportsFunctions:   true,
		SupportsSystemRole:  true,
		MaxTokens:           4096,
		MaxContextLength:    128000, // For GPT-4 models
		SupportedModels:     []string{"gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"},
		SupportedParameters: []string{"temperature", "top_p", "max_tokens", "stream", "functions"},
	}

	return &OpenAIProvider{
		client:       client,
		modelName:    modelName,
		capabilities: capabilities,
	}, nil
}

// Generate generates a completion for the given request
func (o *OpenAIProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if err := ValidateStandardRequest(req.StandardRequest); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	openaiReq, err := o.transformRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to transform request: %w", err)
	}

	resp, err := o.client.CreateChatCompletion(ctx, *openaiReq)
	if err != nil {
		return nil, fmt.Errorf("openai completion error: %w", err)
	}

	standardResp, err := o.transformResponse(&resp)
	if err != nil {
		return nil, fmt.Errorf("failed to transform response: %w", err)
	}

	return &GenerateResponse{
		StandardResponse: standardResp,
	}, nil
}

// StreamGenerate generates a streaming completion for the given request
func (o *OpenAIProvider) StreamGenerate(ctx context.Context, req *GenerateRequest) (io.ReadCloser, error) {
	if err := ValidateStandardRequest(req.StandardRequest); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	openaiReq, err := o.transformRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to transform request: %w", err)
	}

	openaiReq.Stream = true

	stream, err := o.client.CreateChatCompletionStream(ctx, *openaiReq)
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

			// Transform streaming response to standard format
			chunk, err := o.transformStreamChunk(&resp)
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("failed to transform stream chunk: %w", err))
				return
			}

			// Write chunk as JSON
			chunkData, err := json.Marshal(chunk)
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("failed to marshal chunk: %w", err))
				return
			}

			if _, werr := pw.Write(append(chunkData, '\n')); werr != nil {
				_ = pw.CloseWithError(werr)
				return
			}
		}
	}()

	return pr, nil
}

// GetCapabilities returns the capabilities of the OpenAI provider
func (o *OpenAIProvider) GetCapabilities() ProviderCapabilities {
	return o.capabilities
}

// GetInfo returns information about the OpenAI provider
func (o *OpenAIProvider) GetInfo() ProviderInfo {
	return ProviderInfo{
		Name:         "openai",
		Version:      "1.0.0",
		Capabilities: o.capabilities,
		Status:       "active",
		LastUpdated:  time.Now(),
	}
}

// Close closes any underlying resources (no-op for the OpenAI client)
func (o *OpenAIProvider) Close() error {
	return nil
}

// transformRequest converts a StandardRequest to OpenAI format
func (o *OpenAIProvider) transformRequest(req *GenerateRequest) (*openai.ChatCompletionRequest, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))

	for i, msg := range req.Messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		if msg.Name != nil {
			openaiMsg.Name = *msg.Name
		}

		if msg.FunctionCall != nil {
			openaiMsg.FunctionCall = &openai.FunctionCall{
				Name:      msg.FunctionCall.Name,
				Arguments: msg.FunctionCall.Arguments,
			}
		}

		messages[i] = openaiMsg
	}

	openaiReq := &openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   req.Stream,
	}

	if req.MaxTokens != nil {
		openaiReq.MaxTokens = *req.MaxTokens
	}

	if req.Temperature != nil {
		openaiReq.Temperature = float32(*req.Temperature)
	}

	if req.TopP != nil {
		openaiReq.TopP = float32(*req.TopP)
	}

	if len(req.Functions) > 0 {
		functions := make([]openai.FunctionDefinition, len(req.Functions))
		for i, fn := range req.Functions {
			functions[i] = openai.FunctionDefinition{
				Name:        fn.Name,
				Description: fn.Description,
				Parameters:  fn.Parameters,
			}
		}
		openaiReq.Functions = functions
	}

	return openaiReq, nil
}

// transformResponse converts an OpenAI response to StandardResponse
func (o *OpenAIProvider) transformResponse(resp *openai.ChatCompletionResponse) (*StandardResponse, error) {
	choices := make([]Choice, len(resp.Choices))

	for i, choice := range resp.Choices {
		msg := &Message{
			Role:    choice.Message.Role,
			Content: choice.Message.Content,
		}

		if choice.Message.Name != "" {
			msg.Name = &choice.Message.Name
		}

		if choice.Message.FunctionCall != nil {
			msg.FunctionCall = &FunctionCall{
				Name:      choice.Message.FunctionCall.Name,
				Arguments: choice.Message.FunctionCall.Arguments,
			}
		}

		var finishReason *string
		if choice.FinishReason != "" {
			reason := string(choice.FinishReason)
			finishReason = &reason
		}

		choices[i] = Choice{
			Index:        choice.Index,
			Message:      msg,
			FinishReason: finishReason,
		}
	}

	usage := Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	return CreateStandardResponse(resp.ID, resp.Model, choices, usage), nil
}

// transformStreamChunk converts an OpenAI stream response to StreamChunk
func (o *OpenAIProvider) transformStreamChunk(resp *openai.ChatCompletionStreamResponse) (*StreamChunk, error) {
	choices := make([]Choice, len(resp.Choices))

	for i, choice := range resp.Choices {
		delta := &Message{
			Role:    choice.Delta.Role,
			Content: choice.Delta.Content,
		}

		if choice.Delta.FunctionCall != nil {
			delta.FunctionCall = &FunctionCall{
				Name:      choice.Delta.FunctionCall.Name,
				Arguments: choice.Delta.FunctionCall.Arguments,
			}
		}

		var finishReason *string
		if choice.FinishReason != "" {
			reason := string(choice.FinishReason)
			finishReason = &reason
		}

		choices[i] = Choice{
			Index:        choice.Index,
			Delta:        delta,
			FinishReason: finishReason,
		}
	}

	done := len(resp.Choices) > 0 && resp.Choices[0].FinishReason != ""

	return CreateStreamChunk(resp.ID, resp.Model, choices, done), nil
}

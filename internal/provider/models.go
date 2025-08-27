package provider

import (
	"time"
)

// StandardRequest represents a standardized request format that can be transformed to/from provider-specific formats
type StandardRequest struct {
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	Stream      bool                   `json:"stream,omitempty"`
	MaxTokens   *int                   `json:"max_tokens,omitempty"`
	Temperature *float64               `json:"temperature,omitempty"`
	TopP        *float64               `json:"top_p,omitempty"`
	Functions   []Function             `json:"functions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StandardResponse represents a standardized response format
type StandardResponse struct {
	ID       string                 `json:"id"`
	Object   string                 `json:"object"`
	Created  int64                  `json:"created"`
	Model    string                 `json:"model"`
	Choices  []Choice               `json:"choices"`
	Usage    Usage                  `json:"usage"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role         string                 `json:"role"`
	Content      string                 `json:"content"`
	Name         *string                `json:"name,omitempty"`
	FunctionCall *FunctionCall          `json:"function_call,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason *string  `json:"finish_reason,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Function represents a function definition for function calling
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// FunctionCall represents a function call in a message
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ProviderCapabilities represents the capabilities of a provider
type ProviderCapabilities struct {
	SupportsStreaming   bool     `json:"supports_streaming"`
	SupportsFunctions   bool     `json:"supports_functions"`
	SupportsSystemRole  bool     `json:"supports_system_role"`
	MaxTokens           int      `json:"max_tokens"`
	MaxContextLength    int      `json:"max_context_length"`
	SupportedModels     []string `json:"supported_models"`
	SupportedParameters []string `json:"supported_parameters"`
}

// ProviderInfo represents information about a provider
type ProviderInfo struct {
	Name         string               `json:"name"`
	Version      string               `json:"version"`
	Capabilities ProviderCapabilities `json:"capabilities"`
	Status       string               `json:"status"`
	LastUpdated  time.Time            `json:"last_updated"`
}

// GenerateRequest represents a request for text generation
type GenerateRequest struct {
	*StandardRequest
	ProviderSpecific map[string]interface{} `json:"provider_specific,omitempty"`
}

// GenerateResponse represents a response from text generation
type GenerateResponse struct {
	*StandardResponse
	ProviderSpecific map[string]interface{} `json:"provider_specific,omitempty"`
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []Choice     `json:"choices"`
	Done    bool         `json:"done"`
	Usage   *Usage       `json:"usage,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail represents detailed error information
type ErrorDetail struct {
	Type    string                 `json:"type"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Common role constants
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleFunction  = "function"
)

// Common finish reason constants
const (
	FinishReasonStop          = "stop"
	FinishReasonLength        = "length"
	FinishReasonFunctionCall  = "function_call"
	FinishReasonContentFilter = "content_filter"
)

// Common object type constants
const (
	ObjectChatCompletion      = "chat.completion"
	ObjectChatCompletionChunk = "chat.completion.chunk"
)

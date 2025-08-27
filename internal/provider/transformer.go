package provider

import (
	"fmt"
	"time"
)

// RequestTransformer defines the interface for transforming requests between formats
type RequestTransformer interface {
	ToStandard(providerRequest interface{}) (*StandardRequest, error)
	FromStandard(standardRequest *StandardRequest) (interface{}, error)
}

// ResponseTransformer defines the interface for transforming responses between formats
type ResponseTransformer interface {
	ToStandard(providerResponse interface{}) (*StandardResponse, error)
	FromStandard(standardResponse *StandardResponse) (interface{}, error)
}

// TransformationError represents an error during transformation
type TransformationError struct {
	Operation string
	Provider  string
	Reason    string
	Err       error
}

func (e *TransformationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("transformation error in %s for %s: %s - %v", e.Operation, e.Provider, e.Reason, e.Err)
	}
	return fmt.Sprintf("transformation error in %s for %s: %s", e.Operation, e.Provider, e.Reason)
}

// NewTransformationError creates a new transformation error
func NewTransformationError(operation, provider, reason string, err error) *TransformationError {
	return &TransformationError{
		Operation: operation,
		Provider:  provider,
		Reason:    reason,
		Err:       err,
	}
}

// ValidateStandardRequest validates a standard request
func ValidateStandardRequest(req *StandardRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	if len(req.Messages) == 0 {
		return fmt.Errorf("at least one message is required")
	}

	// Validate messages
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d: role is required", i)
		}

		if msg.Content == "" && msg.FunctionCall == nil {
			return fmt.Errorf("message %d: content or function_call is required", i)
		}

		// Validate role values
		switch msg.Role {
		case RoleSystem, RoleUser, RoleAssistant, RoleFunction:
			// Valid roles
		default:
			return fmt.Errorf("message %d: invalid role '%s'", i, msg.Role)
		}
	}

	// Validate parameters
	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return fmt.Errorf("top_p must be between 0 and 1")
	}

	if req.MaxTokens != nil && *req.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	return nil
}

// CreateStandardResponse creates a standard response with common fields populated
func CreateStandardResponse(id, model string, choices []Choice, usage Usage) *StandardResponse {
	return &StandardResponse{
		ID:      id,
		Object:  ObjectChatCompletion,
		Created: time.Now().Unix(),
		Model:   model,
		Choices: choices,
		Usage:   usage,
	}
}

// CreateStreamChunk creates a stream chunk with common fields populated
func CreateStreamChunk(id, model string, choices []Choice, done bool) *StreamChunk {
	chunk := &StreamChunk{
		ID:      id,
		Object:  ObjectChatCompletionChunk,
		Created: time.Now().Unix(),
		Model:   model,
		Choices: choices,
		Done:    done,
	}

	return chunk
}

// MergeCapabilities merges multiple provider capabilities
func MergeCapabilities(capabilities ...ProviderCapabilities) ProviderCapabilities {
	if len(capabilities) == 0 {
		return ProviderCapabilities{}
	}

	merged := capabilities[0]

	for i := 1; i < len(capabilities); i++ {
		cap := capabilities[i]

		// Use logical OR for boolean capabilities
		merged.SupportsStreaming = merged.SupportsStreaming || cap.SupportsStreaming
		merged.SupportsFunctions = merged.SupportsFunctions || cap.SupportsFunctions
		merged.SupportsSystemRole = merged.SupportsSystemRole || cap.SupportsSystemRole

		// Use maximum for numeric capabilities
		if cap.MaxTokens > merged.MaxTokens {
			merged.MaxTokens = cap.MaxTokens
		}
		if cap.MaxContextLength > merged.MaxContextLength {
			merged.MaxContextLength = cap.MaxContextLength
		}

		// Merge string slices
		merged.SupportedModels = mergeStringSlices(merged.SupportedModels, cap.SupportedModels)
		merged.SupportedParameters = mergeStringSlices(merged.SupportedParameters, cap.SupportedParameters)
	}

	return merged
}

// mergeStringSlices merges two string slices, removing duplicates
func mergeStringSlices(slice1, slice2 []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice1)+len(slice2))

	for _, s := range slice1 {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	for _, s := range slice2 {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

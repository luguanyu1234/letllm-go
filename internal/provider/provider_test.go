package provider

import (
	"context"
	"testing"
)

func TestProviderCapabilities(t *testing.T) {
	// Test OpenAI provider capabilities
	openaiProvider, err := NewOpenAIProvider("test-key", "", "gpt-4")
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	openaiCaps := openaiProvider.GetCapabilities()
	if !openaiCaps.SupportsStreaming {
		t.Error("OpenAI should support streaming")
	}

	if !openaiCaps.SupportsFunctions {
		t.Error("OpenAI should support functions")
	}

	if !openaiCaps.SupportsSystemRole {
		t.Error("OpenAI should support system role")
	}

	if len(openaiCaps.SupportedModels) == 0 {
		t.Error("OpenAI should have supported models")
	}

	// Test Gemini provider capabilities
	geminiProvider, err := NewGeminiProvider("test-key", "", "gemini-pro")
	if err != nil {
		t.Fatalf("Failed to create Gemini provider: %v", err)
	}

	geminiCaps := geminiProvider.GetCapabilities()
	if !geminiCaps.SupportsStreaming {
		t.Error("Gemini should support streaming")
	}

	if geminiCaps.SupportsSystemRole {
		t.Error("Gemini should not support system role")
	}

	if len(geminiCaps.SupportedModels) == 0 {
		t.Error("Gemini should have supported models")
	}
}

func TestProviderInfo(t *testing.T) {
	// Test OpenAI provider info
	openaiProvider, err := NewOpenAIProvider("test-key", "", "gpt-4")
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	openaiInfo := openaiProvider.GetInfo()
	if openaiInfo.Name != "openai" {
		t.Errorf("Expected name 'openai', got '%s'", openaiInfo.Name)
	}

	if openaiInfo.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", openaiInfo.Status)
	}

	// Test Gemini provider info
	geminiProvider, err := NewGeminiProvider("test-key", "", "gemini-pro")
	if err != nil {
		t.Fatalf("Failed to create Gemini provider: %v", err)
	}

	geminiInfo := geminiProvider.GetInfo()
	if geminiInfo.Name != "gemini" {
		t.Errorf("Expected name 'gemini', got '%s'", geminiInfo.Name)
	}

	if geminiInfo.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", geminiInfo.Status)
	}
}

func TestRequestValidation(t *testing.T) {
	openaiProvider, err := NewOpenAIProvider("test-key", "", "gpt-4")
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	ctx := context.Background()

	// Test with invalid request (should fail validation)
	invalidReq := &GenerateRequest{
		StandardRequest: &StandardRequest{
			// Missing model and messages
		},
	}

	_, err = openaiProvider.Generate(ctx, invalidReq)
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
	}

	// Test with valid request structure (will fail at API call, but should pass validation)
	validReq := &GenerateRequest{
		StandardRequest: &StandardRequest{
			Model: "gpt-4",
			Messages: []Message{
				{Role: RoleUser, Content: "Hello"},
			},
		},
	}

	// This will fail at the API call level since we're using a test key,
	// but it should pass the validation step
	_, err = openaiProvider.Generate(ctx, validReq)
	if err != nil && err.Error() == "invalid request: model is required" {
		t.Error("Valid request failed validation")
	}
	// We expect an API error here, not a validation error
}

func TestTransformationErrors(t *testing.T) {
	err := NewTransformationError("request", "openai", "invalid format", nil)

	expectedMsg := "transformation error in request for openai: invalid format"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test with wrapped error
	wrappedErr := NewTransformationError("response", "gemini", "parsing failed",
		&TransformationError{Operation: "inner", Provider: "test", Reason: "test"})

	if wrappedErr.Err == nil {
		t.Error("Expected wrapped error to be preserved")
	}
}

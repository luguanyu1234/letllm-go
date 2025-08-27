package provider

import (
	"testing"
)

func TestValidateStandardRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *StandardRequest
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty model",
			req: &StandardRequest{
				Messages: []Message{{Role: RoleUser, Content: "test"}},
			},
			wantErr: true,
		},
		{
			name: "empty messages",
			req: &StandardRequest{
				Model:    "gpt-4",
				Messages: []Message{},
			},
			wantErr: true,
		},
		{
			name: "valid request",
			req: &StandardRequest{
				Model:    "gpt-4",
				Messages: []Message{{Role: RoleUser, Content: "test"}},
			},
			wantErr: false,
		},
		{
			name: "invalid temperature",
			req: &StandardRequest{
				Model:       "gpt-4",
				Messages:    []Message{{Role: RoleUser, Content: "test"}},
				Temperature: floatPtr(3.0),
			},
			wantErr: true,
		},
		{
			name: "invalid top_p",
			req: &StandardRequest{
				Model:    "gpt-4",
				Messages: []Message{{Role: RoleUser, Content: "test"}},
				TopP:     floatPtr(1.5),
			},
			wantErr: true,
		},
		{
			name: "invalid max_tokens",
			req: &StandardRequest{
				Model:     "gpt-4",
				Messages:  []Message{{Role: RoleUser, Content: "test"}},
				MaxTokens: intPtr(-1),
			},
			wantErr: true,
		},
		{
			name: "message without role",
			req: &StandardRequest{
				Model:    "gpt-4",
				Messages: []Message{{Content: "test"}},
			},
			wantErr: true,
		},
		{
			name: "message without content or function call",
			req: &StandardRequest{
				Model:    "gpt-4",
				Messages: []Message{{Role: RoleUser}},
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			req: &StandardRequest{
				Model:    "gpt-4",
				Messages: []Message{{Role: "invalid", Content: "test"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStandardRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStandardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateStandardResponse(t *testing.T) {
	choices := []Choice{
		{
			Index: 0,
			Message: &Message{
				Role:    RoleAssistant,
				Content: "Hello!",
			},
			FinishReason: stringPtr(FinishReasonStop),
		},
	}

	usage := Usage{
		PromptTokens:     10,
		CompletionTokens: 5,
		TotalTokens:      15,
	}

	resp := CreateStandardResponse("test-id", "gpt-4", choices, usage)

	if resp.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", resp.ID)
	}

	if resp.Model != "gpt-4" {
		t.Errorf("Expected Model 'gpt-4', got '%s'", resp.Model)
	}

	if resp.Object != ObjectChatCompletion {
		t.Errorf("Expected Object '%s', got '%s'", ObjectChatCompletion, resp.Object)
	}

	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Usage.TotalTokens != 15 {
		t.Errorf("Expected TotalTokens 15, got %d", resp.Usage.TotalTokens)
	}
}

func TestCreateStreamChunk(t *testing.T) {
	choices := []Choice{
		{
			Index: 0,
			Delta: &Message{
				Role:    RoleAssistant,
				Content: "Hello",
			},
		},
	}

	chunk := CreateStreamChunk("test-id", "gpt-4", choices, false)

	if chunk.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", chunk.ID)
	}

	if chunk.Model != "gpt-4" {
		t.Errorf("Expected Model 'gpt-4', got '%s'", chunk.Model)
	}

	if chunk.Object != ObjectChatCompletionChunk {
		t.Errorf("Expected Object '%s', got '%s'", ObjectChatCompletionChunk, chunk.Object)
	}

	if chunk.Done != false {
		t.Errorf("Expected Done false, got %v", chunk.Done)
	}
}

func TestMergeCapabilities(t *testing.T) {
	cap1 := ProviderCapabilities{
		SupportsStreaming:   true,
		SupportsFunctions:   false,
		MaxTokens:           1000,
		SupportedModels:     []string{"model1", "model2"},
		SupportedParameters: []string{"param1"},
	}

	cap2 := ProviderCapabilities{
		SupportsStreaming:   false,
		SupportsFunctions:   true,
		MaxTokens:           2000,
		SupportedModels:     []string{"model2", "model3"},
		SupportedParameters: []string{"param2"},
	}

	merged := MergeCapabilities(cap1, cap2)

	if !merged.SupportsStreaming {
		t.Error("Expected SupportsStreaming to be true after merge")
	}

	if !merged.SupportsFunctions {
		t.Error("Expected SupportsFunctions to be true after merge")
	}

	if merged.MaxTokens != 2000 {
		t.Errorf("Expected MaxTokens 2000, got %d", merged.MaxTokens)
	}

	expectedModels := []string{"model1", "model2", "model3"}
	if len(merged.SupportedModels) != len(expectedModels) {
		t.Errorf("Expected %d models, got %d", len(expectedModels), len(merged.SupportedModels))
	}

	expectedParams := []string{"param1", "param2"}
	if len(merged.SupportedParameters) != len(expectedParams) {
		t.Errorf("Expected %d parameters, got %d", len(expectedParams), len(merged.SupportedParameters))
	}
}

// Helper functions for tests
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

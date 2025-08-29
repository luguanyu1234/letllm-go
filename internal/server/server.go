package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/luguanyu1234/letllm-go/internal/config"
	"github.com/luguanyu1234/letllm-go/internal/provider"
	"go.uber.org/fx"
	"github.com/gin-gonic/gin"
)

// Module provides the HTTP server lifecycle using Gin
var Module = fx.Module("http-server",
	fx.Provide(NewEngine),
	fx.Invoke(RegisterRoutes),
	fx.Invoke(StartServer),
)

// NewEngine constructs a new gin.Engine
func NewEngine() *gin.Engine {
	// Use release mode unless explicitly set otherwise by the caller
	if gin.Mode() == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	return r
}

// StartServer starts the HTTP server and registers lifecycle hooks
func StartServer(lc fx.Lifecycle, engine *gin.Engine, cfg *config.Config) {
	addr := cfg.Server.Addr
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("http server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		},
	})
}

// RegisterRoutes wires handlers on Gin
func RegisterRoutes(engine *gin.Engine, r *provider.Router) {
	engine.POST("/v1/chat/completions", func(c *gin.Context) {
		var in OpenAIChatCompletionRequest
		if err := c.ShouldBindJSON(&in); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid json: %v", err)})
			return
		}
		if in.Model == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "model is required"})
			return
		}

		p, err := r.ForModel(in.Model)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert to standard request format
		standardReq := convertToStandardRequest(&in)

		if in.Stream {
			// SSE streaming compatible with OpenAI
			c.Writer.Header().Set("Content-Type", "text/event-stream")
			c.Writer.Header().Set("Cache-Control", "no-cache")
			c.Writer.Header().Set("Connection", "keep-alive")
			flusher, ok := c.Writer.(http.Flusher)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
				return
			}

			rc, err := p.StreamGenerate(c.Request.Context(), &provider.GenerateRequest{StandardRequest: standardReq})
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rc.Close()

			// exchange channel between reader and sender
			chunks := make(chan []byte, 8)
			errCh := make(chan error, 1)

			// receiver goroutine: reads from provider stream
			go func() {
				defer close(chunks)
				buf := make([]byte, 4096)
				for {
					n, readErr := rc.Read(buf)
					if n > 0 {
						cp := make([]byte, n)
						copy(cp, buf[:n])
						chunks <- cp
					}
					if readErr == io.EOF {
						return
					}
					if readErr != nil {
						errCh <- readErr
						return
					}
				}
			}()

			enc := json.NewEncoder(c.Writer)
			for {
				select {
				case <-c.Request.Context().Done():
					return
				case err := <-errCh:
					if err != nil {
						return
					}
				case b, ok := <-chunks:
					if !ok {
						// finished
						_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
						flusher.Flush()
						return
					}
					payload := OpenAIChatCompletionChunk{
						Object: "chat.completion.chunk",
						Choices: []OpenAIChatChunkChoice{{
							Delta:        OpenAIChatMessage{Role: "assistant", Content: string(b)},
							Index:        0,
							FinishReason: nil,
						}},
						Model: in.Model,
					}
					_, _ = c.Writer.Write([]byte("data: "))
					if err := enc.Encode(payload); err != nil {
						return
					}
					_, _ = c.Writer.Write([]byte("\n"))
					flusher.Flush()
				}
			}
		}

		// Non-streaming
		resp, err := p.Generate(c.Request.Context(), &provider.GenerateRequest{StandardRequest: standardReq})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Convert back to OpenAI format
		out := convertFromStandardResponse(resp.StandardResponse)
		c.JSON(http.StatusOK, out)
	})
}

// --- Minimal OpenAI-compatible types ---
type OpenAIChatCompletionRequest struct {
	Model    string              `json:"model"`
	Messages []OpenAIChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIChatCompletionResponse struct {
	Object  string             `json:"object"`
	Model   string             `json:"model"`
	Choices []OpenAIChatChoice `json:"choices"`
}

type OpenAIChatChoice struct {
	Index        int               `json:"index"`
	Message      OpenAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type OpenAIChatCompletionChunk struct {
	Object  string                  `json:"object"`
	Model   string                  `json:"model"`
	Choices []OpenAIChatChunkChoice `json:"choices"`
}

type OpenAIChatChunkChoice struct {
	Index        int               `json:"index"`
	Delta        OpenAIChatMessage `json:"delta"`
	FinishReason *string           `json:"finish_reason"`
}

func concatUserMessages(msgs []OpenAIChatMessage) string {
	if len(msgs) == 0 {
		return ""
	}
	out := ""
	for _, m := range msgs {
		if m.Role == "user" {
			if out != "" {
				out += "\n"
			}
			out += m.Content
		}
	}
	return out
}

// convertToStandardRequest converts OpenAI request format to standard format
func convertToStandardRequest(req *OpenAIChatCompletionRequest) *provider.StandardRequest {
	messages := make([]provider.Message, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = provider.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return &provider.StandardRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   req.Stream,
	}
}

// convertFromStandardResponse converts standard response format to OpenAI format
func convertFromStandardResponse(resp *provider.StandardResponse) OpenAIChatCompletionResponse {
	choices := make([]OpenAIChatChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		finishReason := "stop"
		if choice.FinishReason != nil {
			finishReason = *choice.FinishReason
		}

		content := ""
		if choice.Message != nil {
			content = choice.Message.Content
		}

		choices[i] = OpenAIChatChoice{
			Index: choice.Index,
			Message: OpenAIChatMessage{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: finishReason,
		}
	}

	return OpenAIChatCompletionResponse{
		Object:  "chat.completion",
		Model:   resp.Model,
		Choices: choices,
	}
}

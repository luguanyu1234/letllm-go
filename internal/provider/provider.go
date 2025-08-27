package provider

import (
	"context"
	"io"
)

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
	StreamGenerate(ctx context.Context, prompt string) (io.ReadCloser, error)
}

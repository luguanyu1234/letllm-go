package main

import (
	"github.com/luguanyu1234/letllm-go/internal/config"
	"github.com/luguanyu1234/letllm-go/internal/provider"
	"github.com/luguanyu1234/letllm-go/internal/server"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		provider.Module,
		server.Module,
	).Run()
}

package main

import (
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Invoke(func() {
			println("Hello World")
		}),
	).Run()
}

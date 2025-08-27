package provider

import "go.uber.org/fx"

// Module exports the provider router for dependency injection.
var Module = fx.Provide(NewRouter)

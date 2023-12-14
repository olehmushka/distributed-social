package accounts

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewHandlers,
		NewRoutes,
	),
)

package accounts

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ipfs/go-log"
	"github.com/olehmushka/distributed-social/api/middlewares"
	"go.uber.org/fx"
)

type routesParams struct {
	fx.In

	Handlers Handlers
	Logger   *log.ZapEventLogger
}

func NewRoutes(params routesParams) *mux.Router {
	router := mux.NewRouter()
	enrichContext := middlewares.NewEnrichContextMiddleware(params.Logger)
	logRequest := middlewares.NewLoggerMiddleware(params.Logger)
	recovery := middlewares.NewRecoverMiddleware(params.Logger)

	router.Use(mux.MiddlewareFunc(enrichContext))
	router.HandleFunc("/ping", params.Handlers.Ping).Methods(http.MethodGet)
	router.HandleFunc("/info", params.Handlers.Info).Methods(http.MethodGet)
	router.Use(mux.MiddlewareFunc(logRequest))
	router.Use(mux.MiddlewareFunc(recovery))

	return router
}

package accounts

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ipfs/go-log"
	"go.uber.org/fx"

	"github.com/olehmushka/distributed-social/api/middlewares"
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
	router.HandleFunc("/users", params.Handlers.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}", params.Handlers.GetUser).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}/posts", params.Handlers.CreatePost).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}/posts", params.Handlers.ListUserPosts).Methods(http.MethodGet)
	router.HandleFunc("/posts/{id}", params.Handlers.GetPost).Methods(http.MethodGet)
	router.Use(mux.MiddlewareFunc(logRequest))
	router.Use(mux.MiddlewareFunc(recovery))

	return router
}

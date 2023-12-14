package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	logging "github.com/ipfs/go-log"
	"go.uber.org/fx"
)

type params struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner
	Router     *mux.Router
	Logger     *logging.ZapEventLogger
	Addr       Addr
}

var Module = fx.Options(fx.Invoke(func(p params) error {
	httpServer := &http.Server{
		Addr:         string(p.Addr),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      http.TimeoutHandler(p.Router, 30*time.Second, "server timeout"),
	}
	log := p.Logger.Named("http-server")

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				log.Infow("starting http-server...", "addr", string(p.Addr))
				shutdownStatus := httpServer.ListenAndServe()
				log.Infow("http-server was started", "status", shutdownStatus)

				_ = p.Shutdowner.Shutdown() //nolint:errcheck
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Infow("http-server is stopping...")
			return httpServer.Shutdown(ctx)
		},
	})

	return nil
}))

package accounts

import (
	"net/http"

	logging "github.com/ipfs/go-log"
	"github.com/olehmushka/distributed-social/schemas"
	"github.com/olehmushka/distributed-social/server"
	"github.com/olehmushka/distributed-social/utils/httputils"
	"go.uber.org/fx"
)

type Handlers interface {
	Ping(rw http.ResponseWriter, r *http.Request)
	Info(rw http.ResponseWriter, r *http.Request)
}

type HandlersImpl struct {
	Logger *logging.ZapEventLogger
	Name   server.Name
}

type handlersParams struct {
	fx.In

	Logger *logging.ZapEventLogger
	Name   server.Name
}

func NewHandlers(params handlersParams) Handlers {
	return &HandlersImpl{
		Logger: params.Logger,
		Name:   params.Name,
	}
}

func (h *HandlersImpl) Ping(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)
	writer.WriteSuccess(ctx, &schemas.GetPingRespData{Ok: true})
}

func (h *HandlersImpl) Info(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)
	writer.WriteSuccess(ctx, &schemas.GetInfoRespData{
		Name: string(h.Name),
	})
}

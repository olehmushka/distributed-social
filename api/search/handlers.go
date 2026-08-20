package search

import (
	"errors"
	"net/http"
	"strconv"

	logging "github.com/ipfs/go-log"
	"go.uber.org/fx"

	searchdomain "github.com/olehmushka/distributed-social/internal/search"
	"github.com/olehmushka/distributed-social/schemas"
	"github.com/olehmushka/distributed-social/server"
	"github.com/olehmushka/distributed-social/utils/httputils"
)

type Handlers interface {
	Ping(rw http.ResponseWriter, r *http.Request)
	Info(rw http.ResponseWriter, r *http.Request)
	Search(rw http.ResponseWriter, r *http.Request)
}

type HandlersImpl struct {
	Logger  *logging.ZapEventLogger
	Name    server.Name
	Service *searchdomain.Service
}

type handlersParams struct {
	fx.In

	Logger  *logging.ZapEventLogger
	Name    server.Name
	Service *searchdomain.Service
}

func NewHandlers(params handlersParams) Handlers {
	return &HandlersImpl{
		Logger:  params.Logger,
		Name:    params.Name,
		Service: params.Service,
	}
}

func (h *HandlersImpl) Ping(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)
	writer.WriteSuccess(ctx, &schemas.GetPingRespData{Ok: true}) //nolint:errcheck
}

func (h *HandlersImpl) Info(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)
	writer.WriteSuccess(ctx, &schemas.GetInfoRespData{ //nolint:errcheck
		Name: string(h.Name),
	})
}

func (h *HandlersImpl) Search(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	results, err := h.Service.Search(ctx, query, limit, offset)
	if err != nil {
		if errors.Is(err, searchdomain.ErrEmptyQuery) {
			writer.WriteFail(ctx, err, nil) //nolint:errcheck
			return
		}
		writer.WriteError(ctx, err) //nolint:errcheck
		return
	}

	resp := schemas.SearchRespData{Results: make([]schemas.SearchResultRespData, 0, len(results))}
	for _, r := range results {
		resp.Results = append(resp.Results, schemas.SearchResultRespData{
			PostID:         r.PostID,
			AuthorID:       r.AuthorID,
			AuthorUsername: r.AuthorUsername,
			Content:        r.Content,
			CreatedAt:      r.CreatedAt,
			Rank:           r.Rank,
		})
	}
	writer.WriteSuccess(ctx, resp) //nolint:errcheck
}

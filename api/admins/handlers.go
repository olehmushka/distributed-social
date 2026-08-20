package admins

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	logging "github.com/ipfs/go-log"
	"go.uber.org/fx"

	admindomain "github.com/olehmushka/distributed-social/internal/admins"
	"github.com/olehmushka/distributed-social/schemas"
	"github.com/olehmushka/distributed-social/server"
	"github.com/olehmushka/distributed-social/utils/httputils"
)

type Handlers interface {
	Ping(rw http.ResponseWriter, r *http.Request)
	Info(rw http.ResponseWriter, r *http.Request)
	SuspendUser(rw http.ResponseWriter, r *http.Request)
	RestoreUser(rw http.ResponseWriter, r *http.Request)
	RemovePost(rw http.ResponseWriter, r *http.Request)
	ListActions(rw http.ResponseWriter, r *http.Request)
}

type HandlersImpl struct {
	Logger  *logging.ZapEventLogger
	Name    server.Name
	Service *admindomain.Service
}

type handlersParams struct {
	fx.In

	Logger  *logging.ZapEventLogger
	Name    server.Name
	Service *admindomain.Service
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

type moderationRequest struct {
	ActorID string `json:"actorId"`
	Reason  string `json:"reason"`
}

func (h *HandlersImpl) SuspendUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	var req moderationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writer.WriteFail(ctx, err, nil) //nolint:errcheck
		return
	}

	action, err := h.Service.SuspendUser(ctx, req.ActorID, mux.Vars(r)["id"], req.Reason)
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}
	writer.WriteSuccess(ctx, toActionResp(action)) //nolint:errcheck
}

func (h *HandlersImpl) RestoreUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	var req moderationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writer.WriteFail(ctx, err, nil) //nolint:errcheck
		return
	}

	action, err := h.Service.RestoreUser(ctx, req.ActorID, mux.Vars(r)["id"], req.Reason)
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}
	writer.WriteSuccess(ctx, toActionResp(action)) //nolint:errcheck
}

func (h *HandlersImpl) RemovePost(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	var req moderationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writer.WriteFail(ctx, err, nil) //nolint:errcheck
		return
	}

	action, err := h.Service.RemovePost(ctx, req.ActorID, mux.Vars(r)["id"], req.Reason)
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}
	writer.WriteSuccess(ctx, toActionResp(action)) //nolint:errcheck
}

func (h *HandlersImpl) ListActions(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	actions, err := h.Service.ListActions(ctx, limit, offset)
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}

	resp := schemas.ListModerationActionsRespData{Actions: make([]schemas.ModerationActionRespData, 0, len(actions))}
	for _, a := range actions {
		resp.Actions = append(resp.Actions, toActionResp(a))
	}
	writer.WriteSuccess(ctx, resp) //nolint:errcheck
}

func toActionResp(a admindomain.ModerationAction) schemas.ModerationActionRespData {
	return schemas.ModerationActionRespData{
		ID:         a.ID,
		ActorID:    a.ActorID,
		TargetType: string(a.Target),
		TargetID:   a.TargetID,
		Action:     string(a.Action),
		Reason:     a.Reason,
		CreatedAt:  a.CreatedAt,
	}
}

type statusWriter interface {
	WriteFail(ctx context.Context, err error, data interface{}) error
	WriteStatus(ctx context.Context, statusCode int, err error) error
}

func writeServiceError(ctx context.Context, writer statusWriter, err error) {
	switch {
	case errors.Is(err, admindomain.ErrNotFound):
		writer.WriteStatus(ctx, http.StatusNotFound, err) //nolint:errcheck
	default:
		writer.WriteFail(ctx, err, nil) //nolint:errcheck
	}
}

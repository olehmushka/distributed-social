package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	logging "github.com/ipfs/go-log"
	"go.uber.org/fx"

	accountsdomain "github.com/olehmushka/distributed-social/internal/accounts"
	"github.com/olehmushka/distributed-social/schemas"
	"github.com/olehmushka/distributed-social/server"
	"github.com/olehmushka/distributed-social/utils/httputils"
)

type Handlers interface {
	Ping(rw http.ResponseWriter, r *http.Request)
	Info(rw http.ResponseWriter, r *http.Request)
	CreateUser(rw http.ResponseWriter, r *http.Request)
	GetUser(rw http.ResponseWriter, r *http.Request)
	CreatePost(rw http.ResponseWriter, r *http.Request)
	GetPost(rw http.ResponseWriter, r *http.Request)
	ListUserPosts(rw http.ResponseWriter, r *http.Request)
}

type HandlersImpl struct {
	Logger  *logging.ZapEventLogger
	Name    server.Name
	Service *accountsdomain.Service
}

type handlersParams struct {
	fx.In

	Logger  *logging.ZapEventLogger
	Name    server.Name
	Service *accountsdomain.Service
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

type createUserRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

func (h *HandlersImpl) CreateUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writer.WriteFail(ctx, err, nil) //nolint:errcheck
		return
	}

	user, err := h.Service.CreateUser(ctx, req.Username, req.DisplayName)
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}

	writer.WriteSuccess(ctx, toUserResp(user)) //nolint:errcheck
}

func (h *HandlersImpl) GetUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	user, err := h.Service.GetUser(ctx, mux.Vars(r)["id"])
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}

	writer.WriteSuccess(ctx, toUserResp(user)) //nolint:errcheck
}

type createPostRequest struct {
	Content string `json:"content"`
}

func (h *HandlersImpl) CreatePost(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writer.WriteFail(ctx, err, nil) //nolint:errcheck
		return
	}

	post, err := h.Service.CreatePost(ctx, mux.Vars(r)["id"], req.Content)
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}

	writer.WriteSuccess(ctx, toPostResp(post)) //nolint:errcheck
}

func (h *HandlersImpl) GetPost(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	post, err := h.Service.GetPost(ctx, mux.Vars(r)["id"])
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}

	writer.WriteSuccess(ctx, toPostResp(post)) //nolint:errcheck
}

func (h *HandlersImpl) ListUserPosts(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	writer := httputils.NewWriter(h.Logger, rw)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	posts, err := h.Service.ListUserPosts(ctx, mux.Vars(r)["id"], limit, offset)
	if err != nil {
		writeServiceError(ctx, writer, err)
		return
	}

	resp := schemas.ListPostsRespData{Posts: make([]schemas.PostRespData, 0, len(posts))}
	for _, p := range posts {
		resp.Posts = append(resp.Posts, toPostResp(p))
	}
	writer.WriteSuccess(ctx, resp) //nolint:errcheck
}

func toUserResp(u accountsdomain.User) schemas.UserRespData {
	return schemas.UserRespData{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Status:      string(u.Status),
		CreatedAt:   u.CreatedAt,
	}
}

func toPostResp(p accountsdomain.Post) schemas.PostRespData {
	return schemas.PostRespData{
		ID:        p.ID,
		AuthorID:  p.AuthorID,
		Content:   p.Content,
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
	}
}

type statusWriter interface {
	WriteFail(ctx context.Context, err error, data interface{}) error
	WriteStatus(ctx context.Context, statusCode int, err error) error
}

func writeServiceError(ctx context.Context, writer statusWriter, err error) {
	switch {
	case errors.Is(err, accountsdomain.ErrNotFound):
		writer.WriteStatus(ctx, http.StatusNotFound, err) //nolint:errcheck
	case errors.Is(err, accountsdomain.ErrUsernameTaken):
		writer.WriteStatus(ctx, http.StatusConflict, err) //nolint:errcheck
	case errors.Is(err, accountsdomain.ErrAuthorSuspended):
		writer.WriteStatus(ctx, http.StatusForbidden, err) //nolint:errcheck
	default:
		writer.WriteFail(ctx, err, nil) //nolint:errcheck
	}
}

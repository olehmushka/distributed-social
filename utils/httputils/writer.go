package httputils

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/carlmjohnson/versioninfo"
	logging "github.com/ipfs/go-log"
	"github.com/olehmushka/distributed-social/schemas"
	"github.com/olehmushka/distributed-social/utils/contextutils"
	"github.com/olehmushka/distributed-social/utils/stringutils"
	"github.com/olehmushka/distributed-social/utils/timeutils"
	"go.uber.org/zap"
)

type writer struct {
	rw http.ResponseWriter
	l  *logging.ZapEventLogger
}

func NewWriter(l *logging.ZapEventLogger, rw http.ResponseWriter) *writer {
	return &writer{
		rw: rw,
		l:  l,
	}
}

func CreateSuccess[T interface{}](ctx context.Context, payload T) (*schemas.SuccessResp[T], error) {
	metadata, err := GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	return &schemas.SuccessResp[T]{
		Data:     payload,
		Metadata: metadata,
		Status:   schemas.SuccessStatus,
	}, nil
}

func (w *writer) WriteSuccess(ctx context.Context, payload interface{}) error {
	resp, err := CreateSuccess(ctx, payload)
	if err != nil {
		return err
	}
	w.writeHeaders(nil, http.StatusOK)
	return w.write(resp)
}

func CreateFail[T interface{}](ctx context.Context, err error, data T) (*schemas.FailureResp[T], error) {
	metadata, err := GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	return &schemas.FailureResp[T]{
		Data:     data,
		Message:  err.Error(),
		Metadata: metadata,
		Status:   schemas.FailureStatus,
	}, nil
}

func (w *writer) WriteFail(ctx context.Context, inErr error, data interface{}) error {
	if inErr == nil {
		return nil
	}

	resp, err := CreateFail(ctx, inErr, data)
	if err != nil {
		return err
	}
	w.writeHeaders(nil, http.StatusBadRequest)
	return w.write(resp)
}

func CreateError(ctx context.Context, err error) (*schemas.ErrorResp, error) {
	metadata, err := GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	return &schemas.ErrorResp{
		Message:  err.Error(),
		Metadata: metadata,
		Status:   schemas.ErrorStatus,
	}, nil
}

func (w *writer) WriteError(ctx context.Context, inErr error) error {
	if inErr == nil {
		return nil
	}

	resp, err := CreateError(ctx, inErr)
	if err != nil {
		return err
	}
	w.writeHeaders(nil, http.StatusInternalServerError)
	return w.write(resp)
}

func (w *writer) write(payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := w.rw.Write(b); err != nil {
		return err
	}

	return nil
}

func (w *writer) writeHeaders(headers map[string]string, statusCode int) {
	w.rw.Header().Set("Content-Type", "application/json")
	for key, value := range headers {
		w.rw.Header().Set(key, value)
	}
	w.rw.WriteHeader(statusCode)
}

func WriteSuccess(ctx context.Context, l *logging.ZapEventLogger, rw http.ResponseWriter, payload interface{}) {
	if err := NewWriter(l, rw).WriteSuccess(ctx, payload); err != nil {
		l.Errorw("can not write success response", zap.Error(err))
	}
}

func WriteFail(ctx context.Context, l *logging.ZapEventLogger, rw http.ResponseWriter, inErr error, data interface{}) {
	if err := NewWriter(l, rw).WriteFail(ctx, inErr, data); err != nil {
		l.Errorw("can not write fail response", zap.Error(err))
	}
}

func WriteError(ctx context.Context, l *logging.ZapEventLogger, rw http.ResponseWriter, inErr error) {
	if err := NewWriter(l, rw).WriteError(ctx, inErr); err != nil {
		l.Errorw("can not write error response", zap.Error(err))
	}
}

func GetMetadata(ctx context.Context) (*schemas.Metadata, error) {
	duration, err := GetRequestDuration(ctx)
	if err != nil {
		return nil, err
	}

	return &schemas.Metadata{
		RequestID: contextutils.GetValueFromContext(ctx, contextutils.RequestIDKey),
		Timestamp: timeutils.TimeToString(time.Now()),
		Duration:  duration,
		Version:   versioninfo.Short(),
	}, nil
}

func GetRequestDuration(ctx context.Context) (int, error) {
	str := contextutils.GetValueFromContext(ctx, contextutils.StartRequestTimestampKey)
	if str == "" {
		return 0, nil
	}
	start, err := stringutils.StringToInt64(str)
	if err != nil {
		return 0, err
	}
	diff := time.Now().UnixNano() - start

	return int(diff), nil
}

package middlewares

import (
	"net/http"

	logging "github.com/ipfs/go-log"
	"github.com/olehmushka/distributed-social/utils/contextutils"
	"github.com/olehmushka/distributed-social/utils/httputils"
)

func NewLoggerMiddleware(logger *logging.ZapEventLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			duration, err := httputils.GetRequestDuration(ctx)
			if err != nil {
				_ = httputils.NewWriter(logger, rw).WriteFail(ctx, err, nil) //nolint:errcheck
				logger.Errorw("can not get request duration", "err", err)
				return
			}

			h.ServeHTTP(rw, r)
			logger.Infow("request processed",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", duration,
				"ipAddress", contextutils.GetValueFromContext(ctx, contextutils.IPAddressKey),
				"requestId", contextutils.GetValueFromContext(ctx, contextutils.RequestIDKey),
			)
		})
	}

}

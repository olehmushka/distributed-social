package middlewares

import (
	"fmt"
	"net/http"

	logging "github.com/ipfs/go-log"
	"github.com/olehmushka/distributed-social/utils/httputils"
)

func NewRecoverMiddleware(logger *logging.ZapEventLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			defer func() {
				if err := recover(); err != nil {
					_ = httputils.NewWriter(logger, rw).WriteError(ctx, fmt.Errorf("unexpected error (err=%v)", err)) //nolint:errcheck
					logger.Errorw("can not get request duration", "err", err)
					return
				}
			}()

			next.ServeHTTP(rw, r)
		}

		return http.HandlerFunc(fn)
	}
}

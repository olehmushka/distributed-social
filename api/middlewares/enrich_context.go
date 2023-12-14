package middlewares

import (
	"fmt"
	"net/http"
	"time"

	logging "github.com/ipfs/go-log"
	"github.com/olehmushka/distributed-social/utils/contextutils"
	"github.com/olehmushka/distributed-social/utils/httputils"
	"github.com/olehmushka/distributed-social/utils/netutils"
)

func NewEnrichContextMiddleware(logger *logging.ZapEventLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			writer := httputils.NewWriter(logger, rw)

			ip, err := netutils.ExtractIPAddress(r)
			if err != nil {
				_ = writer.WriteFail(ctx, err, nil) //nolint:errcheck
				logger.Errorw("can not extract IP address from request", "err", err)
				return
			}
			ctx = contextutils.SetValue(ctx, contextutils.IPAddressKey, ip)
			ctx = contextutils.SetValue(ctx, contextutils.RequestIDKey, httputils.ExtractRequestID(r))
			ctx = contextutils.SetValue(ctx, contextutils.StartRequestTimestampKey, fmt.Sprint(time.Now().UnixNano()))

			next.ServeHTTP(rw, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

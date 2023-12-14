package contextutils

import "context"

type ContextKey string

const (
	RequestIDKey             ContextKey = "k_request_id"
	StartRequestTimestampKey ContextKey = "k_start_request_timestamp"
	IPAddressKey             ContextKey = "k_ip_address"
)

func GetValueFromContext(ctx context.Context, key ContextKey) string {
	if value, ok := ctx.Value(key).(string); ok {
		return value
	}

	return ""
}

func SetValue(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

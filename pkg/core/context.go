package core

import "context"

type ctxKey string

func InjectCtxValue(ctx context.Context, key string, value any) context.Context {
	return context.WithValue(ctx, ctxKey(key), value)
}

func ExtractCtxValue(ctx context.Context, key string) any {
	return ctx.Value(ctxKey(key))
}

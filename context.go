package inertia

import "context"

type contextKey string

const (
	contextKeyViewData       = contextKey("viewData")
	contextKeyProps          = contextKey("props")
	contextKeyOptionalProps  = contextKey("optionalProps")
	contextKeyAlwaysProps    = contextKey("alwaysProps")
	contextKeyDeferredProps  = contextKey("deferredProps")
	contextKeyMergeProps     = contextKey("mergeProps")
	contextKeyDeepMergeProps = contextKey("deepMergeProps")
	contextKeyPrependProps   = contextKey("prependProps")
	contextKeyScrollProp     = contextKey("scrollProp")
	contextKeyOnceProps      = contextKey("onceProps")
	contextKeyOnce           = contextKey("once")
	contextKeyErrors         = contextKey("errors")
	contextKeyFlash          = contextKey("flash")
	contextKeyClearHistory   = contextKey("clearHistory")
	contextKeyEncryptHistory = contextKey("encryptHistory")
)

type contextDeferredProp struct {
	Group string
	Value func() any
}

type contextMergeableProp struct {
	MatchOn []string
	Value   func() any
}

func contextGet[T any](ctx context.Context, key contextKey) (T, error) {
	v := ctx.Value(key)
	if v == nil {
		var zero T

		return zero, nil
	}

	value, ok := v.(T)
	if !ok {
		var zero T

		return zero, ErrInvalidContextValue
	}

	return value, nil
}

func contextSet[T any](ctx context.Context, key contextKey, propKey string, value T) context.Context {
	v := ctx.Value(key)

	if v != nil {
		props, ok := v.(map[string]T)
		if ok {
			props[propKey] = value

			return context.WithValue(ctx, key, props)
		}
	}

	return context.WithValue(ctx, key, map[string]T{
		propKey: value,
	})
}

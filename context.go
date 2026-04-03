package inertia

import "context"

const (
	// ContextKeyFuncMap key.
	ContextKeyFuncMap = contextKey("funcMap")

	// ContextKeyViewData key.
	ContextKeyViewData = contextKey("viewData")

	// ContextKeyProps key.
	ContextKeyProps = contextKey("props")

	// ContextKeyDeferredProps key.
	ContextKeyDeferredProps = contextKey("deferredProps")

	// ContextKeyMergeProps key.
	ContextKeyMergeProps = contextKey("mergeProps")

	// ContextKeyDeepMergeProps key.
	ContextKeyDeepMergeProps = contextKey("deepMergeProps")

	// ContextKeyPrependProps key.
	ContextKeyPrependProps = contextKey("prependProps")

	// ContextKeyOptionalProps key.
	ContextKeyOptionalProps = contextKey("optionalProps")

	// ContextKeyAlwaysProps key.
	ContextKeyAlwaysProps = contextKey("alwaysProps")

	// ContextKeyOnceProps key.
	ContextKeyOnceProps = contextKey("onceProps")

	// ContextKeyClearHistory key.
	ContextKeyClearHistory = contextKey("clearHistory")

	// ContextKeyEncryptHistory key.
	ContextKeyEncryptHistory = contextKey("encryptHistory")
)

type contextKey string

type contextDeferredProp struct {
	Group string
	Value func() any
}

func contextProp[T any](ctx context.Context, key contextKey, propKey string, value T) context.Context {
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

func contextValue[T any](ctx context.Context, key contextKey) (T, error) {
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

package inertia

import "context"

type contextKey string

// ContextKeyFuncMap key.
const ContextKeyFuncMap = contextKey("funcMap")

// ContextKeyViewData key.
const ContextKeyViewData = contextKey("viewData")

// ContextKeyProps key.
const ContextKeyProps = contextKey("props")

// ContextKeyDeferredProps key.
const ContextKeyDeferredProps = contextKey("deferredProps")

// ContextKeyMergeProps key.
const ContextKeyMergeProps = contextKey("mergeProps")

// ContextKeyDeepMergeProps key.
const ContextKeyDeepMergeProps = contextKey("deepMergeProps")

// ContextKeyPrependProps key.
const ContextKeyPrependProps = contextKey("prependProps")

// ContextKeyOptionalProps key.
const ContextKeyOptionalProps = contextKey("optionalProps")

// ContextKeyAlwaysProps key.
const ContextKeyAlwaysProps = contextKey("alwaysProps")

// ContextKeyOnceProps key.
const ContextKeyOnceProps = contextKey("onceProps")

// ContextKeyClearHistory key.
const ContextKeyClearHistory = contextKey("clearHistory")

// ContextKeyEncryptHistory key.
const ContextKeyEncryptHistory = contextKey("encryptHistory")

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

// ContextValueDeferredProp type.
type ContextValueDeferredProp struct {
	Group string
	Value func() any
}

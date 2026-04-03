package inertia

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

// ContextEntryDeferredProp type.
type ContextEntryDeferredProp struct {
	Group string
	Value func() any
}

// ContextEntryLazyProp type.
type ContextEntryLazyProp struct {
	Value func() any
}

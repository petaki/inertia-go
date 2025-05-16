package inertia

type contextKey string

// ContextKeyProps key.
const ContextKeyProps = contextKey("props")

// ContextKeyFuncMap key.
const ContextKeyFuncMap = contextKey("funcMap")

// ContextKeyViewData key.
const ContextKeyViewData = contextKey("viewData")

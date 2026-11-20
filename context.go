package inertia

import (
	"context"
	"fmt"
)

type contextKey string

const (
	contextKeyViewData         = contextKey("viewData")
	contextKeyProps            = contextKey("props")
	contextKeyOptionalProps    = contextKey("optionalProps")
	contextKeyAlwaysProps      = contextKey("alwaysProps")
	contextKeyDeferredProps    = contextKey("deferredProps")
	contextKeyMergeProps       = contextKey("mergeProps")
	contextKeyDeepMergeProps   = contextKey("deepMergeProps")
	contextKeyPrependProps     = contextKey("prependProps")
	contextKeyScrollProp       = contextKey("scrollProp")
	contextKeyOnceProps        = contextKey("onceProps")
	contextKeyOnce             = contextKey("once")
	contextKeyErrors           = contextKey("errors")
	contextKeyFlash            = contextKey("flash")
	contextKeyClearHistory     = contextKey("clearHistory")
	contextKeyEncryptHistory   = contextKey("encryptHistory")
	contextKeyPreserveFragment = contextKey("preserveFragment")
)

type contextDeferredProp struct {
	Group string
	Value func() (any, error)
}

type contextMergeableProp struct {
	MatchOn []string
	Value   func() any
}

type contextNode[T any] struct {
	key    string
	value  T
	parent *contextNode[T]
}

func (n *contextNode[T]) values() map[string]T {
	values := make(map[string]T)

	for node := n; node != nil; node = node.parent {
		if _, ok := values[node.key]; !ok {
			values[node.key] = node.value
		}
	}

	return values
}

func contextGet[T any](ctx context.Context, key contextKey) (map[string]T, error) {
	v := ctx.Value(key)
	if v == nil {
		return nil, nil
	}

	node, ok := v.(*contextNode[T])
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidContextValue, key)
	}

	return node.values(), nil
}

func contextSet[T any](ctx context.Context, key contextKey, propKey string, value T) context.Context {
	parent, _ := ctx.Value(key).(*contextNode[T])

	return context.WithValue(ctx, key, &contextNode[T]{
		key:    propKey,
		value:  value,
		parent: parent,
	})
}

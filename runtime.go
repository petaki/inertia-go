package inertia

import "net/http"

type runtime struct {
	isPartial      bool
	isSsrEnabled   bool
	version        string
	ssrURL         string
	ssrClient      *http.Client
	props          map[string]any
	sharedProps    map[string]any
	sharedViewData map[string]any
	only           map[string]struct{}
	except         map[string]struct{}
	exceptOnce     map[string]struct{}
	once           map[string]struct{}
	reset          map[string]struct{}
}

func (rt *runtime) isCachedOnce(key string) bool {
	_, isOnce := rt.once[key]
	_, isCached := rt.exceptOnce[key]
	_, isRequested := rt.only[key]

	return isOnce && isCached && !isRequested
}

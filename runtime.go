package inertia

import (
	"net/http"
	"strings"
)

type runtime struct {
	isPartial  bool
	props      map[string]any
	only       map[string]struct{}
	except     map[string]struct{}
	exceptOnce map[string]struct{}
}

func newRuntime(r *http.Request, component string, props map[string]any) *runtime {
	rt := &runtime{
		props:      props,
		only:       make(map[string]struct{}),
		except:     make(map[string]struct{}),
		exceptOnce: make(map[string]struct{}),
	}

	if r.Header.Get(HeaderPartialComponent) == component {
		if partial := r.Header.Get(HeaderPartialOnly); partial != "" {
			rt.isPartial = true

			for value := range strings.SplitSeq(partial, ",") {
				rt.only[value] = struct{}{}
			}
		}

		if partialExcept := r.Header.Get(HeaderPartialExcept); partialExcept != "" {
			rt.isPartial = true

			for value := range strings.SplitSeq(partialExcept, ",") {
				rt.except[value] = struct{}{}
			}
		}
	}

	if exceptOnceHeader := r.Header.Get(HeaderExceptOnceProps); exceptOnceHeader != "" {
		for value := range strings.SplitSeq(exceptOnceHeader, ",") {
			rt.exceptOnce[value] = struct{}{}
		}
	}

	return rt
}

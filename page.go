package inertia

// OncePageProp type.
type OncePageProp struct {
	Prop      string `json:"prop"`
	ExpiresAt *int64 `json:"expiresAt,omitempty"`
}

// ScrollPageProp type.
type ScrollPageProp struct {
	PageName     string `json:"pageName"`
	CurrentPage  any    `json:"currentPage"`
	PreviousPage any    `json:"previousPage,omitempty"`
	NextPage     any    `json:"nextPage,omitempty"`
	Reset        bool   `json:"reset"`
}

// Page type.
type Page struct {
	Component      string                    `json:"component"`
	Props          map[string]any            `json:"props"`
	URL            string                    `json:"url"`
	Version        string                    `json:"version"`
	DeferredProps  map[string][]string       `json:"deferredProps,omitempty"`
	MergeProps     []string                  `json:"mergeProps,omitempty"`
	DeepMergeProps []string                  `json:"deepMergeProps,omitempty"`
	PrependProps   []string                  `json:"prependProps,omitempty"`
	MatchPropsOn   []string                  `json:"matchPropsOn,omitempty"`
	ScrollProps    map[string]ScrollPageProp `json:"scrollProps,omitempty"`
	OnceProps      map[string]OncePageProp   `json:"onceProps,omitempty"`
	Flash          map[string]any            `json:"flash,omitempty"`
	ClearHistory   bool                      `json:"clearHistory,omitempty"`
	EncryptHistory bool                      `json:"encryptHistory,omitempty"`
}

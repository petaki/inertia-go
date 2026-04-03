package inertia

// Page type.
type Page struct {
	Component      string              `json:"component"`
	Props          map[string]any      `json:"props"`
	URL            string              `json:"url"`
	Version        string              `json:"version"`
	DeferredProps  map[string][]string `json:"deferredProps,omitempty"`
	MergeProps     []string            `json:"mergeProps,omitempty"`
	DeepMergeProps []string            `json:"deepMergeProps,omitempty"`
	PrependProps   []string            `json:"prependProps,omitempty"`
	ClearHistory   bool                `json:"clearHistory,omitempty"`
	EncryptHistory bool                `json:"encryptHistory,omitempty"`
}

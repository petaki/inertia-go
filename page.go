package inertia

type Page struct {
	Component string                 `json:"component"`
	Props     map[string]interface{} `json:"props"`
	Url       string                 `json:"url"`
	Version   string                 `json:"version"`
}

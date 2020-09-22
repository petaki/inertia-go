package inertia

import (
	"encoding/json"
	"html/template"
)

func marshal(v interface{}) (template.JS, error) {
	js, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return template.JS(js), nil
}

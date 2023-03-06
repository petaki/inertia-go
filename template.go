package inertia

import (
	"encoding/json"
	"html/template"
	"strings"
)

func marshal(v interface{}) (template.JS, error) {
	js, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return template.JS(js), nil
}

func lines(elems []string) (template.HTML, error) {
	html := strings.Join(elems, "\n")

	return template.HTML(html), nil
}

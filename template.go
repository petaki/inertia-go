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

func raw(v interface{}) (template.HTML, error) {
	elems, ok := v.([]string)
	if ok {
		html := strings.Join(elems, "\n")

		return template.HTML(html), nil
	}

	elem, ok := v.(string)
	if ok {
		return template.HTML(elem), nil
	}

	return "", nil
}

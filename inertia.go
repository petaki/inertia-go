package inertia

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

type Inertia struct {
	url           string
	rootTemplate  string
	version       string
	sharedProps   map[string]interface{}
	sharedFuncMap template.FuncMap
}

func New(url, rootTemplate, version string) *Inertia {
	i := new(Inertia)
	i.url = url
	i.rootTemplate = rootTemplate
	i.version = version
	i.sharedProps = make(map[string]interface{})
	i.sharedFuncMap = template.FuncMap{"marshal": marshal}

	return i
}

func (i *Inertia) Share(key string, value interface{}) {
	i.sharedProps[key] = value
}

func (i *Inertia) ShareFunc(key string, value interface{}) {
	i.sharedFuncMap[key] = value
}

func (i *Inertia) WithProp(r *http.Request, key string, value interface{}) error {
	ctx := r.Context()
	contextProps := ctx.Value(ContextKeyProps)

	if contextProps != nil {
		contextProps, ok := contextProps.(map[string]interface{})
		if !ok {
			return ErrInvalidContextProps
		}

		contextProps[key] = value
		ctx = context.WithValue(ctx, ContextKeyProps, contextProps)
	} else {
		ctx = context.WithValue(ctx, ContextKeyProps, map[string]interface{}{
			key: value,
		})
	}

	r = r.WithContext(ctx)

	return nil
}

func (i *Inertia) WithViewData(r *http.Request, key string, value interface{}) error {
	ctx := r.Context()
	contextViewData := ctx.Value(ContextKeyViewData)

	if contextViewData != nil {
		contextViewData, ok := contextViewData.(map[string]interface{})
		if !ok {
			return ErrInvalidContextViewData
		}

		contextViewData[key] = value
		ctx = context.WithValue(ctx, ContextKeyViewData, contextViewData)
	} else {
		ctx = context.WithValue(ctx, ContextKeyViewData, map[string]interface{}{
			key: value,
		})
	}

	r = r.WithContext(ctx)

	return nil
}

func (i *Inertia) Render(w http.ResponseWriter, r *http.Request, component string, props map[string]interface{}) error {
	only := make(map[string]string)
	partial := r.Header.Get("X-Inertia-Partial-Data")

	if partial != "" && r.Header.Get("X-Inertia-Partial-Component") == component {
		for _, value := range strings.Split(partial, ",") {
			only[value] = value
		}
	}

	page := &Page{
		Component: component,
		Props:     make(map[string]interface{}),
		Url:       r.RequestURI,
		Version:   i.version,
	}

	for key, value := range i.sharedProps {
		if _, ok := only[key]; len(only) == 0 || ok {
			page.Props[key] = value
		}
	}

	contextProps := r.Context().Value(ContextKeyProps)

	if contextProps != nil {
		contextProps, ok := contextProps.(map[string]interface{})
		if !ok {
			return ErrInvalidContextProps
		}

		for key, value := range contextProps {
			if _, ok := only[key]; len(only) == 0 || ok {
				page.Props[key] = value
			}
		}
	}

	for key, value := range props {
		if _, ok := only[key]; len(only) == 0 || ok {
			page.Props[key] = value
		}
	}

	if r.Header.Get("X-Inertia") != "" {
		js, err := json.Marshal(page)
		if err != nil {
			return err
		}

		w.Header().Set("Vary", "Accept")
		w.Header().Set("X-Inertia", "true")
		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(js)
		if err != nil {
			return err
		}

		return nil
	}

	viewData := make(map[string]interface{})
	contextViewData := r.Context().Value(ContextKeyViewData)

	if contextViewData != nil {
		contextViewData, ok := contextViewData.(map[string]interface{})
		if !ok {
			return ErrInvalidContextViewData
		}

		for key, value := range contextViewData {
			viewData[key] = value
		}
	}

	viewData["page"] = page

	ts, err := template.New(filepath.Base(i.rootTemplate)).Funcs(i.sharedFuncMap).ParseFiles(i.rootTemplate)
	if err != nil {
		return err
	}

	err = ts.Execute(w, viewData)
	if err != nil {
		return err
	}

	return nil
}

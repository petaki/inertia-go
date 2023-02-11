package inertia

import (
	"context"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// Inertia type.
type Inertia struct {
	url           string
	rootTemplate  string
	version       string
	sharedProps   map[string]interface{}
	sharedFuncMap template.FuncMap
	templateFS    fs.FS
	ssrURL        string
}

// New function.
func New(url, rootTemplate, version string) *Inertia {
	i := new(Inertia)
	i.url = url
	i.rootTemplate = rootTemplate
	i.version = version
	i.sharedProps = make(map[string]interface{})
	i.sharedFuncMap = template.FuncMap{"marshal": marshal}

	return i
}

// NewWithFS function.
func NewWithFS(url, rootTemplate, version string, templateFS fs.FS) *Inertia {
	i := New(url, rootTemplate, version)
	i.templateFS = templateFS

	return i
}

// IsSsrEnabled function.
func (i *Inertia) IsSsrEnabled() bool {
	return i.ssrURL != ""
}

// EnableSsr function.
func (i *Inertia) EnableSsr(ssrURL string) {
	i.ssrURL = ssrURL
}

// EnableSsrWithDefault function.
func (i *Inertia) EnableSsrWithDefault() {
	i.EnableSsr("http://127.0.0.1:13714")
}

// DisableSsr function.
func (i *Inertia) DisableSsr() {
	i.ssrURL = ""
}

// Share function.
func (i *Inertia) Share(key string, value interface{}) {
	i.sharedProps[key] = value
}

// ShareFunc function.
func (i *Inertia) ShareFunc(key string, value interface{}) {
	i.sharedFuncMap[key] = value
}

// WithProp function.
func (i *Inertia) WithProp(ctx context.Context, key string, value interface{}) context.Context {
	contextProps := ctx.Value(ContextKeyProps)

	if contextProps != nil {
		contextProps, ok := contextProps.(map[string]interface{})
		if ok {
			contextProps[key] = value

			return context.WithValue(ctx, ContextKeyProps, contextProps)
		}
	}

	return context.WithValue(ctx, ContextKeyProps, map[string]interface{}{
		key: value,
	})
}

// WithViewData function.
func (i *Inertia) WithViewData(ctx context.Context, key string, value interface{}) context.Context {
	contextViewData := ctx.Value(ContextKeyViewData)

	if contextViewData != nil {
		contextViewData, ok := contextViewData.(map[string]interface{})
		if ok {
			contextViewData[key] = value

			return context.WithValue(ctx, ContextKeyViewData, contextViewData)
		}
	}

	return context.WithValue(ctx, ContextKeyViewData, map[string]interface{}{
		key: value,
	})
}

// Render function.
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
		URL:       r.RequestURI,
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

	if i.isInertiaRequest(r) {
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

	ts, err := i.createRootTemplate()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")

	err = ts.Execute(w, viewData)
	if err != nil {
		return err
	}

	return nil
}

// Location function.
func (i *Inertia) Location(w http.ResponseWriter, r *http.Request, url string) {
	if i.isInertiaRequest(r) {
		w.Header().Set("X-Inertia-Location", url)
		w.WriteHeader(http.StatusConflict)
	} else {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func (i *Inertia) isInertiaRequest(r *http.Request) bool {
	return r.Header.Get("X-Inertia") != ""
}

func (i *Inertia) createRootTemplate() (*template.Template, error) {
	ts := template.New(filepath.Base(i.rootTemplate)).Funcs(i.sharedFuncMap)

	if i.templateFS != nil {
		return ts.ParseFS(i.templateFS, i.rootTemplate)
	}

	return ts.ParseFiles(i.rootTemplate)
}

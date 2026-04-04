<p align="center"><img src="https://github.com/user-attachments/assets/856bd49b-7bff-4f29-9a9e-fc84e7ab1b49" width="320" alt="Inertia GO"></p>

# Inertia.js Go Adapter

[![Build Status](https://github.com/petaki/inertia-go/workflows/tests/badge.svg)](https://github.com/petaki/inertia-go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen.svg)](LICENSE.md)

An Inertia.js server-side adapter for Go. Visit [inertiajs.com](https://inertiajs.com) to learn more.

## Installation

Install the package using the `go get` command:

```
go get github.com/petaki/inertia-go
```

## Usage

### 1. Create new instance

```go
url := "http://inertia-app.test" // Application URL for redirect
rootTemplate := "./app.gohtml"   // Root template, see the example below
version := ""                    // Asset version

inertiaManager := inertia.New(url, rootTemplate, version)
```

Or create with `embed.FS` for root template:

```go
import "embed"

//go:embed template
var templateFS embed.FS

// ...

inertiaManager := inertia.New(url, rootTemplate, version, templateFS)
```

### 2. Register the middleware

```go
mux := http.NewServeMux()
mux.Handle("/", inertiaManager.Middleware(homeHandler))
```

### 3. Render in handlers

```go
func homeHandler(w http.ResponseWriter, r *http.Request) {
    // ...

    err := inertiaManager.Render(w, r, "home/Index", nil)
    if err != nil {
        // Handle server error...
    }
}
```

Or render with props:

```go
// ...

err := inertiaManager.Render(w, r, "home/Index", map[string]any{
    "total": 32,
})

//...
```

### 4. Server-side Rendering (Optional)

First, enable SSR with the url of the Node server:

```go
inertiaManager.EnableSsrWithDefault() // http://127.0.0.1:13714/render
```

Or with custom url:

```go
inertiaManager.EnableSsr("http://ssr-host:13714/render")
```

Or with the Vite dev server:

```go
inertiaManager.EnableSsr("http://localhost:5173/__inertia_ssr")
```

You can also provide a custom `*http.Client`:

```go
client := &http.Client{
    Timeout: 10 * time.Second,
}

inertiaManager.EnableSsr("http://ssr-host:13714/render", client)
inertiaManager.EnableSsrWithDefault(client)
```

For more information, please read the official Server-side Rendering documentation on [inertiajs.com](https://inertiajs.com).

## Examples

The following examples show how to use the package.

### Share a function with root template (globally)

```go
inertiaManager.ShareFunc("asset", assetFunc)
```

```html
<script src="{{ asset "js/app.js" }}"></script>
```

### Share data with root template (globally)

```go
inertiaManager.ShareViewData("env", "production")
```

```html
{{ if eq .env "production" }}
    ...
{{ end }}
```

### Share data with root template (context based)

```go
ctx := inertiaManager.WithViewData(r.Context(), "meta", meta)
r = r.WithContext(ctx)
```

```html
<meta name="description" content="{{ .meta }}">
```

### Props comparison

| Prop Type | Method(s) | Evaluation | Full Render | Partial Render |
|-----------|-----------|------------|-------------|----------------|
| Base | `Share`, `WithProp`, `Render` | Eager | Included | Included if requested |
| Optional | `WithOptionalProp` | Lazy | Excluded | Included if requested |
| Always | `WithAlwaysProp` | Lazy | Included | Always included |
| Deferred | `WithDeferredProp` | Lazy | Excluded (deferred) | Included if requested |
| Merge | `WithMergeProp` | Lazy | Included | Included if requested |
| Deep Merge | `WithDeepMergeProp` | Lazy | Included | Included if requested |
| Prepend | `WithPrependProp` | Lazy | Included | Included if requested |
| Scroll | `WithScrollProp` | — | Metadata only | Metadata only |
| Once | `WithOnceProp`, `WithOnce` | Lazy | Included | Excluded if in except-once |
| Flash | `WithFlashProp` | Eager | Included | Included |

`WithOnce` can be combined with Deferred, Merge, Deep Merge, Prepend, and Optional props.
`WithScrollProp` adds scroll metadata to the page response for infinite scroll support.

### Share a prop (globally)

```go
inertiaManager.Share("title", "Inertia App Title")
```

### Share a prop (context based)

```go
func authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
        
        ctx := inertiaManager.WithProp(r.Context(), "authUserID", user.ID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Optional prop (context based)

```go
ctx := inertiaManager.WithOptionalProp(r.Context(), "extra", func() any {
    return getExtra()
})
r = r.WithContext(ctx)
```

### Always prop (context based)

```go
ctx := inertiaManager.WithAlwaysProp(r.Context(), "errors", func() any {
    return getErrors()
})
r = r.WithContext(ctx)
```

### Deferred prop (context based)

```go
ctx := inertiaManager.WithDeferredProp(r.Context(), "comments", func() any {
    return getComments()
})
r = r.WithContext(ctx)
```

### Deferred prop with group (context based)

```go
ctx := inertiaManager.WithDeferredProp(r.Context(), "comments", func() any {
    return getComments()
}, "my-group")
r = r.WithContext(ctx)
```

### Merge prop (context based)

```go
ctx := inertiaManager.WithMergeProp(r.Context(), "results", func() any {
    return getResults()
})
r = r.WithContext(ctx)
```

Or with match on:

```go
ctx := inertiaManager.WithMergeProp(r.Context(), "results", func() any {
    return getResults()
}, "id")
r = r.WithContext(ctx)
```

Or with multiple nested match on paths:

```go
ctx := inertiaManager.WithMergeProp(r.Context(), "complexData", func() any {
    return getComplexData()
}, "users.data.id", "messages.uuid")
r = r.WithContext(ctx)
```

### Deep merge prop (context based)

```go
ctx := inertiaManager.WithDeepMergeProp(r.Context(), "settings", func() any {
    return getSettings()
})
r = r.WithContext(ctx)
```

Or with match on:

```go
ctx := inertiaManager.WithDeepMergeProp(r.Context(), "settings", func() any {
    return getSettings()
}, "id")
r = r.WithContext(ctx)
```

### Prepend prop (context based)

```go
ctx := inertiaManager.WithPrependProp(r.Context(), "notifications", func() any {
    return getNotifications()
})
r = r.WithContext(ctx)
```

Or with match on:

```go
ctx := inertiaManager.WithPrependProp(r.Context(), "notifications", func() any {
    return getNotifications()
}, "id")
r = r.WithContext(ctx)
```

### Scroll prop (context based)

```go
ctx := inertiaManager.WithScrollProp(r.Context(), "items", inertia.ScrollPageProp{
    PageName:    "page",
    CurrentPage: 1,
    NextPage:    2,
})
r = r.WithContext(ctx)
```

### Once prop (context based)

```go
ctx := inertiaManager.WithOnceProp(r.Context(), "plans", func() any {
    return getPlans()
})
r = r.WithContext(ctx)
```

### Once modifier (context based)

```go
ctx := inertiaManager.WithMergeProp(r.Context(), "activity", func() any {
    return getActivity()
})
ctx = inertiaManager.WithOnce(ctx, "activity", inertia.OncePageProp{})
r = r.WithContext(ctx)
```

Or with expiration:

```go
expiresAt := time.Now().Add(24 * time.Hour).UnixMilli()
ctx := inertiaManager.WithDeferredProp(r.Context(), "permissions", func() any {
    return getPermissions()
})
ctx = inertiaManager.WithOnce(ctx, "permissions", inertia.OncePageProp{ExpiresAt: &expiresAt})
r = r.WithContext(ctx)
```

### Flash (context based)

```go
ctx := inertiaManager.WithFlashProp(r.Context(), map[string]any{
    "success": "Item created successfully",
})
r = r.WithContext(ctx)
```

### Clear history (context based)

```go
ctx := inertiaManager.WithClearHistory(r.Context())
r = r.WithContext(ctx)
```

### Encrypt history (context based)

```go
ctx := inertiaManager.WithEncryptHistory(r.Context())
r = r.WithContext(ctx)
```

### Root template

```html
<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link href="css/app.css" rel="stylesheet">
        <link rel="icon" type="image/x-icon" href="favicon.ico">
    </head>
    <body>
        <script data-page="app" type="application/json">{{ marshal .page }}</script>
        <div id="app"></div>
        <script src="js/app.js"></script>
    </body>
</html>
```

### Root template with Server-side Rendering (SSR)

```html
<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link href="css/app.css" rel="stylesheet">
        <link rel="icon" type="image/x-icon" href="favicon.ico">
        {{ if .ssr }}
            {{ raw .ssr.Head }}
        {{ end }}
    </head>
    <body>
        {{ if not .ssr }}
            <script data-page="app" type="application/json">{{ marshal .page }}</script>
            <div id="app"></div>
        {{ else }}
            {{ raw .ssr.Body }}
        {{ end }}
        <script src="js/app.js"></script>
    </body>
</html>
```

## Example Apps

### Satellite

<img src="https://github.com/user-attachments/assets/216e9052-4a28-4540-9702-4f039ba0ecda" width="16" alt="Vite"> Vite /
<img src="https://github.com/user-attachments/assets/5fb47ac4-cac5-4820-9701-8ea48fa426fc" width="16" alt="Vue3"> Vue3

https://github.com/petaki/satellite

### Homettp

<img src="https://github.com/user-attachments/assets/216e9052-4a28-4540-9702-4f039ba0ecda" width="16" alt="Vite"> Vite /
<img src="https://github.com/user-attachments/assets/5fb47ac4-cac5-4820-9701-8ea48fa426fc" width="16" alt="Vue3"> Vue3

https://github.com/homettp/homettp

### Waterkube

<img src="https://github.com/user-attachments/assets/216e9052-4a28-4540-9702-4f039ba0ecda" width="16" alt="Vite"> Vite /
<img src="https://github.com/user-attachments/assets/5fb47ac4-cac5-4820-9701-8ea48fa426fc" width="16" alt="Vue3"> Vue3

https://github.com/waterkube/waterkube

## Contributors

- [@monstergron](https://github.com/monstergron) for logo ([ArtStation](https://www.artstation.com/danielmakaro))

## Reporting Issues

If you are facing a problem with this package or found any bug, please open an issue on [GitHub](https://github.com/petaki/inertia-go/issues).

## License

The MIT License (MIT). Please see [License File](LICENSE.md) for more information.

# Inertia.js Go Adapter

[![Build Status](https://github.com/petaki/inertia-go/workflows/tests/badge.svg)](https://github.com/petaki/inertia-go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen.svg)](LICENSE.md)

The Inertia.js server-side adapter for Go. Visit [inertiajs.com](https://inertiajs.com) to learn more.

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

inertiaManager := inertia.NewWithFS(url, rootTemplate, version, templateFS)
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

err := inertiaManager.Render(w, r, "home/Index", map[string]interface{}{
    "total": 32,
})

//...
```

## Examples

The following examples show how to use the package.

### Share a prop globally

```go
inertiaManager.Share("title", "Inertia App Title")
```

### Share a function with root template

```go
inertiaManager.ShareFunc("asset", assetFunc)
```

```html
<script src="{{ asset "js/app.js" }}"></script>
```

### Share a prop from middleware

```go
func authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
        
        ctx := inertiaManager.WithProp(r.Context(), "authUserId", user.Id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Share data with root template

```go
ctx := inertiaManager.WithViewData(r.Context(), "meta", meta)
r = r.WithContext(ctx)
```

```html
<meta name="description" content="{{ .meta }}">
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
        <div id="app" data-page="{{ marshal .page }}"></div>
        <script src="js/app.js"></script>
    </body>
</html>
```

## Example Apps

### Satellite

https://github.com/petaki/satellite

### Homettp

https://github.com/homettp/homettp

## Reporting Issues

If you are facing a problem with this package or found any bug, please open an issue on [GitHub](https://github.com/petaki/inertia-go/issues).

## License

The MIT License (MIT). Please see [License File](LICENSE.md) for more information.

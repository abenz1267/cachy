[![Go Report Card](https://goreportcard.com/badge/abenz1267/cachy)](https://goreportcard.com/report/abenz1267/cachy)

# Cachy - HTML Template Caching Library

Cachy is a simple caching library for templates using Go's html/template package.

## Features
- accepts explicit folders or scans/watches complete directory
- execute single templates
- execute multiple templates (as with ParseFiles("file1", "file2"...))
- filewatcher that updates the cache on template changes
- reloading browser on template change via JavaScript fetch

## Usage

### Simple Example:

```go
c, _ := cachy.New("", ".html", nil) // this will process all *.html files, no FuncMap.
go c.Watch(true) // starts the filewatcher, logging enabled

_ := c.Execute(w, nil, "folder/template", "folder/template2") // io.Writer, data, templates...
```

### Hot-Reloading browser via JavaScript

```go
...

c, _ := cachy.New("/reload", ".html", nil)
go c.Watch(false)

http.Handle("/reload", http.HandlerFunc(c.HotReload))

...
```

In your template you simply have to execute the "reloadScript" template-function.

```go
// end of <body> or wherever you want
{{ reloadScript }}
```

## Benchmarks

```
BenchmarkExecuteSingleTemplate-16    5000000	       338 ns/op	      96 B/op	       2 allocs/op
BenchmarkExecuteDualTemplate-16    	 3000000	       432 ns/op	     144 B/op	       3 allocs/op
```

If you have suggestions or feedback, feel free to contact me! PRs or Issues are welcomed!

I hope this little library is useful to some.

Regards

Andrej Benz
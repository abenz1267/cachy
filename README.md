[![Go Report Card](https://goreportcard.com/badge/github.com/abenz1267/cachy)](https://goreportcard.com/report/github.com/abenz1267/cachy)

# Cachy - HTML Template Caching Library

Cachy is a simple caching library for templates using Go's html/template package.

## Features
- accepts explicit folders or scans/watches complete directory
- execute single templates
- execute multiple templates (as with ParseFiles("file1", "file2"...))
- filewatcher that updates the cache on template changes
- reloading browser on template change via JavaScript fetch
- allow duplicate template files
- add folders recursively

## Usage

If you explicitly set folders, you can tell Cachy to search for nested folders via the "recursive" argument. If no folder is given, Cachy will look for template files in the current folder recursively.

The "allowDuplicates" parameter checks, if templates with the same filename can co-exist or not. If duplications are disallowed, executing templates is easily done by just providing the filename minus the extension. If duplicates are allowed, you have to include the whole path.

### Simple Example:

```go
c, _ := cachy.New(nil, nil) // this will process all *.html files, no FuncMap, no duplicates, will search for template files within whole working dir
go c.Watch(true) // starts the filewatcher, logging enabled

_ := c.Execute(w, nil, "template", "template2") // io.Writer, data, templates...
```

### Hot-Reloading browser via JavaScript

```go
...

c, _ := cachy.New(nil, nil)
go c.Watch(false)

http.Handle(c.URL(), c.SSE)

...
```

In your template you simply have to execute the "reloadScript" template-function.

```go
// end of <body> or wherever you want
{{ reloadScript }}
```

## Benchmarks

```
BenchmarkDefaultSingle-16        4395246               468 ns/op             807 B/op          2 allocs/op
BenchmarkDefaultMultiple-16      4934824               412 ns/op             743 B/op          2 allocs/op
BenchmarkCachySingle-16          4513342               455 ns/op             789 B/op          2 allocs/op
BenchmarkCachyMultiple-16        3543176               567 ns/op            1013 B/op          3 allocs/op
```

If you have suggestions or feedback, feel free to contact me! PRs or Issues are welcomed!

I hope this little library is useful to some.

Regards

Andrej Benz

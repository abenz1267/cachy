# Cachy - HTML Template Caching Library

Cachy is a simple caching library for templates using Go's html/template package.

## Features
- accepts explicit folders or scans/watches complete directory
- execute single templates
- execute multiple templates (as with ParseFiles("file1", "file2"...))
- filewatcher that updates the cache on template changes
- support for [Packr (v2)](https://github.com/gobuffalo/packr/tree/master/v2) (for embedding templates)

## Usage

Example:

```go
c, _ := cachy.New(".html", nil, nil) // this will process all *.html files, no FuncMap, no Packr boxes.
go c.Watch() // starts the filewatcher

_ := c.Execute(w, nil, "folder/template", "folder/template2") // io.Writer, data, templates...
```

Example when using Packr:

```go
boxes := make(map[string]*packr.Box)
boxes["templates"] = packr.New("templates", "./templates")
c, _ := cachy.New(".html", true, nil, boxes)

_ := c.Execute(w, nil, "templates/someTemplate")
```

As you can see this is pretty straightforward.

## Benchmarks

```
BenchmarkExecuteSingleTemplate-16    5000000	       338 ns/op	      96 B/op	       2 allocs/op
BenchmarkExecuteDualTemplate-16    	 3000000	       432 ns/op	     144 B/op	       3 allocs/op
```

If you have suggestions or feedback, feel free to contact me!

I hope this little library is useful to some.

Regards

Andrej Benz

## Licenses
--------------------------------------------------------------------------------

gobuffalo/packr

The MIT License (MIT)
Copyright (c) 2016 Mark Bates

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

---------------------------------------------------------------------------------
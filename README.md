# Cachy - Template Caching Library

## Features
- accepts explicit folders or scans/watches whole working dir
- single templates
- multiple templates (as with ParseFiles("file1", "file2"...))
- filewatcher that updates the cache on template changes

## Usage

Example:

```go
c, _ := cachy.Init(".html", true, nil) // this will process all *.html files, activate the filewatcher, no FuncMap.

_ := c.Execute(w, nil, "folder/template", "folder/template") // io.Writer, data, templates...
```

As you can see this is pretty straightfoward.

## Benchmarks

```
BenchmarkExecuteSingleTemplate-16    	 5000000	       338 ns/op	      96 B/op	       2 allocs/op
BenchmarkExecuteDualTemplate-16    	 3000000	       432 ns/op	     144 B/op	       3 allocs/op
```

If you have suggestions or feedback, feel free to contact me!

I hope this little library is useful to some.

Regards

Andrej Benz
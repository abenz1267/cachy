package cachy

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDefault(t *testing.T) {
	funcs := template.FuncMap{}
	funcs["test"] = func(msg string) string {
		return msg
	}

	c, err := New(nil, funcs)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			t.Error(err)
		}
	}

	var w bytes.Buffer
	err = c.Execute(&w, nil, "index")
	if err != nil {
		t.Error(err)
	}
}

func TestNewRecursive(t *testing.T) {
	p := &Params{URL: "", Ext: "html", Duplicates: false, Recursive: true}
	c, err := New(p, nil, "test_templates")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			t.Error(err)
		}
	}

	var w bytes.Buffer
	err = c.Execute(&w, nil, "index")
	if err != nil {
		t.Error(err)
	}
}

func TestNewDuplicates(t *testing.T) {
	p := &Params{URL: "/reload", Ext: "html", Duplicates: true, Recursive: false}
	c, err := New(p, nil)
	if err != nil {
		t.Error(err)
	}

	var w bytes.Buffer
	err = c.Execute(&w, nil, "test_templates/index")
	if err != nil {
		t.Error(err)
	}
}

func TestNewDuplicatesRecursive(t *testing.T) {
	p := &Params{URL: "/reload", Ext: "html", Duplicates: true, Recursive: true}
	c, err := New(p, nil)
	if err != nil {
		t.Error(err)
	}

	var w bytes.Buffer
	err = c.Execute(&w, nil, "test_templates/index", "test_templates/base")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkDefaultSingle(b *testing.B) {
	var w bytes.Buffer
	b.ReportAllocs()
	t, _ := template.ParseFiles(filepath.Join("test_templates", "base.html"))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		t.Execute(&w, nil)
	}
}

func BenchmarkDefaultMultiple(b *testing.B) {
	var w bytes.Buffer
	b.ReportAllocs()
	t, _ := template.ParseFiles(filepath.Join("test_templates", "base.html"), filepath.Join("test_templates", "index.html"))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		t.Execute(&w, nil)
	}
}

func BenchmarkCachySingle(b *testing.B) {
	b.ReportAllocs()
	var w bytes.Buffer
	c, _ := New(nil, nil, "test_templates")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Execute(&w, nil, "base")
	}
}

func BenchmarkCachyMultiple(b *testing.B) {
	b.ReportAllocs()
	var w bytes.Buffer
	c, _ := New(nil, nil, "test_templates")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Execute(&w, nil, "base", "index")
	}
}

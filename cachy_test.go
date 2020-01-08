package cachy

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	funcs := template.FuncMap{}
	funcs["test"] = func(msg string) string {
		return msg
	}

	_, err := New("", "html", false, false, funcs, "test_templates")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			t.Error(err)
		}
	}
}

func TestNewRecursive(t *testing.T) {
	_, err := New("", "html", false, true, nil, "test_templates", "test_templates/nested")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			t.Error(err)
		}
	}
}

func TestNewRecursiveAllowDuplicates(t *testing.T) {
	_, err := New("/reload", "html", false, false, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			t.Error(err)
		}
	}
}

func TestNoDuplicates(t *testing.T) {
	c, err := New("", "html", false, false, nil, "test_templates", "test_templates/nested")
	if err != nil {
		t.Error(err)
	}

	var b bytes.Buffer
	err = c.Execute(&b, nil, "base", "nested")
	if err != nil {
		t.Error(err)
	}
}

func TestAllowDuplicates(t *testing.T) {
	c, err := New("", "html", true, false, nil, "test_templates", "test_templates/nested", "test_templates/nested/nested2")
	if err != nil {
		t.Error(err)
	}
	var b bytes.Buffer
	err = c.Execute(&b, nil, "test_templates/base", "test_templates/nested/nested", "test_templates/nested/nested2/nested")
	if err != nil {
		t.Error(err)
	}
}

func TestExecute(t *testing.T) {
	c, err := New("", "html", false, false, nil, "test_templates")
	if err != nil {
		t.Error(err)
	}
	var b bytes.Buffer
	err = c.Execute(&b, nil, "base")
	if err != nil {
		t.Error(err)
	}
}

func TestExecuteMultiple(t *testing.T) {
	c, err := New("", "html", false, false, nil, "test_templates")
	if err != nil {
		t.Error(err)
	}
	var b bytes.Buffer
	err = c.Execute(&b, nil, "base", "index")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkSingle(b *testing.B) {
	var w bytes.Buffer
	c, _ := New("", "html", false, false, nil, "test_templates")
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		c.Execute(&w, nil, "base")
	}
}

func BenchmarkMultiple(b *testing.B) {
	var w bytes.Buffer
	c, _ := New("", "html", false, false, nil, "test_templates")
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		c.Execute(&w, nil, "base", "index")
	}
}

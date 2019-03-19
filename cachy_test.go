package cachy

import (
	"bytes"
	"html/template"
	"testing"
)

func TestNew(t *testing.T) {
	funcs := template.FuncMap{}
	funcs["test"] = func(msg string) string {
		return msg
	}

	_, err := New(".html", funcs)
	if err != nil {
		t.Error(err)
	}
}

func TestExecute(t *testing.T) {
	c, err := New(".html", nil)
	if err != nil {
		t.Error(err)
	}
	var b bytes.Buffer
	err = c.Execute(&b, nil, "test_templates/base")
	if err != nil {
		t.Error(err)
	}
}
func TestExecuteMultiple(t *testing.T) {
	c, err := New(".html", nil)
	if err != nil {
		t.Error(err)
	}
	var b bytes.Buffer
	err = c.Execute(&b, nil, "test_templates/base", "test_templates/index")
	if err != nil {
		t.Error(err)
	}
}

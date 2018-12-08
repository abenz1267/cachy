package cachy

import (
	"bytes"
	"testing"
)

func TestLoad(t *testing.T) {
	c, err := Init(".html", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	err = c.Execute(&b, nil, "test_templates/base", "test_templates/index")
	if err != nil {
		t.Fatal(err)
	}

	b.Reset()
	err = c.Execute(&b, nil, "test_templates/base")
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkExecuteSingleTemplate(b *testing.B) {
	c, err := Init(".html", false, nil)
	if err != nil {
		b.Fatal(err)
	}

	var w bytes.Buffer
	for n := 0; n < b.N; n++ {
		c.Execute(&w, nil, "test_templates/base")
		w.Reset()
	}
}

func BenchmarkExecuteDualTemplate(b *testing.B) {
	c, err := Init(".html", false, nil)
	if err != nil {
		b.Fatal(err)
	}

	var w bytes.Buffer
	for n := 0; n < b.N; n++ {
		c.Execute(&w, nil, "test_templates/base", "test_templates/index")
		w.Reset()
	}
}

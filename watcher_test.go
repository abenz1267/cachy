package cachy

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestUpdateTmpl(t *testing.T) {
	c, err := Init(".html", false, nil, "test_templates")
	if err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	err = c.Execute(&b, nil, "test_templates/base", "test_templates/index")
	if err != nil {
		t.Fatal(err)
	}

	err = updateTmpl("test_templates/index", ".html", &c)
	if err != nil {
		t.Fatal(err)
	}

	deleteTmpl("test_templates/index", &c)
}

func TestWatch(t *testing.T) {
	_, err := Init(".html", true, nil)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("new template")

	time.Sleep(1 * time.Second)

	err = ioutil.WriteFile("test_templates/test.html", data, 0777)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	os.Remove("test_templates/test.html")

	time.Sleep(1 * time.Second)
}

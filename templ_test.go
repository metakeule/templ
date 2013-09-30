package templ

import (
	"bytes"
	"testing"
)

var (
	firstname = NewPlaceholder("firstname")
	lastname  = NewPlaceholder("lastname")
	templ     = New("t").MustAdd("Hello, ", firstname, " ", lastname, "!\n").MustParse()
	expected  = "Hello, Donald Duck!\nHello, Mickey Mouse!\n"
)

func TestTemplate(t *testing.T) {
	var b bytes.Buffer
	templ.MustReplaceTo(&b, firstname.Set("Donald"), lastname.Set("Duck"))
	templ.MustReplaceTo(&b, firstname.Set("Mickey"), lastname.Set("Mouse"))

	if r := b.String(); r != expected {
		t.Errorf("Error in setting: expected\n\t%#v\ngot\n\t%#v\n", expected, r)
	}
}

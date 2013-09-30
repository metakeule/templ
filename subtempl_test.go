package templ

import (
	"testing"
)

var (
	fruit      = NewPlaceholder("fruit")
	li         = New("item").MustAdd("<li>", fruit, "</li>").MustParse()
	ul         = New("list").MustAdd("<ul>", li, "</ul>").MustParse()
	listOutput = "<ul><li>Apple</li><li>Pear</li></ul>"
)

func TestSubTemplate(t *testing.T) {
	all := li.New()

	li.MustReplaceTo(all.Buffer, fruit.Set("Apple"))
	li.MustReplaceTo(all.Buffer, fruit.Set("Pear"))

	if r := ul.MustReplace(all).String(); r != listOutput {
		t.Errorf("Error in setting: expected\n\t%#v\ngot\n\t%#v\n", listOutput, r)
	}
}

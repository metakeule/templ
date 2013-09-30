package main

import (
	"bytes"
	"fmt"
	. "github.com/metakeule/templ"
	"html"
)

type (
	Person struct {
		firstname, lastname string
	}
)

func textEscaper(in interface{}) (out string) { return html.EscapeString(in.(string)) }

var (
	_         = fmt.Println
	persons   = []Person{{"M<i>ckey", "Mouse"}, {"Donald", "Duck"}}
	firstname = NewPlaceholder("firstname", textEscaper)
	lastname  = NewPlaceholder("lastname")
	person    = New("person").MustAdd("\n\t<name>", firstname, " ", lastname, "</name>").MustParse()
	list      = New("list").MustAdd("<person>", person, "\n</person>\n").MustParse()
)

func main() {
	fmt.Println("----------------")
	var b1 = person.New()
	for _, p := range persons {
		person.MustReplaceTo(b1.Buffer, firstname.Set(p.firstname), lastname.Set(p.lastname))
	}
	fmt.Println(list.MustReplace(b1))

	fmt.Println("----------------")
	var b2 bytes.Buffer
	for _, p := range persons {
		list.MustReplaceTo(&b2, person.MustReplace(firstname.Set(p.firstname), lastname.Set(p.lastname)))
	}
	fmt.Println(b2.String())
}

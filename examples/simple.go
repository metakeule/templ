package main

import (
	"bytes"
	"fmt"
	"github.com/metakeule/templ"
	"html"
)

func Html(name string) (t templ.Placeholder) {
	return templ.NewPlaceholder(name)
}

func Text(name string) (t templ.Placeholder) {
	return templ.NewPlaceholder(name, func(in interface{}) (out string) {
		return html.EscapeString(in.(string))
	})
}

var (
	person   = Text("person")
	greeting = Html("greeting")
	t        = templ.New("t")
)

func init() {
	t.MustAdd("<h1>Hi, ", person, "</h1>", greeting).MustParse()
}

func main() {
	fmt.Println(t.MustReplace(person.Set("S<o>meone"), greeting.Set("<div>Hi</div>")))

	var buffer bytes.Buffer
	for i := 0; i < 10; i++ {
		t.MustReplaceTo(&buffer,
			person.Setf("Bugs <Bunny> %v", i+1),
			greeting.Set("<p>How are you?</p>\n"))
	}
	fmt.Println(buffer.String())
}

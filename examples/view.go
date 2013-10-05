package main

import (
	. "github.com/metakeule/goh4"
	. "github.com/metakeule/goh4/attr"
	. "github.com/metakeule/goh4/tag"
	. "github.com/metakeule/goh4/tag/short"
	"github.com/metakeule/templ"
	// "log"
	"net/http"
)

/*
   alle nicht gekennzeichneten felder werden mit dem standart escape (key: "") versehen
   felder, die ignoriert werden sollen, werden entweder kleingeschrieben (ignoriert alle)
   oder mit tag:"-" explicit ausgeklammert

   für alle weiteren tagwerte sind diese eine liste von escapern, die nacheinander angewendet werden
*/

type Figure struct {
	FirstName string   `greet:"-"`
	LastName  string   `greet:"text"`
	Greeting  *Element `greet:"html" person:"-"`
	Width     int      `greet:"px"`
	Link      string   `greet:"urlparam"`
}

var (
	donald = Figure{
		FirstName: "Donald",
		LastName:  "<Duck>",
		Greeting:  P("Are you fine?"),
		Width:     24,
		Link:      "Peter&Paul",
	}

	mickey = Figure{FirstName: "Mickey", LastName: "M<o>use"}

	greet  = View(Figure{}, `greet`)
	person = View(Figure{}, `person`)
	other  = templ.NewPlaceholder("other")

	t = LI(
		Width(greet.Placeholder("Width").String()),
		AHref("/d&p="+greet.Placeholder("Link").String(), greet.Placeholder("LastName")),
		greet.Placeholder("Greeting"),
		DIV(other),
		H1(person.Placeholder("FirstName"), " ", person.Placeholder("LastName")),
	).Compile("entry")

	layout = HTML5(BODY(UL(t))).Compile("layout")
)

func main() {

	handler := func(w http.ResponseWriter, r *http.Request) {
		bf := t.New()
		t.MustReplaceMixedTo(
			bf.Buffer,
			greet.Set(donald),
			person.Set(mickey),
			other.Set("other 1"),
		)
		t.MustReplaceMixedTo(
			bf.Buffer,
			greet.Set(mickey),
			person.Set(donald),
			other.Set("other 2"),
		)
		layout.MustReplace(bf).WriteTo(w)
	}

	http.ListenAndServe("localhost:8080", http.HandlerFunc(handler))
}

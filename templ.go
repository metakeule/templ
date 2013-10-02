package templ

import (
	"bytes"
	"fmt"
	"github.com/metakeule/meta"
	"github.com/metakeule/replacer"
	"io"
	"reflect"
	"strings"
)

// a named buffer, fullfills the Setter interface
type Buffer struct {
	*bytes.Buffer
	name string
}

func newBuffer(name string) *Buffer {
	bf := &Buffer{}
	bf.name = name
	bf.Buffer = &bytes.Buffer{}
	return bf
}

func (b *Buffer) Name() string {
	return b.name
}

type Escaper map[string]func(interface{}) string

type View struct {
	_type         string
	_tag          string
	_placeholders map[string]Placeholder
}

func (esc Escaper) View(stru interface{}, tag string) *View {
	s := &View{_type: structName(stru), _tag: tag}
	s.scanPlaceholders(stru, esc)
	return s
}

func structName(stru interface{}) string {
	return strings.Replace(fmt.Sprintf("%T", stru), "*", "", 1)
}

func (str *View) Tag() string  { return str._tag }
func (str *View) Type() string { return str._type }

func (str *View) Placeholder(field string) Placeholder {
	p, ok := str._placeholders[field]
	if !ok {
		panic(fmt.Sprintf("no placeholder for field %s in struct %s (tag: %s)", field, str._type, str._tag))
	}
	return p
}

func (str *View) HasPlaceholder(field string) bool {
	_, ok := str._placeholders[field]
	return ok
}

func (str *View) Set(stru interface{}) (ss []Setter) {
	if structName(stru) != str._type {
		panic(fmt.Sprintf("wrong type: %T, needed %s or *%s", stru, str._type, str._type))
	}
	for field, ph := range str._placeholders {
		f := meta.Struct.Field(stru, field)
		// we need to handle the nil pointers differently,
		// since they may be handled via interfaces
		// and then they are not nil
		if f.Kind() == reflect.Ptr && f.IsNil() {
			ss = append(ss, ph.Set(nil))
			continue
		}
		ss = append(ss, ph.Set(f.Interface()))
	}
	return
}

func (str *View) scanPlaceholders(stru interface{}, escaper Escaper) {
	str._placeholders = map[string]Placeholder{}
	meta.Struct.EachRaw(stru,
		func(field reflect.StructField, v reflect.Value) {
			phName := fieldName(stru, field.Name, str._tag)
			ph := NewPlaceholder(phName)
			if t := field.Tag.Get(str._tag); t != "" {
				if t != "-" { // "-" signals ignorance
					for _, escaperKey := range strings.Split(t, ",") {
						escFunc, ok := escaper[escaperKey]
						if !ok {
							panic("unknown escaper " + escaperKey + " needed by " + phName)
						}
						ph.Escaper = append(ph.Escaper, escFunc)
					}
					str._placeholders[field.Name] = ph
				}
				return
			}

			escFunc, ok := escaper[""]
			if !ok {
				panic(`missing empty escaper (key: "") needed by ` + phName)
			}
			ph.Escaper = append(ph.Escaper, escFunc)
			str._placeholders[field.Name] = ph
		})
	return
}

type Placeholder struct {
	name, Value string
	Escaper     []func(interface{}) string
}

func NewPlaceholder(name string, escaper ...func(interface{}) string) Placeholder {
	return Placeholder{name: name, Escaper: escaper}
}

func (p Placeholder) Name() string { return p.name }

func (p Placeholder) Set(val interface{}) Setter {
	var value string
	if len(p.Escaper) == 0 {
		value = fmt.Sprintf("%v", val)
	}

	for i, esc := range p.Escaper {
		if i == 0 {
			value = esc(val)
			continue
		}
		value = esc(value)
	}
	return Placeholder{name: p.name, Value: value, Escaper: p.Escaper}
}

func (p Placeholder) Setf(format string, vals ...interface{}) Setter {
	return p.Set(fmt.Sprintf(format, vals...))
}

func (p Placeholder) WriteTo(w io.Writer) (int64, error) {
	i, err := w.Write([]byte(p.Value))
	return int64(i), err
}

type Setter interface {
	io.WriterTo
	Name() string
}

type (
	stringer interface {
		String() string
	}

	writerto interface {
		WriteTo(io.Writer) (int64, error)
	}
)

type Template struct {
	*Buffer
	*replacer.Replacer
}

func New(name string) *Template {
	t := &Template{}
	t.Buffer = newBuffer(name)
	t.Replacer = replacer.New()
	return t
}

func (t *Template) New() *Buffer {
	return newBuffer(t.name)
}

func (t *Template) WriteSetter(p Setter) (err error) {
	_, err = t.Buffer.Write(t.Replacer.Delimiter())
	if err != nil {
		return
	}
	_, err = t.Buffer.WriteString(p.Name())
	if err != nil {
		return
	}
	_, err = t.Buffer.Write(t.Replacer.Delimiter())
	return
}

func (t *Template) MustWriteSetter(s Setter) {
	err := t.WriteSetter(s)
	if err != nil {
		panic(err.Error())
	}
}

func (t *Template) Add(data ...interface{}) (err error) {
	for _, d := range data {
		switch v := d.(type) {
		case Setter:
			err = t.WriteSetter(v)
		case []byte:
			_, err = t.Buffer.Write(v)
		case string:
			_, err = t.Buffer.WriteString(v)
		case byte:
			err = t.Buffer.WriteByte(v)
		case rune:
			_, err = t.Buffer.WriteRune(v)
		case writerto:
			_, err = v.WriteTo(t.Buffer)
		case stringer:
			_, err = t.Buffer.WriteString(v.String())
		default:
			_, err = t.Buffer.WriteString(fmt.Sprintf("%v", v))
		}
		if err != nil {
			return
		}
	}
	return
}

// add data to the template
func (t *Template) MustAdd(data ...interface{}) *Template {
	for _, d := range data {
		switch v := d.(type) {
		case Setter:
			t.MustWriteSetter(v)
		case []byte:
			t.MustWrite(v)
		case string:
			t.MustWriteString(v)
		case byte:
			t.MustWriteByte(v)
		case rune:
			t.MustWriteRune(v)
		case writerto:
			_, err := v.WriteTo(t.Buffer)
			if err != nil {
				panic(err.Error())
			}
		case stringer:
			t.MustWriteString(v.String())
		default:
			t.MustWriteString(fmt.Sprintf("%v", v))
		}
	}
	return t
}

func (t *Template) MustWrite(b []byte) {
	_, err := t.Buffer.Write(b)
	if err != nil {
		panic(err.Error())
	}
}

func (t *Template) MustWriteString(s string) {
	_, err := t.Buffer.WriteString(s)
	if err != nil {
		panic(err.Error())
	}
}

func (t *Template) MustWriteByte(b byte) {
	err := t.Buffer.WriteByte(b)
	if err != nil {
		panic(err.Error())
	}
}

func (t *Template) MustWriteRune(r rune) {
	_, err := t.Buffer.WriteRune(r)
	if err != nil {
		panic(err.Error())
	}
}

func (t *Template) MustWriteTo(w io.Writer) {
	_, err := t.Buffer.WriteTo(w)
	if err != nil {
		panic(err.Error())
	}
}

func (t *Template) Parse() error {
	return t.Replacer.Parse(t.Buffer.Bytes())
}

func (t *Template) MustParse() *Template {
	err := t.Replacer.Parse(t.Buffer.Bytes())
	if err != nil {
		panic(err.Error())
	}
	return t
}

func mixedSetters(mixed ...interface{}) (ss []Setter) {
	for _, m := range mixed {
		switch v := m.(type) {
		case View:
			ss = append(ss, v.Set(v)...)
		case *View:
			ss = append(ss, v.Set(v)...)
		case Setter:
			ss = append(ss, v)
		case []Setter:
			ss = append(ss, v...)
		default:
			panic(fmt.Sprintf("unsupported type: %T, supported are: View, *View, Setter and []Setter", v))
		}
	}
	return
}

func (r *Template) ReplaceMixed(mixed ...interface{}) (bf *Buffer, errors map[string]error) {
	ss := mixedSetters(mixed...)
	return r.Replace(ss...)
}

func (r *Template) ReplaceMixedTo(b *bytes.Buffer, mixed ...interface{}) (bf *Buffer, errors map[string]error) {
	ss := mixedSetters(mixed...)
	return r.ReplaceTo(b, ss...)
}

func (r *Template) MustReplaceMixed(mixed ...interface{}) (bf *Buffer) {
	ss := mixedSetters(mixed...)
	return r.MustReplace(ss...)
}

func (r *Template) MustReplaceMixedTo(b *bytes.Buffer, mixed ...interface{}) (bf *Buffer) {
	ss := mixedSetters(mixed...)
	return r.MustReplaceTo(b, ss...)
}

func (r *Template) ReplaceTo(b *bytes.Buffer, setters ...Setter) (bf *Buffer, errors map[string]error) {
	m := map[string]io.WriterTo{}
	for _, s := range setters {
		m[s.Name()] = s
	}
	errors = r.Set(b, m)
	if len(errors) > 0 {
		return
	}
	bf = &Buffer{Buffer: b, name: r.Name()}
	return
}

// like New but doesn't return errors and panics instead
func (r *Template) MustReplaceTo(b *bytes.Buffer, setters ...Setter) *Buffer {
	bf, errs := r.ReplaceTo(b, setters...)
	if len(errs) > 0 {
		panic(fmt.Sprintf("errors in placeholder: %v", errs))
	}
	return bf
}

// calls Must with a new buffer for every
func (r *Template) Replace(setters ...Setter) (bf *Buffer, errors map[string]error) {
	b := bytes.Buffer{}
	return r.ReplaceTo(&b, setters...)
}

// calls Must with a new buffer for every
func (r *Template) MustReplace(setters ...Setter) (bf *Buffer) {
	b := bytes.Buffer{}
	return r.MustReplaceTo(&b, setters...)
}

/*
// returns a fieldname as used in placeholders of structs
func FieldName(stru interface{}, field string) string {
	f := meta.Struct.Field(stru, field)
	r := fieldName(stru, field)
	if f.Interface() == nil {
		panic("field does not exist: " + r)
	}
	return r
}
*/

func fieldName(stru interface{}, field string, tag string) string {
	return strings.Replace(fmt.Sprintf("%T.%s#%s", stru, field, tag), "*", "", 1)
}

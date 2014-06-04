// This file is part of monsti/form.
// Copyright 2012-2014 Christian Neumann

// monsti/form is free software: you can redistribute it and/or modify it under
// the terms of the GNU Lesser General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.

// monsti/form is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
// details.

// You should have received a copy of the GNU Lesser General Public License
// along with monsti/form. If not, see <http://www.gnu.org/licenses/>.

package form

import (
	"encoding"
	"fmt"
	"html"
	"html/template"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FieldRenderData contains the data needed for field rendering.
type FieldRenderData struct {
	// Lebel is the field's label.
	Label string
	// LabelTag is the html code for the field's label, e.g.
	// `<label for="the_id">The Label</label>`.
	LabelTag template.HTML
	// Input is the input html for the field.
	Input template.HTML
	// Help is the help string.
	Help string
	// Errors contains any validation errors.
	Errors []string
}

// RenderData contains the data needed for form rendering.
type RenderData struct {
	Fields []FieldRenderData
	Errors []string
	// EncTypeAttr is set to 'enctype="multipart/form-data"' if the Form
	// contains a File widget. Should be used as optional attribute for the form
	// element if the form may contain file input elements.
	EncTypeAttr template.HTMLAttr
	Action      string
}

type Widget interface {
	HTML(name string, value interface{}) template.HTML
}

// timeConverter converts a string to a time.Time
func timeConverter(in string) reflect.Value {
	out, err := time.Parse(time.RFC3339, in)
	if err != nil {
		out, err = time.Parse("2006-01-02", in)
	}
	if err != nil {
		out, _ = time.Parse("15:04:05", in)
	}
	return reflect.ValueOf(out)
}

type DateTimeWidget int

func (t DateTimeWidget) HTML(field string, value interface{}) template.HTML {
	var out string
	if obj, ok := value.(time.Time); ok {
		out = obj.Format(time.RFC3339)
	} else if obj, ok := value.(*time.Time); ok {
		if obj == nil {
			out = ""
		} else {
			out = obj.Format(time.RFC3339)
		}
	} else {
		out = fmt.Sprintf("%v", obj)
	}
	return template.HTML(fmt.Sprintf(
		`<input id="%v" type="datetime" name="%v" value="%v"/>`,
		field, field, html.EscapeString(out)))
}

type DateWidget int

func (t DateWidget) HTML(field string, value interface{}) template.HTML {
	var out string
	if obj, ok := value.(time.Time); ok {
		out = obj.Format("2006-01-02")
	} else if obj, ok := value.(*time.Time); ok {
		if obj == nil {
			out = ""
		} else {
			out = obj.Format("2006-01-02")
		}
	} else {
		out = fmt.Sprintf("%v", obj)
	}
	return template.HTML(fmt.Sprintf(
		`<input id="%v" type="date" name="%v" value="%v"/>`,
		field, field, html.EscapeString(out)))
}

type TimeWidget int

func (t TimeWidget) HTML(field string, value interface{}) template.HTML {
	var out string
	if obj, ok := value.(time.Time); ok {
		out = obj.Format("15:04:05")
	} else if obj, ok := value.(*time.Time); ok {
		if obj == nil {
			out = ""
		} else {
			out = obj.Format("15:04:05")
		}
	} else {
		out = fmt.Sprintf("%v", obj)
	}
	return template.HTML(fmt.Sprintf(
		`<input id="%v" type="time" name="%v" value="%v"/>`,
		field, field, html.EscapeString(out)))
}

type Text int

func (t Text) HTML(field string, value interface{}) template.HTML {
	return template.HTML(fmt.Sprintf(
		`<input id="%v" type="text" name="%v" value="%v"/>`,
		field, field, html.EscapeString(
			fmt.Sprintf("%v", value))))
}

type AlohaEditor int

func (t AlohaEditor) HTML(field string, value interface{}) template.HTML {
	return template.HTML(fmt.Sprintf(
		`<textarea class="editor" id="%v" name="%v"/>%v</textarea>`,
		field, field, html.EscapeString(
			fmt.Sprintf("%v", value))))
}

type TextArea int

func (t TextArea) HTML(field string, value interface{}) template.HTML {
	return template.HTML(fmt.Sprintf(
		`<textarea id="%v" name="%v"/>%v</textarea>`,
		field, field, html.EscapeString(
			fmt.Sprintf("%v", value))))
}

// Option of a select widget.
type Option struct {
	Value, Text string
}

// SelectWidget renders a selection field.
type SelectWidget struct {
	Options []Option
}

func (t SelectWidget) HTML(field string, value interface{}) template.HTML {
	var options string
	for _, v := range t.Options {
		selected := ""
		if v.Value == value.(string) {
			selected = " selected"
		}
		options += fmt.Sprintf("<option value=\"%v\"%v>%v</option>\n",
			v.Value, selected, v.Text)
	}
	ret := fmt.Sprintf("<select id=\"%v\" name=\"%v\">\n%v</select>",
		field, field, options)
	return template.HTML(ret)
}

// HiddenWidget renders a hidden input field.
type HiddenWidget int

func (t HiddenWidget) HTML(field string, value interface{}) template.HTML {
	return template.HTML(
		fmt.Sprintf(`<input id="%v" type="hidden" name="%v" value="%v"/>`,
			field, field, value))
}

// PasswordWidget renders a password field.
type PasswordWidget int

func (t PasswordWidget) HTML(field string, value interface{}) template.HTML {
	return template.HTML(
		fmt.Sprintf(`<input id="%v" type="password" name="%v"/>`,
			field, field))
}

// FileWidget renders a file upload field.
type FileWidget int

func (t FileWidget) HTML(field string, value interface{}) template.HTML {
	return template.HTML(
		fmt.Sprintf(`<input id="%v" type="file" name="%v"/>`,
			field, field))
}

// Field contains settings for a form field.
type Field struct {
	Label, Help string
	Validator   Validator
	Widget      Widget
}

// Fields is a map of field names to field settings.
type Fields map[string]Field

// Form represents an html form.
type Form struct {
	Fields map[string]Field
	data   interface{}
	errors map[string][]string
	// Action defines the action parameter of the HTML form
	Action string
}

// NewForm creates a new Form with the given fields with data stored in the
// given pointer to a structure.
//
// In panics if data is not a pointer to a struct.
func NewForm(data interface{}, fields Fields) *Form {
	if dataType := reflect.TypeOf(data); (dataType.Kind() != reflect.Ptr ||
		dataType.Elem().Kind() != reflect.Struct) &&
		dataType.Kind() != reflect.Map {
		panic("NewForm(data, fields) expects data to be a map or a pointer to a struct.")
	}
	form := Form{data: data, Fields: fields,
		errors: make(map[string][]string, len(fields))}
	return &form
}

// RenderData returns a RenderData struct for the form.
//
// It panics if a registered field is not present in the data struct.
func (f Form) RenderData() (renderData RenderData) {
	renderData.Action = f.Action
	renderData.Fields = make([]FieldRenderData, 0)
	for name, field := range f.Fields {
		widget := field.Widget
		if widget == nil {
			widget = new(Text)
		} else if _, ok := widget.(*FileWidget); ok {
			renderData.EncTypeAttr = `enctype="multipart/form-data"`
		}
		value, err := f.getNestedField(name)
		if err != nil {
			value = reflect.ValueOf("")
		}
		renderData.Fields = append(renderData.Fields, FieldRenderData{
			Label: field.Label,
			LabelTag: template.HTML(fmt.Sprintf(`<label for="%v">%v</label>`,
				name, field.Label)),
			Input:  widget.HTML(name, value.Interface()),
			Help:   field.Help,
			Errors: f.errors[name]})
	}
	renderData.Errors = f.errors[""]
	return
}

// AddError adds an error to a field's error list.
//
// To add global form errors, use an empty string as the field's name.
func (f *Form) AddError(field string, error string) {
	if f.errors[field] == nil {
		f.errors[field] = make([]string, 0, 1)
	}
	f.errors[field] = append(f.errors[field], error)
}

const (
	rawField = iota
	boolField
	stringField
)

type fieldType struct {
	IsArray   bool
	ValueType int
}

// getNestedField searches for the given nested field in the given data
func (f Form) getNestedField(field string) (reflect.Value, error) {
	return f.findNestedField(field, nil)
}

// findNestedField searches for the given field in the form data.
//
// If setValue is given, it will be set to the field.
func (f *Form) findNestedField(field string, setValue interface{}) (reflect.Value, error) {
	parts := strings.Split(field, ".")
	value := reflect.ValueOf(f.data)
	for len(parts) != 0 {
		setIt := len(parts) == 1 && setValue != nil
		part := parts[0]
		switch value.Type().Kind() {
		case reflect.Ptr, reflect.Interface:
			value = value.Elem()
			continue
		case reflect.Struct:
			value = value.FieldByName(part)
		case reflect.Map:
			if setIt {
				value.SetMapIndex(reflect.ValueOf(part), reflect.ValueOf(setValue))
				return reflect.Value{}, nil
			}
			value = value.MapIndex(reflect.ValueOf(part))
		default:
			return reflect.Value{},
				fmt.Errorf("form: Can't find field %q in data", field)
		}
		if !value.IsValid() {
			return reflect.Value{},
				fmt.Errorf("form: Invalid field %q in data", field)
		}
		parts = parts[1:]
	}
	if setValue != nil {
		if value.Type().Kind() == reflect.Ptr {
			v := reflect.New(value.Type().Elem())
			v.Elem().Set(reflect.ValueOf(setValue))
			value.Set(v)
		} else {
			value.Set(reflect.ValueOf(setValue))
		}
	}
	if value.Type().Kind() == reflect.Interface {
		value = value.Elem()
	}
	return value, nil
}

// stringToValue converts the given source string to a value of the
// given target type.
func stringToValue(src string, target reflect.Type) interface{} {
	if target.Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
		if target.Kind() == reflect.Ptr {
			target = target.Elem()
		}
		val := reflect.New(target)
		val.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(src))
		return val.Elem().Interface()
	}
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	switch target.Kind() {
	case reflect.String:
		return src
	case reflect.Int:
		v, err := strconv.ParseInt(src, 0, 0)
		if err != nil {
			return 0
		}
		return int(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(src)
		if err != nil {
			return false
		}
		return v
	default:
		panic(fmt.Sprintln("form: Unknown field kind", target.Kind()))
	}
	return nil
}

// setNestedField searches for the given nested field in the given data
func (f *Form) setNestedField(field string, value string) {
	val, err := f.findNestedField(field, nil)
	if err == nil {
		f.findNestedField(field, stringToValue(value, val.Type()))
	}
}

// Fill fills the form data with the given values and validates the form.
//
// It panics if a field has been set up which is not present in the
// data struct.
//
// Values that don't match a field will be ignored.
//
// Returns true iff the form validates.
func (f *Form) Fill(values url.Values) bool {
	for param, paramValue := range values {
		if _, ok := f.Fields[param]; ok {
			fieldValue, err := f.getNestedField(param)
			if err != nil {
				continue
			}
			fieldType := fieldValue.Type()
			if fieldType.Kind() == reflect.Slice {
				fieldType = fieldType.Elem()
			}
			for _, value := range paramValue {
				f.setNestedField(param, value)
			}
		}
	}
	return f.validate()
}

// validate validates the currently present data.
//
// Resets any previous errors.
// Returns true iff the data validates.
func (f *Form) validate() bool {
	anyError := false
	for name, field := range f.Fields {
		value, err := f.getNestedField(name)
		if err != nil {
			return false
		}
		if field.Validator != nil {
			if errors := field.Validator(value.Interface()); errors != nil {
				f.errors[name] = errors
				anyError = true
			}
		}
	}
	return !anyError
}

// Validator is a function which validates the given data and returns error
// messages if the data does not validate.
type Validator func(interface{}) []string

// And is a Validator that collects errors of all given validators.
func And(vs ...Validator) Validator {
	return func(value interface{}) []string {
		errors := []string{}
		for _, v := range vs {
			errors = append(errors, v(value)...)
		}
		if len(errors) == 0 {
			return nil
		}
		return errors
	}
}

// Required creates a Validator to check for non empty values.
//
// msg is set as validation error.
func Required(msg string) Validator {
	return func(value interface{}) []string {
		if value == reflect.Zero(reflect.TypeOf(value)).Interface() {
			return []string{msg}
		}
		return nil
	}
}

// Regex creates a Validator to check a string for a matching regexp.
//
// If the expression does not match the string to be validated,
// the given error msg is returned.
func Regex(exp, msg string) Validator {
	return func(value interface{}) []string {
		if matched, _ := regexp.MatchString(exp, value.(string)); !matched {
			return []string{msg}
		}
		return nil
	}
}

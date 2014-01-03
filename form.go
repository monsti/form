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
	"fmt"
	"html"
	"html/template"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/gorilla/schema"
)

var schemaDecoder = schema.NewDecoder()

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
}

type Widget interface {
	HTML(name string, value interface{}) template.HTML
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
}

// NewForm creates a new Form with the given fields with data stored in the
// given pointer to a structure.
//
// In panics if data is not a pointer to a struct.
func NewForm(data interface{}, fields Fields) *Form {
	if dataType := reflect.TypeOf(data); dataType.Kind() != reflect.Ptr ||
		dataType.Elem().Kind() != reflect.Struct {
		panic("NewForm(data, fields) expects data to be a pointer to a struct.")
	}
	form := Form{data: data, Fields: fields,
		errors: make(map[string][]string, len(fields))}
	return &form
}

// RenderData returns a RenderData struct for the form.
//
// It panics if a registered field is not present in the data struct.
func (f Form) RenderData() (renderData RenderData) {
	dataVal := reflect.ValueOf(f.data).Elem()
	renderData.Fields = make([]FieldRenderData, 0, dataVal.NumField()-1)
	for name, field := range f.Fields {
		widget := field.Widget
		if widget == nil {
			widget = new(Text)
		} else if _, ok := widget.(*FileWidget); ok {
			renderData.EncTypeAttr = `enctype="multipart/form-data"`
		}
		matchFunc := func(field string) bool {
			if strings.ToLower(field) == strings.ToLower(name) {
				return true
			}
			return false
		}
		fieldVal := dataVal.FieldByNameFunc(matchFunc)
		if !fieldVal.IsValid() {
			panic("form: Registered field not present in data struct: " + name)
		}
		renderData.Fields = append(renderData.Fields, FieldRenderData{
			Label: field.Label,
			LabelTag: template.HTML(fmt.Sprintf(`<label for="%v">%v</label>`,
				name, field.Label)),
			Input:  widget.HTML(name, fieldVal.Interface()),
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

// Fill fills the form data with the given values and validates the form.
//
// It panics if a field has been set up which is not present in the
// data struct.
//
// Returns true iff the form validates.
func (f *Form) Fill(values url.Values) bool {
	decoder := schema.NewDecoder()
	decoder.RegisterConverter(time.Time{}, timeConverter)
	error := decoder.Decode(f.data, values)
	switch e := error.(type) {
	case nil:
		return f.validate()
	case schema.MultiError:
		for field, msg := range e {
			v, ok := f.errors[field]
			switch {
			case !ok:
				f.errors[""] = []string{msg.Error()}
			case v == nil:
				f.errors[field] = []string{msg.Error()}
			default:
				f.errors[field] = append(f.errors[field], msg.Error())
			}
		}
		return false
	default:
		panic(error.Error())
	}
}

// validate validates the currently present data.
//
// Resets any previous errors.
// Returns true iff the data validates.
func (f *Form) validate() bool {
	anyError := false
	for name, field := range f.Fields {
		value := reflect.ValueOf(f.data).Elem().FieldByName(name)
		if value == reflect.ValueOf(nil) {
			panic(fmt.Sprintf("Field '%v' not present in form data structure.",
				name))
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

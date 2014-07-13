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
	"html/template"
	"net/url"
	"reflect"
	"testing"
	"time"
)

type TestDataEmbed struct {
	Title string
}

type TestData struct {
	TestDataEmbed
	Name  string
	Age   int
	Extra map[string]interface{}
}

func TestRender(t *testing.T) {
	data := TestData{}
	data.Extra = make(map[string]interface{})
	data.Extra["ExtraField"] = ""
	form := NewForm(&data, []Field{
		Field{"Title", "Your title", "", nil, nil},
		Field{"Name", "Your name", "Your full name", Required("Req!"), nil},
		Field{"Age", "Your age", "Years since your birth.", Required("Req!"), nil},
		Field{"Extra.ExtraField", "Extra Field", "", nil, nil},
	})
	vals := url.Values{
		"Title":            []string{""},
		"Name":             []string{""},
		"Age":              []string{"14"},
		"Extra.ExtraField": []string{"Hey!"},
	}
	form.Fill(vals)
	form.Action = "targetURL"
	renderData := form.RenderData()
	if renderData.Action != "targetURL" {
		t.Errorf(`renderData.Action = %q, should be "targetURL"`, renderData.Action)
	}
	fieldTests := []struct {
		Field    string
		Expected FieldRenderData
	}{
		{
			Field: "Title",
			Expected: FieldRenderData{
				Label:    "Your title",
				LabelTag: `<label for="Title">Your title</label>`,
				Help:     "",
				Errors:   nil,
				Input:    `<input id="Title" type="text" name="Title" value=""/>`}},
		{
			Field: "Name",
			Expected: FieldRenderData{
				Label:    "Your name",
				LabelTag: `<label for="Name">Your name</label>`,
				Help:     "Your full name",
				Errors:   []string{"Req!"},
				Input:    `<input id="Name" type="text" name="Name" value=""/>`}},
		{
			Field: "AGE",
			Expected: FieldRenderData{
				Label:    "Your age",
				LabelTag: `<label for="Age">Your age</label>`,
				Help:     "Years since your birth.",
				Errors:   nil,
				Input:    `<input id="Age" type="text" name="Age" value="14"/>`}},
		{
			Field: "ExtraField",
			Expected: FieldRenderData{
				Label:    "Extra Field",
				LabelTag: `<label for="Extra.ExtraField">Extra Field</label>`,
				Help:     "",
				Errors:   nil,
				Input:    `<input id="Extra.ExtraField" type="text" name="Extra.ExtraField" value="Hey!"/>`}},
	}
	for i, test := range fieldTests {
		if len(renderData.Errors) > 0 {
			t.Errorf("RenderData contains general errors: %v", renderData.Errors)
		}
		if !reflect.DeepEqual(renderData.Fields[i], test.Expected) {
			t.Errorf("RenderData for Field '%v' =\n%v,\nexpected\n%v",
				test.Field, renderData.Fields[i], test.Expected)
		}
	}
}

func TestMapRender(t *testing.T) {
	data := make(map[string]interface{})
	data["Name"] = new(string)
	data["Age"] = new(int)
	data["Foo"] = map[string]string{
		"Bar": "ee"}

	form := NewForm(data, []Field{
		Field{"Name", "Your name", "Your full name", Required("Req!"), nil},
		Field{"Age", "Your age", "Years since your birth.", Required("Req!"), nil},
		Field{"Foo.Bar", "Bar", "Some foo's bar.", Required("Req!"), nil},
	})
	vals := url.Values{
		"Name":    []string{""},
		"Age":     []string{"14"},
		"Foo.Bar": []string{"Bla"},
	}
	form.Fill(vals)
	renderData := form.RenderData()
	fieldTests := []struct {
		Field    string
		Expected FieldRenderData
	}{
		{
			Field: "Name",
			Expected: FieldRenderData{
				Label:    "Your name",
				LabelTag: `<label for="Name">Your name</label>`,
				Help:     "Your full name",
				Errors:   []string{"Req!"},
				Input:    `<input id="Name" type="text" name="Name" value=""/>`}},
		{
			Field: "AGE",
			Expected: FieldRenderData{
				Label:    "Your age",
				LabelTag: `<label for="Age">Your age</label>`,
				Help:     "Years since your birth.",
				Errors:   nil,
				Input:    `<input id="Age" type="text" name="Age" value="14"/>`}},
		{
			Field: "Foo.Bar",
			Expected: FieldRenderData{
				Label:    "Bar",
				LabelTag: `<label for="Foo.Bar">Bar</label>`,
				Help:     "Some foo's bar.",
				Errors:   nil,
				Input:    `<input id="Foo.Bar" type="text" name="Foo.Bar" value="Bla"/>`}},
	}
	for i, test := range fieldTests {
		if len(renderData.Errors) > 0 {
			t.Errorf("RenderData contains general errors: %v", renderData.Errors)
		}
		if !reflect.DeepEqual(renderData.Fields[i], test.Expected) {
			t.Errorf("RenderData for Field '%v' =\n%v,\nexpected\n%v",
				test.Field, renderData.Fields[i], test.Expected)
		}
	}
}

func TestAddError(t *testing.T) {
	data := TestData{}
	form := NewForm(&data, []Field{
		Field{"Name", "Your name", "Your full name", Required("Req!"), nil},
		Field{"Age", "Your age", "Years since your birth.", Required("Req!"), nil}})
	form.AddError("Name", "Foo")
	form.AddError("", "Bar")
	renderData := form.RenderData()
	if len(renderData.Fields[0].Errors) != 1 ||
		renderData.Fields[0].Errors[0] != "Foo" {
		t.Errorf(`Field "Name" should have error "Foo"`)
	}
	if len(renderData.Errors) != 1 ||
		renderData.Errors[0] != "Bar" {
		t.Errorf(`Missing global error "Bar"`)
	}
	if len(renderData.Fields[1].Errors) != 0 {
		t.Errorf(`Field "Bar" should have no errors`)
	}
}

type TestDataEncTypeAttr struct {
	Name string
	File string
}

func TestEncTypeAttr(t *testing.T) {
	data := TestDataEncTypeAttr{}
	vals := url.Values{
		"Name": []string{""}}
	fieldTests := []struct {
		Form    *Form
		EncType string
	}{
		{
			Form: NewForm(&data, []Field{
				Field{"Name", "Your name", "Your full name", Required("Req!"),
					nil},
				Field{"File", "File Dummy", "", nil, nil}}),
			EncType: ""},
		{
			Form: NewForm(&data, []Field{
				Field{"Name", "Your name", "Your full name", Required("Req!"), nil},
				Field{"File", "File!", "", nil, new(FileWidget)}}),
			EncType: `enctype="multipart/form-data"`}}

	for i, v := range fieldTests {
		v.Form.Fill(vals)
		renderData := v.Form.RenderData()
		if string(renderData.EncTypeAttr) != v.EncType {
			t.Errorf("Test %v: RenderData.EncTypeAttr is %q, should be %q", i,
				renderData.EncTypeAttr, v.EncType)
		}
	}
}

func TestFill(t *testing.T) {
	data := TestData{}
	data.Extra = make(map[string]interface{}, 0)
	data.Extra["Number"] = new(int)
	form := NewForm(&data, []Field{
		Field{"Name", "Your name", "Your full name", Required("Req!"), nil},
		Field{"Age", "Your age", "Years since your birth.", Required("Req!"), nil},
		Field{"Extra.Number", "Number", "", nil, nil},
	})
	vals := url.Values{
		"Name":         []string{"Foo"},
		"Age":          []string{"14"},
		"Foo":          []string{"noting here"},
		"Extra.Number": []string{"10"},
	}
	expected := TestData{Name: "Foo", Age: 14}
	expected.Extra = make(map[string]interface{}, 0)
	number := 10
	expected.Extra["Number"] = number
	if !form.Fill(vals) {
		t.Errorf("form.Fill(..) returns false, should be true. Errors: %v",
			form.RenderData().Errors)
	}
	if !reflect.DeepEqual(expected, data) {
		t.Errorf("Filled data should be %v, is %v", expected, data)
	}
	vals["Name"] = []string{""}
	data.Name = ""
	if form.Fill(vals) {
		t.Errorf("form.Fill(..) returns true, should be false.")
	}
}

func TestRequire(t *testing.T) {
	invalid, valid := "", "foo"
	validator := Required("Req!")
	err := validator(valid)
	if err != nil {
		t.Errorf("require(%v) = %v, want %v", valid, err, nil)
	}
	err = validator(invalid)
	if err == nil {
		t.Errorf("require(%v) = %v, want %v", invalid, err, "'Required.'")
	}
}

func TestRegex(t *testing.T) {
	tests := []struct {
		Exp    string
		String string
		Valid  bool
	}{
		{"^[\\w]+$", "", false},
		{"^[\\w]+$", "foobar", true},
		{"", "", true},
		{"", "foobar", true},
		{"^[^!]+$", "foobar", true},
		{"^[^!]+$", "foo!bar", false}}

	for _, v := range tests {
		ret := Regex(v.Exp, "damn!")(v.String)
		if (ret == nil && !v.Valid) || (ret != nil && v.Valid) {
			t.Errorf(`Regex("%v")("%v") = %v, this is wrong!`, v.Exp, v.String,
				ret)
		}
	}
}

func TestAnd(t *testing.T) {
	tests := []struct {
		String     string
		Validators []Validator
		Valid      bool
	}{
		{"Hey! 1", []Validator{Required("Req!")}, true},
		{"", []Validator{Required("Req!")}, false},
		{"Hey! 2", []Validator{Required("Req!"), Regex("Oink", "No way!")}, false},
		{"Hey! 3", []Validator{Required("Req!"), Regex("Hey", "No way!")}, true}}
	for _, v := range tests {
		ret := And(v.Validators...)(v.String)
		if (ret == nil && !v.Valid) || (ret != nil && v.Valid) {
			t.Errorf(`And(...)("%v") = %v, this is wrong!`, v.String, ret)
		}
	}
}

func TestSelectWidget(t *testing.T) {
	widget := SelectWidget{[]Option{
		Option{"foo", "The Foo!"},
		Option{"bar", "The Bar!"}}}
	tests := []struct {
		Name, Value, Expected string
	}{
		{"TestSelect", "", `<select id="TestSelect" name="TestSelect">
<option value="foo">The Foo!</option>
<option value="bar">The Bar!</option>
</select>`},
		{"TestSelect2", "unknown!", `<select id="TestSelect2" name="TestSelect2">
<option value="foo">The Foo!</option>
<option value="bar">The Bar!</option>
</select>`},
		{"TestSelect3", "foo", `<select id="TestSelect3" name="TestSelect3">
<option value="foo" selected>The Foo!</option>
<option value="bar">The Bar!</option>
</select>`},
		{"TestSelect4", "bar", `<select id="TestSelect4" name="TestSelect4">
<option value="foo">The Foo!</option>
<option value="bar" selected>The Bar!</option>
</select>`}}
	for _, v := range tests {
		ret := widget.HTML(v.Name, v.Value)
		if string(ret) != v.Expected {
			t.Errorf(`SelectWidget.HTML("%v", "%v") = "%v", should be "%v".`,
				v.Name, v.Value, ret, v.Expected)
		}
	}
}

func TestHiddenWidget(t *testing.T) {
	widget := new(HiddenWidget)
	ret := widget.HTML("foo", "bar")
	expected := `<input id="foo" type="hidden" name="foo" value="bar"/>`
	if string(ret) != expected {
		t.Errorf(`HiddenWidget.HTML("Foo", "bar") = "%v", should be "%v".`,
			ret, expected)
	}
}

func TestFileWidget(t *testing.T) {
	widget := new(FileWidget)
	ret := widget.HTML("foo", "")
	expected := `<input id="foo" type="file" name="foo"/>`
	if string(ret) != expected {
		t.Errorf(`FileWidget.HTML("Foo", "") = "%v", should be "%v".`,
			ret, expected)
	}
}

func TestPasswordWidget(t *testing.T) {
	widget := new(PasswordWidget)
	ret := widget.HTML("foo", "")
	expected := `<input id="foo" type="password" name="foo"/>`
	if string(ret) != expected {
		t.Errorf(`PasswordWidget.HTML("Foo", "") = "%v", should be "%v".`,
			ret, expected)
	}
}

func testWidget(t *testing.T, widget Widget, data interface{}, input,
	nilInput string, value interface{}, urlValue string) {
	form := NewForm(data, []Field{Field{"ID", "T", "H", nil, widget}})
	vals := url.Values{"ID": []string{urlValue}}
	renderData := form.RenderData()
	if renderData.Fields[0].Input != template.HTML(nilInput) {
		t.Errorf("Input field for nil value is\n%v\nshould be \n%v",
			renderData.Fields[0].Input, nilInput)
	}
	form.Fill(vals)
	renderData = form.RenderData()
	if renderData.Fields[0].Input != template.HTML(input) {
		t.Errorf("Input field is\n%v\nshould be \n%v",
			renderData.Fields[0].Input, input)
	}
	if reflect.DeepEqual(reflect.ValueOf(value),
		reflect.ValueOf(data).Elem().FieldByName("ID")) {
		t.Errorf("Data is\n%v\nshould be \n%v", data, value)
	}
}

type TestDateTimeWidgetData struct {
	ID *time.Time
}

func TestDateTimeWidget(t *testing.T) {
	data := TestDateTimeWidgetData{}
	input := `<input id="ID" type="datetime" name="ID" value="2008-09-08T22:47:31-07:00"/>`
	nilInput := `<input id="ID" type="datetime" name="ID" value=""/>`
	value, err := time.Parse(time.RFC3339, "2008-09-08T22:47:31-07:00")
	if err != nil {
		t.Fatal(err)
	}
	testWidget(t, new(DateTimeWidget), &data, input, nilInput, value,
		"2008-09-08T22:47:31-07:00")
}

/*
func TestDateWidget(t *testing.T) {
	data := TestDateTimeWidgetData{}
	input := `<input id="ID" type="date" name="ID" value="2008-09-08"/>`
	nilInput := `<input id="ID" type="date" name="ID" value=""/>`
	value, err := time.Parse("2006-01-02", "2008-09-08")
	if err != nil {
		t.Fatal(err)
	}
	testWidget(t, new(DateWidget), &data, input, nilInput, value, "2008-09-08")
}

func TestTimeWidget(t *testing.T) {
	data := TestDateTimeWidgetData{}
	input := `<input id="ID" type="time" name="ID" value="22:47:31"/>`
	nilInput := `<input id="ID" type="time" name="ID" value=""/>`
	value, err := time.Parse("15:04:05", "22:47:31")
	if err != nil {
		t.Fatal(err)
	}
	testWidget(t, new(TimeWidget), &data, input, nilInput, value, "22:47:31")
}
*/

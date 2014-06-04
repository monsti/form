// This file is part of monsti/form.
// Copyright 2012 Christian Neumann

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

/*
Package form implements a web form generator and validator.

For each form, you have to define a struct for the form fields:
	type formData struct {
		Name string
		Age int
	}

Then you can use the form like this:
	func handle(req Request, res *Response) {
		data := formData{}
		form := form.NewForm(&data, form.Fields{
			"Name": form.Field{G("Name"), "Your Name", form.Required(G("Required.")), nil},
			"Age": form.Field{G("Age"), "Your Age",
			form.Required(G("Required.")), nil}})
		switch req.Method {
		case "GET":
			data.Name = "Default Name"
		case "POST":
			if form.Fill(req.GetFormData()) {
				save(data.Name, data.Age)
			}
		default:
			panic()
		}
		fmt.Fprint(res, renderTemplate(form.RenderData()))
	}

Fill the render data into a form template like this (html/template):
	<form action="{{.Action}}" method="POST" accept-charset="utf-8" {{.EncTypeAttr}}>
		<fieldset>
			<div class="control-group {{if .Errors}}error{{end}}">
				<div class="controls">
					<span class="help-block"
						>{{range .Errors}}{{.}}{{end}}</span>
				</div>
			</div>
			{{range .Fields}}
			<div class="control-group {{if .Errors}}error{{end}}">
				<label class="control-label" for="name">{{.Label}}</label>
				<div class="controls">{{.Input}}
					<span class="help-block">{{.Help}}
						{{range .Errors}}{{.}}{{end}}</span></div>
			</div>
			{{end}}
			<div class="control-group">
				<div class="controls">
					<button type="submit" class="btn btn-primary">{{G "Submit"}}</button>
				</div>
			</div>
		</fieldset>
	</form>
*/
package form

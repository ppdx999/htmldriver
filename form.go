package htmldriver

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Form represents an HTML form element
type Form struct {
	dom       *DOM
	selection *goquery.Selection
	values    url.Values
}

// Form finds and returns a form element matching the selector
func (d *DOM) Form(selector string) (*Form, error) {
	sel, err := d.findOne(selector)
	if err != nil {
		return nil, err
	}

	// Verify it's actually a form element
	if !sel.Is("form") {
		return nil, fmt.Errorf("selected element is not a form: %s", selector)
	}

	// Initialize form values with defaults from the HTML
	values := make(url.Values)
	sel.Find("input, select, textarea").Each(func(i int, s *goquery.Selection) {
		name, exists := s.Attr("name")
		if !exists || name == "" {
			return
		}

		inputType, _ := s.Attr("type")
		inputType = strings.ToLower(inputType)

		switch s.Get(0).Data {
		case "input":
			switch inputType {
			case "checkbox", "radio":
				// Only add if checked by default
				if _, checked := s.Attr("checked"); checked {
					value, _ := s.Attr("value")
					if value == "" {
						value = "on"
					}
					values.Add(name, value)
				}
			case "submit", "button", "image", "reset":
				// Don't include button values by default
			default:
				// text, password, hidden, email, etc.
				value, _ := s.Attr("value")
				values.Set(name, value)
			}
		case "select":
			// Find selected option
			s.Find("option[selected]").Each(func(j int, opt *goquery.Selection) {
				value, _ := opt.Attr("value")
				if value == "" {
					value = opt.Text()
				}
				values.Add(name, value)
			})
		case "textarea":
			values.Set(name, s.Text())
		}
	})

	return &Form{
		dom:       d,
		selection: sel,
		values:    values,
	}, nil
}

// Fill sets the value of a form field
func (f *Form) Fill(name, value string) (*Form, error) {
	// Check if field exists
	field := f.selection.Find(fmt.Sprintf("[name=\"%s\"]", name))
	if field.Length() == 0 {
		return f, fmt.Errorf("field not found: %s", name)
	}

	// Verify it's a valid input type for Fill
	inputType, _ := field.Attr("type")
	inputType = strings.ToLower(inputType)

	if field.Is("input") && (inputType == "checkbox" || inputType == "radio") {
		return f, fmt.Errorf("cannot use Fill for checkbox/radio field: %s (use CheckCheckbox or CheckRadio)", name)
	}

	f.values.Set(name, value)
	return f, nil
}

// MustFill is like Fill but panics on error
func (f *Form) MustFill(name, value string) *Form {
	form, err := f.Fill(name, value)
	if err != nil {
		panic(err)
	}
	return form
}

// CheckCheckbox checks a checkbox
func (f *Form) CheckCheckbox(name string) (*Form, error) {
	field := f.selection.Find(fmt.Sprintf("input[name=\"%s\"][type=\"checkbox\"]", name))
	if field.Length() == 0 {
		return f, fmt.Errorf("checkbox not found: %s", name)
	}

	value, exists := field.Attr("value")
	if !exists || value == "" {
		value = "on"
	}

	f.values.Set(name, value)
	return f, nil
}

// MustCheckCheckbox is like CheckCheckbox but panics on error
func (f *Form) MustCheckCheckbox(name string) *Form {
	form, err := f.CheckCheckbox(name)
	if err != nil {
		panic(err)
	}
	return form
}

// UncheckCheckbox unchecks a checkbox
func (f *Form) UncheckCheckbox(name string) (*Form, error) {
	field := f.selection.Find(fmt.Sprintf("input[name=\"%s\"][type=\"checkbox\"]", name))
	if field.Length() == 0 {
		return f, fmt.Errorf("checkbox not found: %s", name)
	}

	f.values.Del(name)
	return f, nil
}

// MustUncheckCheckbox is like UncheckCheckbox but panics on error
func (f *Form) MustUncheckCheckbox(name string) *Form {
	form, err := f.UncheckCheckbox(name)
	if err != nil {
		panic(err)
	}
	return form
}

// CheckRadio selects a radio button with the given value
func (f *Form) CheckRadio(name, value string) (*Form, error) {
	field := f.selection.Find(fmt.Sprintf("input[name=\"%s\"][type=\"radio\"][value=\"%s\"]", name, value))
	if field.Length() == 0 {
		return f, fmt.Errorf("radio button not found: %s with value %s", name, value)
	}

	f.values.Set(name, value)
	return f, nil
}

// MustCheckRadio is like CheckRadio but panics on error
func (f *Form) MustCheckRadio(name, value string) *Form {
	form, err := f.CheckRadio(name, value)
	if err != nil {
		panic(err)
	}
	return form
}

// Select selects an option in a select element
func (f *Form) Select(name, value string) (*Form, error) {
	selectEl := f.selection.Find(fmt.Sprintf("select[name=\"%s\"]", name))
	if selectEl.Length() == 0 {
		return f, fmt.Errorf("select element not found: %s", name)
	}

	// Verify the option exists
	option := selectEl.Find(fmt.Sprintf("option[value=\"%s\"]", value))
	if option.Length() == 0 {
		return f, fmt.Errorf("option not found: %s in select %s", value, name)
	}

	f.values.Set(name, value)
	return f, nil
}

// MustSelect is like Select but panics on error
func (f *Form) MustSelect(name, value string) *Form {
	form, err := f.Select(name, value)
	if err != nil {
		panic(err)
	}
	return form
}

// GetValue returns the current value of a form field
func (f *Form) GetValue(name string) (string, error) {
	if !f.HasField(name) {
		return "", fmt.Errorf("field not found: %s", name)
	}
	return f.values.Get(name), nil
}

// HasField returns true if the form has a field with the given name
func (f *Form) HasField(name string) bool {
	field := f.selection.Find(fmt.Sprintf("[name=\"%s\"]", name))
	return field.Length() > 0
}

// Find finds and returns a generic element within the form matching the selector
// Supports standard CSS selectors
func (f *Form) Find(selector string) (*Element, error) {
	sel := f.selection.Find(selector)

	if sel.Length() == 0 {
		return nil, fmt.Errorf("element not found in form: %s", selector)
	}

	if sel.Length() > 1 {
		return nil, fmt.Errorf("multiple elements found in form for selector: %s (found %d)", selector, sel.Length())
	}

	return &Element{selection: sel}, nil
}

// Submit submits the form and returns the response
func (f *Form) Submit() (Response, error) {
	// Get form action and method
	action, exists := f.selection.Attr("action")
	if !exists || action == "" {
		action = "/"
	}

	method, exists := f.selection.Attr("method")
	if !exists || method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	// Resolve URL
	actionURL, err := url.Parse(action)
	if err != nil {
		return Response{}, fmt.Errorf("invalid form action URL: %w", err)
	}

	// Make relative URLs absolute
	if !actionURL.IsAbs() {
		actionURL = f.dom.baseURL.ResolveReference(actionURL)
	}

	// Build request
	req := f.dom.buildRequest(method, actionURL)

	if method == http.MethodGet {
		// For GET, append form data to URL
		actionURL.RawQuery = f.values.Encode()
		req.URL = actionURL
	} else {
		// For POST, add to form data
		req.Form = f.values
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Send request
	resp, err := f.dom.transport.Do(req)
	if err != nil {
		return Response{}, err
	}

	// Update cookies from response
	f.dom.updateCookies(resp)

	return resp, nil
}

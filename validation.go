package crud

import (
	"reflect"
	"regexp"

	validator "github.com/go-phings/struct-validator"
)

// Validate checks object's fields. It returns result of validation as a bool and list of fields with invalid value
func (c *Controller) Validate(obj interface{}, filters map[string]interface{}) (bool, map[string]int, error) {
	if filters != nil {
		valid, failedFields := validator.Validate(obj, &validator.ValidationOptions{
			OverwriteTagName:     "crud",
			ValidateWhenSuffix:   true,
			OverwriteFieldValues: filters,
			RestrictFields:       c.mapWithInterfacesToMapBool(filters),
		})
		return valid, failedFields, nil
	}

	valid, failedFields := validator.Validate(obj, &validator.ValidationOptions{
		OverwriteTagName:   "crud",
		ValidateWhenSuffix: true,
	})
	return valid, failedFields, nil
}

// validateFieldRequired checks if field that is required has a value
func (c *Controller) validateFieldRequired(valueField reflect.Value, canBeZero bool) bool {
	if valueField.Type().Name() == "string" && valueField.String() == "" {
		return false
	}
	if valueField.Type().Name() == "int" && valueField.Int() == 0 && !canBeZero {
		return false
	}
	if valueField.Type().Name() == "int64" && valueField.Int() == 0 && !canBeZero {
		return false
	}
	return true
}

// validateFieldLength checks string field's length
func (c *Controller) validateFieldLength(valueField reflect.Value, length [2]int) bool {
	if valueField.Type().Name() != "string" {
		return true
	}
	if length[0] > -1 && len(valueField.String()) < length[0] {
		return false
	}
	if length[1] > -1 && len(valueField.String()) > length[1] {
		return false
	}
	return true
}

// validateFieldValue checks int field's value
func (c *Controller) validateFieldValue(valueField reflect.Value, value [2]int, minIsZero bool, maxIsZero bool) bool {
	if valueField.Type().Name() != "int" && valueField.Type().Name() != "int64" {
		return true
	}
	// Minimal value is 0 only when canBeZero is true; otherwise it's not defined
	if ((minIsZero && value[0] == 0) || value[0] != 0) && valueField.Int() < int64(value[0]) {
		return false
	}
	// Maximal value is 0 only when canBeZero is true; otherwise it's not defined
	if ((maxIsZero && value[1] == 0) || value[1] != 0) && valueField.Int() > int64(value[1]) {
		return false
	}
	return true
}

// validateFieldEmail checks if email field has a valid value
func (c *Controller) validateFieldEmail(valueField reflect.Value) bool {
	if valueField.Type().Name() != "string" {
		return true
	}
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return emailRegex.MatchString(valueField.String())
}

// validateFieldRegExp checks if string field's value matches the regular expression
func (c *Controller) validateFieldRegExp(valueField reflect.Value, re *regexp.Regexp) bool {
	if valueField.Type().Name() != "string" {
		return true
	}
	if !re.MatchString(valueField.String()) {
		return false
	}
	return true
}

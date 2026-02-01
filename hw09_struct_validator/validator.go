package hw09structvalidator

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotStruct      = errors.New("expected struct")
	ErrInvalidTag     = errors.New("invalid validate tag")
	ErrInvalidRule    = errors.New("invalid validation rule")
	ErrInvalidRegexp  = errors.New("invalid regexp")
	ErrLen            = errors.New("invalid length")
	ErrMin            = errors.New("value is less than min")
	ErrMax            = errors.New("value is greater than max")
	ErrIn             = errors.New("value is not in set")
	ErrRegexpMismatch = errors.New("value does not match regexp")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var b strings.Builder
	for i, e := range v {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(e.Field)
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

func Validate(v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	typ := value.Type()
	var errs ValidationErrors

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" { // private field
			continue
		}

		fieldValue := value.Field(i)
		if fieldErr := validateField(field.Name, fieldValue, field.Tag.Get("validate")); fieldErr.Err != nil {
			errs = append(errs, fieldErr)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateField(fieldName string, value reflect.Value, tag string) ValidationError {
	if tag == "" {
		return ValidationError{}
	}

	rules := strings.Split(tag, "|")
	kind := value.Kind()

	switch {
	case isStringKind(kind):
		return validateStringField(fieldName, value.String(), rules)
	case isIntKind(kind):
		return validateIntField(fieldName, int(value.Int()), rules)
	case kind == reflect.Slice:
		return validateSliceField(fieldName, value, rules)
	default:
		return ValidationError{Field: fieldName, Err: ErrInvalidRule}
	}
}

func validateStringField(fieldName, value string, rules []string) ValidationError {
	for _, rule := range rules {
		if ruleErr := validateStringRule(value, rule); ruleErr != nil {
			return ValidationError{Field: fieldName, Err: ruleErr}
		}
	}
	return ValidationError{}
}

func validateIntField(fieldName string, value int, rules []string) ValidationError {
	for _, rule := range rules {
		if ruleErr := validateIntRule(value, rule); ruleErr != nil {
			return ValidationError{Field: fieldName, Err: ruleErr}
		}
	}
	return ValidationError{}
}

func validateSliceField(fieldName string, slice reflect.Value, rules []string) ValidationError {
	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		kind := elem.Kind()

		var elemErr ValidationError
		switch {
		case isStringKind(kind):
			elemErr = validateStringField(fieldName+"["+strconv.Itoa(i)+"]", elem.String(), rules)
		case isIntKind(kind):
			elemErr = validateIntField(fieldName+"["+strconv.Itoa(i)+"]", int(elem.Int()), rules)
		default:
			return ValidationError{Field: fieldName, Err: ErrInvalidRule}
		}

		if elemErr.Err != nil {
			return elemErr
		}
	}
	return ValidationError{}
}

// Проверяет, является ли тип строкой.
func isStringKind(kind reflect.Kind) bool {
	return kind == reflect.String
}

// Проверяет, является ли тип целым числом.
func isIntKind(kind reflect.Kind) bool {
	return kind >= reflect.Int && kind <= reflect.Int64
}

func validateStringRule(value, rule string) error {
	name, param, err := splitRule(rule)
	if err != nil {
		return err
	}

	switch name {
	case "len":
		return validateLen(value, param)
	case "regexp":
		return validateRegexp(value, param)
	case "in":
		return validateInString(value, param)
	default:
		return ErrInvalidRule
	}
}

func validateIntRule(value int, rule string) error {
	name, param, err := splitRule(rule)
	if err != nil {
		return err
	}

	switch name {
	case "min":
		return validateMin(value, param)
	case "max":
		return validateMax(value, param)
	case "in":
		return validateInInt(value, param)
	default:
		return ErrInvalidRule
	}
}

func validateLen(value, param string) error {
	length, err := strconv.Atoi(param)
	if err != nil {
		return ErrInvalidRule
	}
	if len(value) != length {
		return ErrLen
	}
	return nil
}

func validateRegexp(value, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ErrInvalidRegexp
	}
	if !re.MatchString(value) {
		return ErrRegexpMismatch
	}
	return nil
}

func validateInString(value, param string) error {
	set := strings.Split(param, ",")
	for _, v := range set {
		if value == strings.TrimSpace(v) {
			return nil
		}
	}
	return ErrIn
}

func validateMin(value int, param string) error {
	minVal, err := strconv.Atoi(param)
	if err != nil {
		return ErrInvalidRule
	}
	if value < minVal {
		return ErrMin
	}
	return nil
}

func validateMax(value int, param string) error {
	maxVal, err := strconv.Atoi(param)
	if err != nil {
		return ErrInvalidRule
	}
	if value > maxVal {
		return ErrMax
	}
	return nil
}

func validateInInt(value int, param string) error {
	values := strings.Split(param, ",")
	for _, v := range values {
		v = strings.TrimSpace(v)
		i, err := strconv.Atoi(v)
		if err != nil {
			return ErrInvalidRule
		}
		if value == i {
			return nil
		}
	}
	return ErrIn
}

func splitRule(rule string) (string, string, error) {
	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidTag
	}
	return parts[0], strings.TrimSpace(parts[1]), nil
}

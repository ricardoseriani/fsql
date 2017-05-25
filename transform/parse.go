package transform

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ParseParams holds the params for a parse-modifier function.
type ParseParams struct {
	Attribute string
	Value     interface{}

	Name string
	Args []string
}

// Parse runs the associated modifier function for the provided parameters.
// Depending on the type of p.Value, we may recursively run this method
// on every element of the structure.
//
// We're using reflect _quite_ heavily for this, meaning it's kind of unsafe,
// it'd be great if we could find another solution while keeping it as
// abstract as it is.
func Parse(p *ParseParams) (val interface{}, err error) {
	kind := reflect.TypeOf(p.Value).Kind()
	// If we have a slice/array, recursively run Parse on each element.
	if kind == reflect.Slice || kind == reflect.Array {
		s := reflect.ValueOf(p.Value)
		for i := 0; i < s.Len(); i++ {
			p.Value = s.Index(i).Interface()
			if val, err = Parse(p); err != nil {
				return nil, err
			}
			s.Index(i).Set(reflect.ValueOf(val))
		}
		return s.Interface(), nil
	}

	// If we have a map, recursively run Parse on each KEY and create a new
	// map out of the return values.
	if kind == reflect.Map {
		result := reflect.MakeMap(reflect.TypeOf(p.Value))
		for _, key := range reflect.ValueOf(p.Value).MapKeys() {
			p.Value = key.Interface()
			if val, err = Parse(p); err != nil {
				return nil, err
			}
			result.SetMapIndex(reflect.ValueOf(val), reflect.ValueOf(true))
		}
		return result.Interface(), nil
	}

	switch strings.ToUpper(p.Name) {
	case "FORMAT":
		val, err = pFormat(p)
	case "UPPER":
		val, err = upper(p.Value.(string)), nil
	case "LOWER":
		val, err = lower(p.Value.(string)), nil
	}

	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, &ErrNotImplemented{p.Name, p.Attribute}
	}
	return val, nil
}

func pFormat(p *ParseParams) (val interface{}, err error) {
	switch p.Attribute {
	case "name":
		val, err = formatName(p.Args[0], p.Value.(string)), nil
	case "size":
		val, err = pFormatSize(p)
	case "time":
		val, err = pFormatTime(p)
	}

	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, &ErrUnsupportedFormat{p.Args[0], p.Attribute}
	}
	return val, nil
}

func pFormatSize(p *ParseParams) (interface{}, error) {
	size, err := strconv.ParseFloat(p.Value.(string), 64)
	if err != nil {
		return nil, err
	}

	switch strings.ToUpper(p.Args[0]) {
	case "B":
		size *= 1
	case "KB":
		size *= 1 << 10
	case "MB":
		size *= 1 << 20
	case "GB":
		size *= 1 << 20
	default:
		return nil, nil
	}

	return size, nil
}

func pFormatTime(p *ParseParams) (interface{}, error) {
	var t time.Time
	var err error

	switch strings.ToUpper(p.Args[0]) {
	case "ISO":
		t, err = time.Parse(time.RFC3339, p.Value.(string))
	case "UNIX":
		t, err = time.Parse(time.UnixDate, p.Value.(string))
	default:
		t, err = time.Parse("Jan 02 2006 15 04", p.Value.(string))
	}

	if err != nil {
		return nil, err
	}

	return t, nil
}

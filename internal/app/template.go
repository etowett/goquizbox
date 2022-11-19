package app

import (
	"fmt"
	"html/template"
	"time"
)

// TemplateFuncMap is the list of template functions.
var TemplateFuncMap = template.FuncMap{
	"deref":        deref,
	"htmlDate":     timestampFormatter("2006-01-02"),
	"htmlTime":     timestampFormatter("15:04"),
	"htmlDatetime": timestampFormatter(time.UnixDate),
}

// timestampFormatter returns a function that formats the given timestamp.
func timestampFormatter(f string) func(i interface{}) (string, error) {
	return func(i interface{}) (string, error) {
		switch t := i.(type) {
		case nil:
			return "", nil
		case time.Time:
			if t.IsZero() {
				return "", nil
			}
			return t.UTC().Format(f), nil
		case *time.Time:
			if t == nil || t.IsZero() {
				return "", nil
			}
			return t.UTC().Format(f), nil
		case string:
			return t, nil
		default:
			return "", fmt.Errorf("unknown type %v", t)
		}
	}
}

// deref dereferences a pointer into its concrete type.
func deref(i interface{}) (string, error) {
	switch t := i.(type) {
	case nil:
		return "", nil
	case *string:
		if t == nil {
			return "", nil
		}
		return *t, nil
	case *int:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *int8:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *int16:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *int32:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *int64:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *uint:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *uint8:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *uint16:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *uint32:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	case *uint64:
		if t == nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", *t), nil
	default:
		return "", fmt.Errorf("unknown type %T: %v", t, t)
	}
}

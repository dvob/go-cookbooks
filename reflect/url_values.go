package reflect

import (
	"encoding"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func ReadURLValuesIntoStruct(urlValues url.Values, target any) error {
	const tagName = "url"

	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Pointer && t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target is not a pointer to a struct")
	}

	t = t.Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tag, ok := field.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		urlKeyName, rawSettings, ok := strings.Cut(tag, ",")
		var settings []string
		if ok {
			settings = strings.Split(rawSettings, ",")
		}
		_ = settings

		values, ok := urlValues[urlKeyName]
		if !ok {
			continue
		}

		structValue := reflect.ValueOf(target).Elem()
		fieldValue := structValue.Field(i)

		// handle scalar values and values which support TextUnmarshaler
		read, err := readInto(fieldValue, values[0])
		if err != nil {
			return fmt.Errorf("failed to parse '%s': %w", urlKeyName, err)
		}

		if read {
			continue
		}

		// If value could not be read we check for slices
		if fieldValue.Type().Kind() == reflect.Slice {
			elementType := field.Type.Elem()
			newValues := []reflect.Value{}
			for i, value := range values {
				newValue := reflect.New(elementType)
				read, err := readInto(newValue, value)
				if err != nil {
					return fmt.Errorf("failed to parse '%s' (index=%d): %w", urlKeyName, i, err)
				}
				if !read {
					break
				}
				newValues = append(newValues, newValue.Elem())
			}
			fieldValue.Set(reflect.Append(fieldValue, newValues...))
		}
	}
	return nil
}

func readInto(fieldValue reflect.Value, str string) (bool, error) {
	if fieldValue.Type().Kind() == reflect.Pointer {
		fieldValue = fieldValue.Elem()
	}

	if u, ok := fieldValue.Addr().Interface().(encoding.TextUnmarshaler); ok {
		err := u.UnmarshalText([]byte(str))
		if err != nil {
			return false, err
		}
		return true, nil
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(str)
		return true, nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(str, fieldValue.Type().Bits())
		if err != nil {
			return false, err
		}
		fieldValue.SetFloat(f)
		return true, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(str, 10, fieldValue.Type().Bits())
		if err != nil {
			return false, err
		}
		fieldValue.SetInt(i)
		return true, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(str, 10, fieldValue.Type().Bits())
		if err != nil {
			return false, err
		}
		fieldValue.SetUint(u)
		return true, nil
	case reflect.Bool:
		b, err := strconv.ParseBool(str)
		if err != nil {
			return false, err
		}
		fieldValue.SetBool(b)
		return true, nil
	default:
		return false, nil
	}
}

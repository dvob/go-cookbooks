package main

import (
	"fmt"
	"slices"
	"strings"
)

type sliceValue struct {
	slice         *[]string
	itemSeperator string
}

// newSliceValue returns an new sliceValue which implements the flag.Value
// interface. If Set gets called the value is appended to slice. If the value
// is `-` the slice gets reset to an empty slice. If the value is `-name` the
// item `name` gets removed from the slice. If itemSeperator the values are
// first split with this seperator. This way you can define multiple values
// with one flag. This can be useful if you also read settings from the
// environment where you can't have the same environment variable multiple
// times.
func newSliceValue(slice *[]string, itemSeperator string) *sliceValue {
	if slice == nil {
		panic("slice can't be nil")
	}
	return &sliceValue{
		slice:         slice,
		itemSeperator: itemSeperator,
	}
}

// Set implements the flag.Value interface
func (s sliceValue) Set(value string) error {
	var elements []string

	if s.itemSeperator == "" {
		elements = []string{value}
	} else {
		elements = strings.Split(value, s.itemSeperator)
	}

	for _, element := range elements {
		// remove all elements
		if element == "-" {
			*s.slice = []string{}
			continue
		}

		// remove a single element
		if element != "" && element[0] == '-' {
			*s.slice = slices.DeleteFunc(*s.slice, func(cur string) bool { return cur == element[1:] })
			continue
		}

		*s.slice = append(*s.slice, element)
	}
	return nil
}

func (s sliceValue) String() string {
	if s.slice == nil {
		return ""
	}
	return strings.Join(*s.slice, ",")
}

// newMapValue returns an new mapValue which implements the flag.Value
// interface. Flags has to be in the format `key=value`. If Set gets called the
// key value pair gets inserted into the map. If Set gets called with `-` all
// keys are removed from the map. If Set gets called with `-key` the `key` get
// removed from the map. If itemSeperator is set the values are first split
// with this seperator. This way you can define multiple values with one flag.
// This can be useful if you also read settings from the environment where you
// can't have the same environment variable multiple times.
func newMapValue(strMap map[string]string, keyValueSeperator string, itemSeperator string) *mapValue {
	if strMap == nil {
		panic("strMap can't be nil")
	}
	return &mapValue{
		value:             strMap,
		itemSeperator:     itemSeperator,
		keyValueSeperator: keyValueSeperator,
	}
}

type mapValue struct {
	value             map[string]string
	keyValueSeperator string
	itemSeperator     string
}

// Set implements the flag.Value interface
func (m mapValue) Set(value string) error {
	var elements []string

	if m.itemSeperator == "" {
		elements = []string{value}
	} else {
		elements = strings.Split(value, m.itemSeperator)
	}

	for _, element := range elements {
		key, value, found := strings.Cut(element, m.keyValueSeperator)
		if !found {
			// reset entire map
			if key == "-" {
				clear(m.value)
				continue
			}

			// remove one key from map
			if key != "" && key[0] == '-' {
				delete(m.value, key[1:])
				continue
			}

			return fmt.Errorf("missing value for '%s': missing '%s'", key, m.keyValueSeperator)
		}

		m.value[key] = value
	}
	return nil
}

// String implements the flag.Value interface
func (m mapValue) String() string {
	str := strings.Builder{}
	if m.value == nil {
		return ""
	}
	first := true
	for key, value := range m.value {
		if !first {
			str.WriteString(m.itemSeperator)
		}
		str.WriteString(key + m.keyValueSeperator + value)
		first = false
	}
	return str.String()
}

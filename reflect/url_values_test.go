package reflect

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type MyCustomType []string

func (s *MyCustomType) UnmarshalText(text []byte) error {
	*s = strings.Split(string(text), " ")
	return nil
}

func TestReadURLValues(t *testing.T) {
	for _, test := range []struct {
		name           string
		urlValues      url.Values
		ExpectedTarget any
		expectedError  string
	}{
		{
			name: "scalar",
			urlValues: map[string][]string{
				"string": {"foo"},
				"int":    {"42"},
				"bool":   {"true"},
				"custom": {"one two"},
			},
			ExpectedTarget: struct {
				String string       `url:"string"`
				Int    int          `url:"int"`
				Bool   bool         `url:"bool"`
				Custom MyCustomType `url:"custom"`
			}{
				String: "foo",
				Int:    42,
				Bool:   true,
				Custom: []string{
					"one",
					"two",
				},
			},
		},
		{
			name: "slice",
			urlValues: map[string][]string{
				"string": {"foo", "bar"},
				"int":    {"42", "1337"},
				"bool":   {"true", "true", "false"},
				"custom": {"one two", "two one"},
			},
			ExpectedTarget: struct {
				String []string       `url:"string"`
				Int    []int          `url:"int"`
				Bool   []bool         `url:"bool"`
				Custom []MyCustomType `url:"custom"`
			}{
				String: []string{"foo", "bar"},
				Int:    []int{42, 1337},
				Bool:   []bool{true, true, false},
				Custom: []MyCustomType{
					{"one", "two"},
					{"two", "one"},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := reflect.New(reflect.TypeOf(test.ExpectedTarget)).Interface()
			err := ReadURLValuesIntoStruct(test.urlValues, value)
			if test.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error with text '%s', got no error", test.expectedError)
				}

				if !strings.Contains(err.Error(), test.expectedError) {
					t.Fatalf("expected error with text '%s', got '%s'", test.expectedError, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			// we need an interface which points to the struct
			// directly and not to a pointer of it.
			value = reflect.ValueOf(value).Elem().Interface()
			if !reflect.DeepEqual(value, test.ExpectedTarget) {
				want, err := json.MarshalIndent(&test.ExpectedTarget, "", "  ")
				if err != nil {
					panic(err)
				}
				got, err := json.MarshalIndent(value, "", "  ")
				if err != nil {
					panic(err)
				}
				t.Fatalf("got='%s' want='%s'", got, want)
			}
		})
	}

}

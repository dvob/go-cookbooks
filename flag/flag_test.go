package main

import (
	"reflect"
	"testing"
)

func TestSliceValue(t *testing.T) {
	for _, test := range []struct {
		name          string
		value         []string
		sep           string
		flags         []string
		expectedValue []string
	}{
		{
			name:          "add",
			value:         []string{},
			flags:         []string{"foo"},
			expectedValue: []string{"foo"},
		},
		{
			name:          "default",
			value:         []string{"default"},
			flags:         []string{"foo"},
			expectedValue: []string{"default", "foo"},
		},
		{
			name:          "add_multiple",
			value:         []string{"default"},
			flags:         []string{"foo", "bar"},
			expectedValue: []string{"default", "foo", "bar"},
		},
		{
			name:          "remove",
			value:         []string{"default"},
			flags:         []string{"foo", "bar", "-default"},
			expectedValue: []string{"foo", "bar"},
		},
		{
			name:          "reset",
			value:         []string{"default"},
			flags:         []string{"foo", "bar", "-", "new"},
			expectedValue: []string{"new"},
		},
		{
			name:          "with_item_seperator",
			value:         []string{"default"},
			sep:           ",",
			flags:         []string{"foo", "bar,bla"},
			expectedValue: []string{"foo", "bar", "bla"},
		},
	} {

		t.Run(test.name, func(t *testing.T) {
			v := newSliceValue(&test.value, test.sep)

			for _, flag := range test.flags {
				v.Set(flag)
			}

			if !reflect.DeepEqual(test.value, test.expectedValue) {
				t.Errorf("expected: %v, got: %v", test.expectedValue, test.value)
			}
		})
	}
}
func TestMapValue(t *testing.T) {
	for _, test := range []struct {
		name          string
		value         map[string]string
		keyValueSep   string
		itemSep       string
		flags         []string
		expectedValue map[string]string
	}{
		{
			name:        "add",
			value:       map[string]string{},
			keyValueSep: "=",
			flags:       []string{"foo=bar"},
			expectedValue: map[string]string{
				"foo": "bar",
			},
		},
		{
			name:        "add_multiple",
			value:       map[string]string{},
			keyValueSep: "=",
			flags:       []string{"foo=bar", "text=green, blue, yellow"},
			expectedValue: map[string]string{
				"foo":  "bar",
				"text": "green, blue, yellow",
			},
		},
		{
			name: "respect_default",
			value: map[string]string{
				"default": "123",
			},
			keyValueSep: "=",
			flags:       []string{"foo=bar", "abc=test"},
			expectedValue: map[string]string{
				"foo":     "bar",
				"abc":     "test",
				"default": "123",
			},
		},
		{
			name: "remove",
			value: map[string]string{
				"default": "123",
			},
			keyValueSep: "=",
			flags:       []string{"foo=bar", "abc=test", "-default"},
			expectedValue: map[string]string{
				"foo": "bar",
				"abc": "test",
			},
		},
		{
			name: "reset",
			value: map[string]string{
				"default": "123",
			},
			keyValueSep:   "=",
			flags:         []string{"foo=bar", "abc=test", "-"},
			expectedValue: map[string]string{},
		},
		{
			name: "customize_seperators",
			value: map[string]string{
				"default": "123",
			},
			keyValueSep: ":",
			itemSep:     ",",
			flags: []string{
				"foo:bar",
				"key1:value1,key2:value2",
			},
			expectedValue: map[string]string{
				"default": "123",
				"foo":     "bar",
				"key1":    "value1",
				"key2":    "value2",
			},
		},
	} {

		t.Run(test.name, func(t *testing.T) {
			v := newMapValue(test.value, test.keyValueSep, test.itemSep)

			for _, flag := range test.flags {
				v.Set(flag)
			}

			if !reflect.DeepEqual(test.value, test.expectedValue) {
				t.Errorf("expected: %v, got: %v", test.expectedValue, test.value)
			}
		})
	}
}

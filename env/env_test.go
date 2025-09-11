package env

import (
	"errors"
	"testing"
)

func TestGet(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expected string
		err      error
		envMap   map[string]string
	}{
		{
			name:     "success",
			key:      "key1",
			expected: "value1",
			envMap: map[string]string{
				"key1": "value1",
			},
		},
		{
			name:     "non-existing key",
			key:      "key2",
			expected: "",
			err:      ErrKeyNotFound,
			envMap: map[string]string{
				"key1": "value1",
			},
		},
		{
			name:     "empty value",
			key:      "key3",
			expected: "",
			err:      ErrValueIsEmpty,
			envMap: map[string]string{
				"key1": "value1",
				"key3": "",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New(tc.envMap)
			got, err := e.Get(tc.key)
			if err != nil {
				if !errors.Is(err, tc.err) {
					t.Errorf("expected error %v, got %v", tc.err, err)
				}
				return
			}
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

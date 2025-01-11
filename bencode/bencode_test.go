package bencode

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeBint64(t *testing.T) {

	tcs := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "test zero",
			input:    "i0e",
			expected: 0,
		},
		{
			name:     "test -1",
			input:    "i-1e",
			expected: -1,
		},
		{
			name:     "test 1",
			input:    "i1e",
			expected: 1,
		},
		{
			name:     "test 11",
			input:    "i11e",
			expected: 11,
		},
		{
			name:     "test -11",
			input:    "i-11e",
			expected: -11,
		},
		{
			name:     "test max int64 value",
			input:    fmt.Sprintf("i%ve", math.MaxInt64),
			expected: math.MaxInt64,
		},
		{
			name:     "test max int64 value",
			input:    fmt.Sprintf("i%ve", math.MaxInt64),
			expected: math.MaxInt64,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			value, _, err := DecodeBInt64([]byte(tc.input))
			if err != nil {
				t.Fatalf("got error in testcase: %v, e: %v", tc, err)
			}

			val, ok := value.(BInt64)
			if !ok {
				t.Fatalf("cannot convert bencode to bint64")
			}

			require.Equal(t, int64(val), tc.expected, "want: %v got: %v", tc.expected, int64(val))
		})
	}
}

func TestDecodeBString(t *testing.T) {

	tcs := []struct {
		name     string
		tcString string
	}{
		{"empty string", ""},
		{"test 1", "string1"},
		{"test number", "1241"},
		{"test name", "shreyas"},
		{"really long string", strings.Repeat("a", 256)},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			value, _, err := DecodeBString([]byte(fmt.Sprintf("%d:%s", len(tc.tcString), tc.tcString)))
			if err != nil {
				t.Fatalf("got error in testcase: %v, e: %v", tc, err)
			}

			val, ok := value.(BString)
			if !ok {
				t.Fatalf("cannot convert bencode to bstring")
			}

			if string(val) != tc.tcString {
				t.Errorf("want: %v got: %v", tc.tcString, value)
			}
		})
	}
}

func TestDecodeBList(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected BList
		err      error
	}{
		{
			name:     "Valid Bencoded list with integers",
			input:    []byte("l3:foo3:bare"),
			expected: BList{BString("foo"), BString("bar")},
			err:      nil,
		},
		{
			name:     "Empty Bencoded list",
			input:    []byte("le"),
			expected: BList{}, // Empty list
			err:      nil,
		},
		{
			name:     "Invalid Bencoded list - no closing 'e'",
			input:    []byte("l3:foo3:bar"),
			expected: nil,
			err:      fmt.Errorf("EOF while decoding Blist"),
		},
		{
			name:     "Single element Bencoded list",
			input:    []byte("l3:fooe"),
			expected: BList{BString("foo")}, // Single element
			err:      nil,
		},
		{
			name:     "Non-list input",
			input:    []byte("3:foo"),
			expected: nil,
			err:      fmt.Errorf("expected list but got something else"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, _, err := DecodeBList(tt.input)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestDecodeBMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected BMap
		err      error
	}{
		{
			name:     "Valid Bencoded map",
			input:    []byte("d3:foo3:bar3:baz3:quxe"),
			expected: BMap{BString("foo"): BString("bar"), BString("baz"): BString("qux")},
			err:      nil,
		},
		{
			name:     "Empty Bencoded map",
			input:    []byte("de"),
			expected: BMap{}, // Empty map
			err:      nil,
		},
		{
			name:     "Invalid start character (not a map)",
			input:    []byte("l3:foo3:bar"),
			expected: nil,
			err:      fmt.Errorf("expected dict found something else"),
		},
		{
			name:     "Non-BString key",
			input:    []byte("d3:foo3:bari123ee"),
			expected: nil,
			err:      fmt.Errorf("key not a BString"),
		},
		{
			name:     "Missing closing 'e'",
			input:    []byte("d3:foo3:bar"),
			expected: nil,
			err:      fmt.Errorf("EOF while decoding BMap"),
		},
		{
			name:     "Single key-value pair",
			input:    []byte("d3:fooi123ee"),
			expected: BMap{BString("foo"): BInt64(123)}, // Single key-value pair
			err:      nil,
		},
		{
			name:     "Multiple key-value pairs",
			input:    []byte("d3:fooi123e3:baz3:quxe"),
			expected: BMap{BString("foo"): BInt64(123), BString("baz"): BString("qux")},
			err:      nil,
		},
		{
			name:     "Multiple key-value pairs with non-string values",
			input:    []byte("d3:fooi123e3:bar3:quxe"),
			expected: BMap{BString("foo"): BInt64(123), BString("bar"): BString("qux")},
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, _, err := DecodeBMap(tt.input)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

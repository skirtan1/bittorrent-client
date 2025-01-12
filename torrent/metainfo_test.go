package torrent

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skirtan1/bittorrent/bencode"
	"github.com/stretchr/testify/require"
)

func TestDecodeFilesFromBencode(t *testing.T) {

	require := require.New(t)

	tcs := []struct {
		name     string
		input    string
		expected File
		err      error
	}{
		{
			name:     "valid value",
			input:    getBencStringForFile(t, 255, []string{"hello.txt"}),
			expected: File{255, filepath.Join("hello.txt")},
			err:      nil,
		},
		{
			name:     "multiple path values",
			input:    getBencStringForFile(t, 255, []string{"dir1", "hello.txt"}),
			expected: File{255, filepath.Join("dir1", "hello.txt")},
			err:      nil,
		},
		{
			name:  "no length key",
			input: fmt.Sprintf("d%d:%sl%d:%see", 4, "path", 9, "hello.txt"),
			err:   ErrKeyNotPresent,
		},
		{
			name:  "no path key",
			input: "d6:lengthi20ee",
			err:   ErrKeyNotPresent,
		},
		{
			name:  "zero path key",
			input: "d6:lengthi20e4:pathlee",
			err:   ErrZeroLengthFilePathList,
		},
		{
			name:  "bencode not a dictionary",
			input: "le",
			err:   ErrTypeAssertionFromBencode,
		},
	}

	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			benc, _, err := bencode.Decode([]byte(tt.input))
			require.Nil(err)

			file, err := DecodeFilesFromBencode(benc)
			if tt.err != nil {
				require.NotNil(err)
				require.True(errors.Is(err, tt.err))
			} else {
				require.Nil(err)
				require.Equal(*file, tt.expected)
			}

		})
	}
}

func TestDecodeFilesInfoFromBencode(t *testing.T) {

	require := require.New(t)
	tcs := []struct {
		name     string
		input    string
		expected File
		err      error
	}{
		{
			name: "valid value",
			input: getBencFilelist(t, []string{
				getBencStringForFile(t, 255, []string{"hello.txt"}),
				getBencStringForFile(t, 255, []string{"hello.txt"})}),
			expected: File{255, filepath.Join("hello.txt")},
			err:      nil,
		},
		{
			name: "multiple path values",
			input: getBencFilelist(t, []string{
				getBencStringForFile(t, 255, []string{"dir1", "hello.txt"}),
				getBencStringForFile(t, 255, []string{"dir1", "hello.txt"})}),
			expected: File{255, filepath.Join("dir1", "hello.txt")},
			err:      nil,
		},
		{
			name:  "no length key",
			input: getBencFilelist(t, []string{fmt.Sprintf("d%d:%sl%d:%see", 4, "path", 9, "hello.txt")}),
			err:   ErrKeyNotPresent,
		},
		{
			name:  "no path key",
			input: getBencFilelist(t, []string{"d6:lengthi20ee"}),
			err:   ErrKeyNotPresent,
		},
		{
			name:  "zero path key",
			input: getBencFilelist(t, []string{"d6:lengthi20e4:pathlee"}),
			err:   ErrZeroLengthFilePathList,
		},
		{
			name:  "empty files info",
			input: "le",
			err:   ErrEmptyFilesInfo,
		},
		{
			name:  "files info is not a list",
			input: "de",
			err:   ErrTypeAssertionFromBencode,
		},
	}

	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			benc, _, err := bencode.Decode([]byte(tt.input))
			require.Nil(err)

			fileinfo, err := DecodeFilesInfoFromBencode(benc)
			if tt.err != nil {
				require.NotNil(err)
				require.True(errors.Is(err, tt.err))
			} else {
				require.Nil(err)

				for _, val := range fileinfo {
					require.Equal(*val, tt.expected)
				}
			}

		})
	}
}

func getBencStringForFile(t *testing.T, length int64, filepath []string) string {
	t.Helper()

	ret := strings.Builder{}

	ret.WriteString(fmt.Sprintf("d%d:%si%de%d:%sl", len("length"), "length",
		length, len("path"), "path"))

	for _, v := range filepath {
		ret.WriteString(fmt.Sprintf("%d:%s", len(v), v))
	}

	ret.WriteString("ee")

	return ret.String()
}

func getBencFilelist(t *testing.T, arr []string) string {
	t.Helper()

	ret := strings.Builder{}

	ret.WriteByte('l')
	for _, val := range arr {
		ret.WriteString(val)
	}
	ret.WriteByte('e')

	return ret.String()
}

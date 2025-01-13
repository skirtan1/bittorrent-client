package torrent

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skirtan1/bittorrent-client/bencode"
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

func TestDecodeInfoFromBencode(t *testing.T) {

	tests := []struct {
		name         string
		bencodeInput bencode.Bencode
		expectedInfo *Info
		err          error
	}{
		{
			name: "Success",
			bencodeInput: bencode.BMap{
				bencode.BString("name"):         bencode.BString("temp"),
				bencode.BString("piece length"): bencode.BInt64(212314),
				bencode.BString("pieces"):       bencode.BString(strings.Repeat("a", 40)),
				bencode.BString("length"):       bencode.BInt64(212314 * 2),
			},
			expectedInfo: &Info{
				Name:        "temp",
				PieceLength: 212314,
				Pieces: [][20]byte{{'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
					'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a'}, {'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
					'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a'}},
				Length: 212314 * 2,
			},
		},
		{
			name: "Missing piece length",
			bencodeInput: bencode.BMap{
				bencode.BString("pieces"): bencode.BString("abc123def456..."),
			},
			err: ErrKeyNotPresent,
		},
		{
			name: "Missing pieces",
			bencodeInput: bencode.BMap{
				bencode.BString("piece length"): bencode.BInt64(262144),
			},
			err: ErrKeyNotPresent,
		},
		{
			name: "Invalid pieces length",
			bencodeInput: bencode.BMap{
				bencode.BString("name"):         bencode.BString("temp"),
				bencode.BString("piece length"): bencode.BInt64(262144),
				bencode.BString("pieces"):       bencode.BString("abc123"),
			},
			err: ErrPieceNotCorrentLen,
		},
		{
			name: "Multi-file case",
			bencodeInput: bencode.BMap{
				bencode.BString("name"):         bencode.BString("temp"),
				bencode.BString("piece length"): bencode.BInt64(262144),
				bencode.BString("pieces"):       bencode.BString(strings.Repeat("a", 40)),
				bencode.BString("files"): bencode.BList{
					bencode.BMap{
						bencode.BString("path"):   bencode.BList{bencode.BString("file1.txt")},
						bencode.BString("length"): bencode.BInt64(1000),
					},
					bencode.BMap{
						bencode.BString("path"):   bencode.BList{bencode.BString("file2.txt")},
						bencode.BString("length"): bencode.BInt64(2000),
					},
				},
			},
			expectedInfo: &Info{
				Name:        "temp",
				PieceLength: 262144,
				Pieces: [][20]byte{{'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
					'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a'}, {'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
					'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a'}},
				FilesInfo: []*File{{Path: "file1.txt", Length: 1000}, {Path: "file2.txt", Length: 2000}},
			},
		},
		{
			name: "Error decoding file info",
			bencodeInput: bencode.BMap{
				bencode.BString("name"):         bencode.BString("temp"),
				bencode.BString("piece length"): bencode.BInt64(262144),
				bencode.BString("pieces"):       bencode.BString("abc123def456..."),
				bencode.BString("files"):        bencode.BList{},
			},
			err: ErrPieceNotCorrentLen,
		},
		{
			name:         "Invalid Bencode type",
			bencodeInput: bencode.BString("invalid"),
			err:          ErrTypeAssertionFromBencode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeInfoFromBencode(tt.bencodeInput)

			// Check if error is expected
			if tt.err != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.err))
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)

				copy(tt.expectedInfo.InfoHash[:], result.InfoHash[:])
				require.NoError(t, err)
				require.Equal(t, tt.expectedInfo, result)
			}
		})
	}
}

func TestDecodeMetaInfo(t *testing.T) {

	info := bencode.BMap{
		bencode.BString("name"):         bencode.BString("temp"),
		bencode.BString("piece length"): bencode.BInt64(212314),
		bencode.BString("pieces"):       bencode.BString(strings.Repeat("a", 40)),
		bencode.BString("length"):       bencode.BInt64(212314 * 2),
	}

	infoStruct := func(t *testing.T) *Info {
		t.Helper()

		val, err := DecodeInfoFromBencode(info)
		if err != nil {
			t.Error(err)
		}

		return val
	}(t)

	func(i *Info) {
		t.Helper()
		enc, err := bencode.Encode(info)
		if err != nil {
			t.FailNow()
			return
		}

		i.InfoHash = sha1.Sum(enc)
	}(infoStruct)

	tests := []struct {
		name         string
		bencodeInput bencode.Bencode
		expectedMeta *Metainfo
		err          error
	}{
		{
			name: "valid metainfo",
			bencodeInput: bencode.BMap{
				bencode.BString("announce"): bencode.BString("here i come"),
				bencode.BString("info"):     info,
			},
			expectedMeta: &Metainfo{
				Announce: "here i come",
				Info:     *infoStruct,
			},
		},
		{
			name:         "not a bmap metainfo",
			bencodeInput: bencode.BList{},
			err:          ErrTypeAssertionFromBencode,
		},
		{
			name: "doesn't have announce metainfo",
			bencodeInput: bencode.BMap{
				bencode.BString("not announce"): bencode.BString("here i come"),
				bencode.BString("info"):         info,
			},
			err: ErrKeyNotPresent,
		},
		{
			name: "doesn't have info",
			bencodeInput: bencode.BMap{
				bencode.BString("announce"): bencode.BString("here i come"),
			},
			err: ErrKeyNotPresent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := DecodeMetaInfoFromBencode(tt.bencodeInput)

			if tt.err != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.err))
			} else {
				require.Nil(t, err)
				require.Equal(t, val, tt.expectedMeta)
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

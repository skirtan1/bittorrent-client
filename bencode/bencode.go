package bencode

import (
	"fmt"
	"strconv"
)

type Bencode any

type BInt64 int64
type BString string
type BList []Bencode
type BMap map[BString]Bencode

func Decode(d []byte) (Bencode, int, error) {

	switch {
	case d[0] == 'i':
		value, idx, err := DecodeBInt64(d)
		if err != nil {
			return nil, 0, err
		}

		return value, idx, nil
	case d[0] >= '0' && d[0] <= '9':
		value, idx, err := DecodeBString(d)
		if err != nil {
			return nil, 0, err
		}

		return value, idx, nil
	case d[0] == 'l':
		value, idx, err := DecodeBList(d)
		if err != nil {
			return nil, 0, err
		}

		return value, idx, err
	case d[0] == 'd':
		value, idx, err := DecodeBMap(d)
		if err != nil {
			return nil, 0, err
		}

		return value, idx, err
	default:
		return nil, 0, fmt.Errorf("invalid first token: %c while decoding", d[0])
	}
}

func DecodeBInt64(d []byte) (Bencode, int, error) {
	idx := 1

	if len(d) < 3 {
		return nil, 0, fmt.Errorf("shortest bint64 is of len 3, buffer len: %v", len(d))
	}

	for ; idx < len(d) && d[idx] != 'e'; idx += 1 {
	}
	if idx == len(d) {
		return nil, 0, fmt.Errorf("EOF while decoding int")
	}

	value, err := strconv.Atoi(string(d[1:idx]))
	if err != nil {
		return nil, 0, err
	}

	idx += 1
	return BInt64(value), idx, nil
}

func DecodeBString(d []byte) (Bencode, int, error) {
	idx := 0

	for ; idx < len(d) && d[idx] != ':'; idx += 1 {
	}

	if idx == len(d) && d[idx] != ':' {
		return nil, 0, fmt.Errorf("EOF while decoding string")
	}

	strLen, err := strconv.Atoi(string(d[:idx]))
	if err != nil {
		return nil, 0, fmt.Errorf("invalid string len while decoding string")
	}

	if len(d) < (idx + strLen + 1) {
		return nil, 0, fmt.Errorf("string exceeds bufferlen")
	}

	return BString(d[idx+1 : idx+strLen+1]), idx + 1 + strLen, nil
}

func DecodeBList(d []byte) (Bencode, int, error) {
	if d[0] != 'l' {
		return nil, 0, fmt.Errorf("expected list but got something else")
	}
	idx := 1
	ret := make([]Bencode, 0)
	for idx < len(d) && d[idx] != 'e' {
		value, incr, err := Decode(d[idx:])
		if err != nil {
			return nil, 0, err
		}

		ret = append(ret, value)
		idx += incr
	}

	if idx == len(d) || d[idx] != 'e' {
		return nil, 0, fmt.Errorf("EOF while decoding Blist")
	}

	return BList(ret), idx, nil
}

func DecodeBMap(d []byte) (Bencode, int, error) {
	if d[0] != 'd' {
		return nil, 0, fmt.Errorf("expected dict found something else")
	}

	idx := 1
	ret := make(map[BString]Bencode)

	for idx < len(d) && d[idx] != 'e' {
		value, incr, err := Decode(d[idx:])
		if err != nil {
			return nil, 0, err
		}

		key, ok := value.(BString)
		if !ok {
			return nil, 0, fmt.Errorf("key not a BString")
		}

		idx += incr
		value, incr, err = Decode(d[idx:])
		if err != nil {
			return nil, 0, err
		}

		ret[BString(string(key))] = value
		idx += incr

	}

	if idx == len(d) {
		return nil, 0, fmt.Errorf("EOF while decoding BMap")
	}

	return BMap(ret), idx, nil
}

package torrent

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/skirtan1/bittorrent-client/bencode"
)

type Metainfo struct {
	Announce string
	Info     Info
}

type File struct {
	Length int64
	Path   string
}

type Info struct {
	Name        string
	PieceLength int64
	Pieces      [][20]byte
	Length      int64
	FilesInfo   []*File
	InfoHash    [20]byte
}

var (
	ErrTypeAssertionFromBencode = errors.New("cannot convert to expected B type from Bencode")
	ErrKeyNotPresent            = errors.New("key not present in bmap")
	ErrZeroLengthFilePathList   = errors.New("path in files.path is of zero length")
	ErrNeitherLengthOrFile      = errors.New("neither length or file present in info dict")
	ErrPieceNotCorrentLen       = errors.New("pieces should be a multiple of 20")
	ErrEmptyFilesInfo           = errors.New("files info should not be empty")
)

func DecodeFilesFromBencode(b bencode.Bencode) (*File, error) {
	value, ok := b.(bencode.BMap)

	if !ok {
		err := fmt.Errorf("unable to construct bmap from bencode: %w", ErrTypeAssertionFromBencode)
		slog.Error("decode file info error", "err", err)
		return nil, err
	}

	ret := File{}

	length, ok := value[bencode.BString("length")]
	if !ok {
		err := fmt.Errorf("cannot get length key in file dict: %w", ErrKeyNotPresent)
		slog.Error("decode file info error", "err", err)
		return nil, err
	}

	ret.Length = int64(length.(bencode.BInt64))
	list, ok := value[bencode.BString("path")]
	if !ok {
		err := fmt.Errorf("cannot get path key in file dict: %w", ErrKeyNotPresent)
		slog.Error("decode file info error", "err", err)
		return nil, err
	}

	pathlist := list.(bencode.BList)
	if len(pathlist) == 0 {
		slog.Error("decode file info error", "err", ErrZeroLengthFilePathList)
		return nil, ErrZeroLengthFilePathList
	}

	path := make([]string, 0)
	for _, value := range pathlist {
		val := value.(bencode.BString)
		path = append(path, string(val))
	}

	ret.Path = filepath.Join(path...)
	return &ret, nil
}

func DecodeFilesInfoFromBencode(b bencode.Bencode) ([]*File, error) {
	value, ok := b.(bencode.BList)
	if !ok {
		return nil, fmt.Errorf("decode files info, not a list: %w", ErrTypeAssertionFromBencode)
	}

	if len(value) == 0 {
		return nil, ErrEmptyFilesInfo
	}

	ret := make([]*File, 0)
	for _, v := range value {
		val := v.(bencode.BMap)

		finfo, err := DecodeFilesFromBencode(val)
		if err != nil {
			return nil, fmt.Errorf("decode file info error: %w", err)
		}
		ret = append(ret, finfo)
	}
	return ret, nil
}

func DecodeInfoFromBencode(b bencode.Bencode) (*Info, error) {
	value, ok := b.(bencode.BMap)

	ret := Info{}
	if !ok {
		err := fmt.Errorf("unable to construct bmap from bencode: %w", ErrTypeAssertionFromBencode)
		slog.Error("decode info error", "err", err)
		return nil, err
	}

	name, ok := value[bencode.BString("name")]
	if !ok {
		err := fmt.Errorf("unable to get name from info bencode: %w", ErrKeyNotPresent)
		slog.Error("decode info error", "err", err)
		return nil, err
	}

	ret.Name = string(name.(bencode.BString))

	pieceslength, ok := value[bencode.BString("piece length")]
	if !ok {
		err := fmt.Errorf("unable to get piece length from info bencode: %w", ErrKeyNotPresent)
		slog.Error("decode info error", "err", err)
		return nil, err
	}

	ret.PieceLength = int64(pieceslength.(bencode.BInt64))

	pieces, ok := value[bencode.BString("pieces")]
	if !ok {
		err := fmt.Errorf("unable to get pieces from info bencode: %w", ErrKeyNotPresent)
		slog.Error("decode info error", "err", err)
		return nil, err
	}

	if len(pieces.(bencode.BString))%20 != 0 {
		slog.Error("decode info error", "err", ErrPieceNotCorrentLen)
		return nil, ErrPieceNotCorrentLen
	}

	picesBytes := []byte(pieces.(bencode.BString))
	var temp [20]byte
	for i := 0; i < len(picesBytes); i += 20 {
		copy(temp[:], picesBytes[i:i+20])
		ret.Pieces = append(ret.Pieces, temp)
	}

	length, ok := value[bencode.BString("length")]
	if !ok {
		slog.Debug("decode info: length key not present, multi file")

		fileInfo, ok := value[bencode.BString("files")]
		if !ok {
			slog.Error("decode info error", "err", ErrNeitherLengthOrFile)
			return nil, ErrNeitherLengthOrFile
		}

		inf, err := DecodeFilesInfoFromBencode(fileInfo)
		if err != nil {
			return nil, fmt.Errorf("decode info error: %w", err)
		}
		ret.FilesInfo = inf
	} else {
		ret.Length = int64(length.(bencode.BInt64))
	}

	enc, err := bencode.Encode(b)
	if err != nil {
		return nil, fmt.Errorf("decode info err, cannot encode a bencode value")
	}

	ret.InfoHash = sha1.Sum(enc)
	return &ret, nil
}

func DecodeMetaInfoFromBencode(b bencode.Bencode) (*Metainfo, error) {
	value, ok := b.(bencode.BMap)

	ret := Metainfo{}
	if !ok {
		err := fmt.Errorf("unable to construct bmap from bencode: %w", ErrTypeAssertionFromBencode)
		slog.Error("decode metainfo error", "err", err)
		return nil, err
	}

	announce, ok := value[bencode.BString("announce")]
	if !ok {
		err := fmt.Errorf("announce not present in metainfo: %w", ErrKeyNotPresent)
		slog.Error("decode metainfo error", "err", err)
		return nil, err
	}

	ret.Announce = string(announce.(bencode.BString))

	infobencode, ok := value[bencode.BString("info")]
	if !ok {
		err := fmt.Errorf("info dict not present in metainfo: %w", ErrKeyNotPresent)
		slog.Error("decode metainfo error", "err", err)
		return nil, err
	}

	info, err := DecodeInfoFromBencode(infobencode)
	if err != nil {
		return nil, fmt.Errorf("decode metainfo error: %w", err)
	}

	ret.Info = *info
	return &ret, nil
}

func GetMetaInfoFromTorrentFile(torrentFilePath string) (*Metainfo, error) {

	data, err := os.ReadFile(torrentFilePath)
	if err != nil {
		return nil, fmt.Errorf("error building metainfo from torrentfile: %w", err)
	}

	benc, _, err := bencode.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("error decoding bencode from torrent file: %w", err)
	}

	minfo, err := DecodeMetaInfoFromBencode(benc)
	if err != nil {
		return nil, fmt.Errorf("error geting metainfo from benc: %w", err)
	}

	return minfo, nil
}

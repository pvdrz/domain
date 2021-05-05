package doc

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
)

type Document struct {
	Title     string
	Authors   []string
	Keywords  []string
	Extension string
	Hash      Hash
}

type Hash [32]byte

func (hash Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(hash[:]))
}

func (hash *Hash) UnmarshalJSON(data []byte) error {
	var str string

	err := json.Unmarshal(data, &str)

	if err != nil {
		return err
	}

	bytes, err := hex.DecodeString(str)

	if err != nil {
		return err
	}

	if len(bytes) != 32 {
		return errors.New("Hash of document does not have length 32")
	}

	copy(hash[0:32], bytes)

	return nil
}

type DocumentID [8]byte

func (id DocumentID) ToString() string {
	index := binary.BigEndian.Uint64(id[:])
	return strconv.FormatUint(index, 10)
}

func DocumentIDFromString(str string) (DocumentID, error) {
	var id DocumentID

	index, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return id, err
	}

	binary.BigEndian.PutUint64(id[:], index)

	return id, nil
}
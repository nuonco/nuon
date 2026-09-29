package shortid

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	base           = 36
	intBytes       = 8
	shortIDLen     = 26
	uuidLen        = 36
	uuidBytes      = 64
	nanoIDAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	nanoIDLen      = 23
)

func ToUUIDs(strs ...string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, len(strs))

	for idx, s := range strs {
		id, err := ToUUID(s)
		if err != nil {
			return nil, fmt.Errorf("unable to parse UUID: %v: %w", idx, err)
		}
		ids[idx] = id
	}

	return ids, nil
}

func ToUUID(s string) (uuid.UUID, error) {
	var u uuid.UUID
	if len(s) < shortIDLen {
		return u, fmt.Errorf("invalid shortid length")
	}

	middle := shortIDLen / 2
	i1, err := strconv.ParseUint(s[:middle], base, uuidBytes)
	if err != nil {
		return u, err
	}

	i2, err := strconv.ParseUint(s[middle:], base, uuidBytes)
	if err != nil {
		return u, err
	}

	bs := [16]byte{}
	binary.BigEndian.PutUint64(bs[:intBytes], i1)
	binary.BigEndian.PutUint64(bs[intBytes:], i2)

	u = uuid.UUID(bs)

	return u, nil
}

func ParseUUID(u uuid.UUID) string {
	msi := binary.BigEndian.Uint64(u[:intBytes])
	lsi := binary.BigEndian.Uint64(u[intBytes:])

	mss := strconv.FormatUint(msi, base)
	lss := strconv.FormatUint(lsi, base)

	return fmt.Sprintf("%026s", mss+lss)
}

func ParseString(s string) (string, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return "", err
	}
	return ParseUUID(u), nil
}

func New() string {
	return ParseUUID(uuid.New())
}

func NewNanoID(prefix string) string {
	id, err := gonanoid.Generate(nanoIDAlphabet, nanoIDLen)
	if err != nil {
		panic(err)
	}
	if prefix != "" {
		return prefix + id
	}
	return "def" + id
}

func ParseStrings(strs ...string) ([]string, error) {
	ids := make([]string, len(strs))

	for idx, s := range strs {
		id, err := ParseString(s)
		if err != nil {
			return nil, fmt.Errorf("unable to parse short id: %v: %w", idx, err)
		}
		ids[idx] = id
	}

	return ids, nil
}

func IsShortID(val string) bool {
	return len(val) == shortIDLen
}

func ToShortID(s string) (string, error) {
	switch len(s) {
	case 0:
		return "", fmt.Errorf("empty string is not a valid shortid")
	case shortIDLen:
		uu, err := ToUUID(s)
		if err != nil {
			return "", fmt.Errorf("invalid shortid: %w", err)
		}
		return ParseUUID(uu), nil
	case uuidLen:
		return ParseString(s)
	default:
		return "", fmt.Errorf("id incorrect length")
	}
}

func RegisterFakes() {
}

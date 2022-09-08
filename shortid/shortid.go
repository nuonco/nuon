package shortid

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

const (
	base       = 36
	intBytes   = 8
	shortIDLen = 26
	uuidBytes  = 64
)

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

func ParseUUID(u uuid.UUID) (string, error) {
	msi := binary.BigEndian.Uint64(u[:intBytes])
	lsi := binary.BigEndian.Uint64(u[intBytes:])

	mss := strconv.FormatUint(msi, base)
	lss := strconv.FormatUint(lsi, base)

	return fmt.Sprintf("%026s", mss+lss), nil
}

func ParseString(s string) (string, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return "", err
	}
	return ParseUUID(u)
}

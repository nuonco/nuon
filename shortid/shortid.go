package shortid

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

const (
	base     = 36
	intBytes = 8
)

func ToUUID(s string) (uuid.UUID, error) {
	var u uuid.UUID
	if len(s) < 26 {
		return u, fmt.Errorf("invalid shortid length")
	}

	i1, err := strconv.ParseUint(s[:13], base, 64)
	if err != nil {
		return u, err
	}

	i2, err := strconv.ParseUint(s[13:], base, 64)
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

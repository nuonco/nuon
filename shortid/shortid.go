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

// ToUUIDs parses a list of string shortids into a list of uuids, failing all if any fail
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

// ToUUID converts a short id string back into a uuid
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

// ParseUUID parses a uuid into a short id string
func ParseUUID(u uuid.UUID) (string, error) {
	msi := binary.BigEndian.Uint64(u[:intBytes])
	lsi := binary.BigEndian.Uint64(u[intBytes:])

	mss := strconv.FormatUint(msi, base)
	lss := strconv.FormatUint(lsi, base)

	return fmt.Sprintf("%026s", mss+lss), nil
}

// ParseString parses a string uuid into a short id, returning an error if invalid
func ParseString(s string) (string, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return "", err
	}
	return ParseUUID(u)
}

// ParseStrings parses a list of string UUIDs into a list of shortids, failing all if any fail
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

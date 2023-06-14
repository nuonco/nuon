package shortid

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var validTests = []struct {
	s string
	u uuid.UUID
}{
	{s: "00000000000000000000000000", u: uuid.MustParse("00000000-0000-0000-0000-000000000000")},
	{s: "3ocwtlzqh0qhp21818cc5qls0i", u: uuid.MustParse("f1c4cad2-3160-4f1d-85c0-3283abfacf12")},
	{s: "255qopka2e87j228kqxug3uzex", u: uuid.MustParse("8cf0deea-1d39-4b0f-879a-da78638151c9")},
	{s: "2rc870xu5npl72ve4gkl2i2qjs", u: uuid.MustParse("b574a30a-49b8-4e4b-bcdb-b98ac24ddce8")},
	{s: "23qx3xpx8k3n12ko2yk6axe6wm", u: uuid.MustParse("8a5cbc19-4de0-4b6d-a945-51dcd4a3a376")},
	{s: "38bwg68ukixfx2vvoylfk3icwb", u: uuid.MustParse("d47dd735-d69b-45cd-bdbf-f084098df4cb")},
	{s: "0kw13a6s7wewn2t9ycdrjrp93g", u: uuid.MustParse("262855da-4ec4-4a77-b8fe-55126654114c")},
	{s: "2k6c10zhkkkzq203st5m1vhy5d", u: uuid.MustParse("a85ec8e2-a6c4-4f46-83b5-965547018a21")},
	{s: "19xb6yg8dbwrf25szh6x39ck2y", u: uuid.MustParse("53e35346-8c02-479b-8e1e-cbd81cd315aa")},
	{s: "06c3i83wve75i22toww4ps214l", u: uuid.MustParse("0b92d266-b6e6-4ab6-88ad-219b687d5235")},
	{s: "1qk5lbudc126627ndn8hiaencp", u: uuid.MustParse("7245e2b9-12c8-413e-917d-33733e674fb9")},
	{s: "3w5e11264sgsf3w5e11264sgsf", u: uuid.MustParse("FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF")},
}

func TestParseString(t *testing.T) {
	for _, test := range validTests {
		t.Run(test.s, func(t *testing.T) {
			s, err := ParseString(test.u.String())
			assert.NoError(t, err)
			assert.Equal(t, test.s, s)
		})
	}
}

func TestParseUUID(t *testing.T) {
	for _, test := range validTests {
		t.Run(test.s, func(t *testing.T) {
			s := ParseUUID(test.u)
			assert.Equal(t, test.s, s)
		})
	}
}

func TestUUID(t *testing.T) {
	for _, test := range validTests {
		t.Run(test.s, func(t *testing.T) {
			u, err := ToUUID(test.s)
			assert.NoError(t, err)
			assert.Equal(t, test.u, u)

			s := ParseUUID(u)
			assert.Equal(t, test.s, s)
		})
	}
}

func TestParseString_error(t *testing.T) {
	var errorTests = []struct {
		s   string
		err string
	}{
		{s: "0", err: "invalid UUID length"},
		{s: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", err: "invalid UUID format"},
		{s: "z----zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", err: "invalid UUID format"},
		{s: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", err: "invalid UUID length"},
	}
	for _, test := range errorTests {
		t.Run(test.s, func(t *testing.T) {
			_, err := ParseString(test.s)
			assert.ErrorContains(t, err, test.err)
		})
	}
}

func TestUUID_error(t *testing.T) {
	var errorTests = []struct {
		s   string
		err string
	}{
		{s: "0", err: "invalid shortid length"},
		{s: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", err: "value out of range"},
		{s: "z----zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", err: "invalid syntax"},
		{s: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", err: "value out of range"},
	}
	for _, test := range errorTests {
		t.Run(test.s, func(t *testing.T) {
			_, err := ToUUID(test.s)
			assert.ErrorContains(t, err, test.err)
		})
	}
}

func Test_ParseStrings(t *testing.T) {
	tests := map[string]struct {
		idsFn       func() []string
		assertFn    func(*testing.T, []string)
		errExpected error
	}{
		"happy path": {
			idsFn: func() []string {
				ids := make([]string, len(validTests))
				for idx, testID := range validTests {
					ids[idx] = testID.u.String()
				}
				return ids
			},
			assertFn: func(t *testing.T, ids []string) {
				assert.Equal(t, len(validTests), len(ids))

				for idx, testID := range validTests {
					assert.Equal(t, testID.s, ids[idx])
				}
			},
			errExpected: nil,
		},
		"error": {
			idsFn: func() []string {
				return []string{"invalid"}
			},
			errExpected: fmt.Errorf("invalid"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ids := test.idsFn()
			results, err := ParseStrings(ids...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, results)
		})
	}
}

func Test_ToUUIDs(t *testing.T) {
	tests := map[string]struct {
		idsFn       func() []string
		assertFn    func(*testing.T, []uuid.UUID)
		errExpected error
	}{
		"happy path": {
			idsFn: func() []string {
				ids := make([]string, len(validTests))
				for idx, testID := range validTests {
					ids[idx] = testID.s
				}
				return ids
			},
			assertFn: func(t *testing.T, ids []uuid.UUID) {
				assert.Equal(t, len(validTests), len(ids))

				for idx, testID := range validTests {
					assert.Equal(t, testID.u, ids[idx])
				}
			},
			errExpected: nil,
		},
		"error": {
			idsFn: func() []string {
				return []string{"invalid"}
			},
			errExpected: fmt.Errorf("invalid"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ids := test.idsFn()
			results, err := ToUUIDs(ids...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, results)
		})
	}
}

func TestToShortID(t *testing.T) {
	for _, test := range validTests {
		t.Run(test.s, func(t *testing.T) {
			result, err := ToShortID(test.s)
			assert.NoError(t, err)
			assert.Equal(t, test.s, result)

			result, err = ToShortID(test.u.String())
			assert.NoError(t, err)
			assert.Equal(t, test.s, result)
		})
	}
}

func TestToShortID_Errors(t *testing.T) {
	tests := map[string]struct {
		input       string
		errExpected string
	}{
		"empty string": {
			input:       "",
			errExpected: "empty",
		},
		"wrong length": {
			input:       "nope",
			errExpected: "incorrect length",
		},
		"invalid shortid": {
			input:       "!5wum892ok3t32mhreas0h26ba",
			errExpected: "invalid",
		},
		"invalid uuid": {
			input:       "!57c6ae2-be20-47bd-8ab8-cbc3428829bb",
			errExpected: "invalid",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := ToShortID(test.input)
			assert.ErrorContains(t, err, test.errExpected)
			assert.Equal(t, "", result)
		})
	}
}

func TestNewNanoID(t *testing.T) {
	tests := map[string]struct {
		inputPrefix string
		errExpected error
	}{
		"get ID with specific prefix": {
			inputPrefix: "cmp",
			errExpected: nil,
		},
		"get ID with default prefix": {
			inputPrefix: "",
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := NewNanoID(test.inputPrefix)
			assert.Len(t, result, 26)
			assert.Regexp(t, regexp.MustCompile("^[a-z0-9]*$"), result[3:])
			if test.inputPrefix == "" {
				assert.Equal(t, "def", result[:3])
				return
			}
			assert.Equal(t, test.inputPrefix, result[:3])
		})
	}
}

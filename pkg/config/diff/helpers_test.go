package diff

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"hegel.dev/go/hegel"
)

type HelpersSuite struct {
	suite.Suite
}

func TestHelpersSuite(t *testing.T) {
	suite.Run(t, new(HelpersSuite))
}

func (s *HelpersSuite) TestMapDiff_BothNil() {
	result := MapDiff("test", nil, nil)
	s.Nil(result)
}

func (s *HelpersSuite) TestMapDiff_BothEmpty() {
	result := MapDiff("test", map[string]string{}, map[string]string{})
	s.Nil(result)
}

func (s *HelpersSuite) TestMapDiff_Identical() {
	old := map[string]string{"a": "1", "b": "2"}
	new := map[string]string{"a": "1", "b": "2"}
	result := MapDiff("env_vars", old, new)
	s.Require().NotNil(result)
	s.Equal("env_vars", result.Key)
	s.Len(result.Children, 2)
	summary := result.Summary()
	s.False(summary.HasChanged)
	s.Equal(2, summary.Unchanged)
}

func (s *HelpersSuite) TestMapDiff_Added() {
	old := map[string]string{}
	new := map[string]string{"a": "1"}
	result := MapDiff("vars", old, new)
	s.Require().NotNil(result)
	s.Len(result.Children, 1)
	s.Equal(OpAdd, result.Children[0].Diff.Op)
}

func (s *HelpersSuite) TestMapDiff_Removed() {
	old := map[string]string{"a": "1"}
	new := map[string]string{}
	result := MapDiff("vars", old, new)
	s.Require().NotNil(result)
	s.Len(result.Children, 1)
	s.Equal(OpRemove, result.Children[0].Diff.Op)
}

func (s *HelpersSuite) TestMapDiff_Changed() {
	old := map[string]string{"a": "1"}
	new := map[string]string{"a": "2"}
	result := MapDiff("vars", old, new)
	s.Require().NotNil(result)
	s.Len(result.Children, 1)
	s.Equal(OpChange, result.Children[0].Diff.Op)
	s.Contains(result.Children[0].Diff.Diff, "'1' -> '2'")
}

func (s *HelpersSuite) TestMapDiff_Mixed() {
	old := map[string]string{"a": "1", "b": "2", "c": "3"}
	new := map[string]string{"a": "1", "b": "changed", "d": "4"}
	result := MapDiff("env_vars", old, new)
	s.Require().NotNil(result)
	s.Len(result.Children, 4) // a, b, c, d
	summary := result.Summary()
	s.True(summary.HasChanged)
	s.Equal(1, summary.Unchanged) // a
	s.Equal(1, summary.Changed)   // b
	s.Equal(1, summary.Removed)   // c
	s.Equal(1, summary.Added)     // d
}

func (s *HelpersSuite) TestMapDiff_SortedKeys() {
	old := map[string]string{"z": "1", "a": "2", "m": "3"}
	new := map[string]string{"z": "1", "a": "2", "m": "3"}
	result := MapDiff("vars", old, new)
	s.Require().NotNil(result)
	s.Equal("a", result.Children[0].Key)
	s.Equal("m", result.Children[1].Key)
	s.Equal("z", result.Children[2].Key)
}

func (s *HelpersSuite) TestWithBoolDiff_Same() {
	d := NewDiff(WithKey("flag"), WithBoolDiff(true, true))
	s.Equal(OpNoop, d.Diff.Op)
}

func (s *HelpersSuite) TestWithBoolDiff_Changed() {
	d := NewDiff(WithKey("flag"), WithBoolDiff(false, true))
	s.Equal(OpChange, d.Diff.Op)
	s.Contains(d.Diff.Diff, "'false' -> 'true'")
}

func (s *HelpersSuite) TestWithOptionalStringDiff_BothNil() {
	d := NewDiff(WithKey("sched"), WithOptionalStringDiff(nil, nil))
	s.Equal(OpNoop, d.Diff.Op)
}

func (s *HelpersSuite) TestWithOptionalStringDiff_NilToValue() {
	v := "0 2 * * *"
	d := NewDiff(WithKey("sched"), WithOptionalStringDiff(nil, &v))
	s.Equal(OpAdd, d.Diff.Op)
}

func (s *HelpersSuite) TestWithOptionalStringDiff_ValueToNil() {
	v := "0 2 * * *"
	d := NewDiff(WithKey("sched"), WithOptionalStringDiff(&v, nil))
	s.Equal(OpRemove, d.Diff.Op)
}

func (s *HelpersSuite) TestWithOptionalStringDiff_Changed() {
	old := "0 2 * * *"
	new := "0 4 * * *"
	d := NewDiff(WithKey("sched"), WithOptionalStringDiff(&old, &new))
	s.Equal(OpChange, d.Diff.Op)
}

func (s *HelpersSuite) TestWithOptionalBoolDiff_BothNil() {
	d := NewDiff(WithKey("flag"), WithOptionalBoolDiff(nil, nil))
	s.Equal(OpNoop, d.Diff.Op)
}

func (s *HelpersSuite) TestWithOptionalBoolDiff_NilToTrue() {
	v := true
	d := NewDiff(WithKey("flag"), WithOptionalBoolDiff(nil, &v))
	s.Equal(OpChange, d.Diff.Op)
	s.Contains(d.Diff.Diff, "'false' -> 'true'")
}

func (s *HelpersSuite) TestWithStringSliceDiff_BothEmpty() {
	d := NewDiff(WithKey("deps"), WithStringSliceDiff(nil, nil))
	s.Equal(OpNoop, d.Diff.Op)
}

func (s *HelpersSuite) TestWithStringSliceDiff_Added() {
	d := NewDiff(WithKey("deps"), WithStringSliceDiff(nil, []string{"a", "b"}))
	s.Equal(OpAdd, d.Diff.Op)
}

func (s *HelpersSuite) TestWithStringSliceDiff_SameUnordered() {
	d := NewDiff(WithKey("deps"), WithStringSliceDiff([]string{"b", "a"}, []string{"a", "b"}))
	s.Equal(OpNoop, d.Diff.Op)
}

func (s *HelpersSuite) TestWithStringSliceDiff_Changed() {
	d := NewDiff(WithKey("deps"), WithStringSliceDiff([]string{"a", "b"}, []string{"a", "c"}))
	s.Equal(OpChange, d.Diff.Op)
	s.Contains(d.Diff.Diff, "'a, b' -> 'a, c'")
}

func TestValueDiffProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		old, new := hegel.Draw(t, hegel.Text()), hegel.Draw(t, hegel.Text())
		if hegel.Draw(t, hegel.Booleans()) {
			new = old
		}
		want := OpChange
		switch {
		case old == new:
			want = OpNoop
		case old == "":
			want = OpAdd
		case new == "":
			want = OpRemove
		}
		text := NewDiff(WithStringDiff(old, new))
		contents := NewDiff(WithContentDiff(old, new))
		require.Equal(t, want, text.Diff.Op)
		require.Equal(t, want, contents.Diff.Op)
		require.Equal(t, old, contents.Diff.Before)
		require.Equal(t, new, contents.Diff.After)
		require.Equal(t, old != new, text.Summary().HasChanged)
		require.Equal(t, old != new, contents.Summary().HasChanged)
		require.Equal(t, old == new, contents.FormatChanged("") == "")
	})
}

func TestMapDiffProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		keySet := hegel.Draw(t, hegel.Maps(hegel.FromRegex(`[a-z]{1,8}`, true), hegel.Just(struct{}{})).MaxSize(12))
		keys := slices.Sorted(maps.Keys(keySet))
		old, new := map[string]string{}, map[string]string{}
		expected := map[string]Op{}
		for _, key := range keys {
			oldValue, newValue := hegel.Draw(t, hegel.Text()), hegel.Draw(t, hegel.Text())
			if hegel.Draw(t, hegel.Booleans()) {
				newValue = oldValue
			}
			wasPresent, isPresent := hegel.Draw(t, hegel.Booleans()), hegel.Draw(t, hegel.Booleans())
			if wasPresent {
				old[key] = oldValue
			}
			if isPresent {
				new[key] = newValue
			}
			switch {
			case !wasPresent && isPresent:
				expected[key] = OpAdd
			case wasPresent && !isPresent:
				expected[key] = OpRemove
			case wasPresent && oldValue != newValue:
				expected[key] = OpChange
			case wasPresent:
				expected[key] = OpNoop
			}
		}
		oldSnapshot, newSnapshot := maps.Clone(old), maps.Clone(new)
		got := MapDiff("values", old, new)
		if len(expected) == 0 {
			require.Nil(t, got)
			return
		}
		require.NotNil(t, got)
		require.Len(t, got.Children, len(expected))
		inverse := map[Op]Op{OpAdd: OpRemove, OpRemove: OpAdd, OpChange: OpChange, OpNoop: OpNoop}
		reverse := MapDiff("values", new, old)
		for i, child := range got.Children {
			require.Equal(t, expected[child.Key], child.Diff.Op, "presence and value for %s", child.Key)
			require.Equal(t, expected[child.Key] != OpNoop, child.Summary().HasChanged)
			require.Equal(t, expected[child.Key] == OpNoop, child.FormatChanged("") == "")
			require.Equal(t, child.Key, reverse.Children[i].Key)
			require.Equal(t, inverse[child.Diff.Op], reverse.Children[i].Diff.Op)
			if i > 0 {
				require.Less(t, got.Children[i-1].Key, child.Key)
			}
		}
		require.Equal(t, got, MapDiff("values", old, new))
		require.Equal(t, oldSnapshot, old)
		require.Equal(t, newSnapshot, new)
		permutedOld, permutedNew := map[string]string{}, map[string]string{}
		for _, key := range drawPermutation(t, keys) {
			if value, ok := old[key]; ok {
				permutedOld[key] = value
			}
		}
		for _, key := range drawPermutation(t, keys) {
			if value, ok := new[key]; ok {
				permutedNew[key] = value
			}
		}
		permuted := MapDiff("values", permutedOld, permutedNew)
		require.Equal(t, got, permuted)
		machine, err := json.Marshal(got)
		require.NoError(t, err)
		permutedMachine, err := json.Marshal(permuted)
		require.NoError(t, err)
		require.Equal(t, string(machine), string(permutedMachine))
		require.False(t, MapDiff("values", old, old).Summary().HasChanged)
		independent := MapDiff("values", old, new)
		got.Children[0].Key, got.Children[0].Diff.Diff = "mutated", "mutated"
		require.Equal(t, independent, MapDiff("values", old, new))
		require.Equal(t, oldSnapshot, old)
		require.Equal(t, newSnapshot, new)
	})
}

func TestStringSliceDiffDistinguishesCollections(t *testing.T) {
	for _, pair := range []struct {
		old, new []string
		want     Op
	}{
		{[]string{"a, b"}, []string{"a", "b"}, OpChange},
		{nil, []string{""}, OpAdd},
		{[]string{""}, nil, OpRemove},
		{[]string{""}, []string{"a"}, OpChange},
		{[]string{"a"}, []string{""}, OpChange},
		{[]string{"a"}, []string{"a", "a"}, OpChange},
		{[]string{"", ", ", "\x00\n"}, []string{"\x00\n", ", ", ""}, OpNoop},
	} {
		got := NewDiff(WithStringSliceDiff(pair.old, pair.new))
		require.Equal(t, pair.want, got.Diff.Op, "%q versus %q", pair.old, pair.new)
		require.Equal(t, pair.want != OpNoop, got.Summary().HasChanged)
	}
}

func sliceFrequency(values []string) map[string]int {
	counts := make(map[string]int)
	for _, value := range values {
		counts[value]++
	}
	return counts
}

func drawPermutation[T any](t hegel.TestCase, values []T) []T {
	result := slices.Clone(values)
	for i := len(result) - 1; i > 0; i-- {
		j := hegel.Draw(t, hegel.Integers(0, i))
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func assertSliceDiff(t *hegel.T, old, new []string) {
	want := OpChange
	switch {
	case maps.Equal(sliceFrequency(old), sliceFrequency(new)):
		want = OpNoop
	case len(old) == 0:
		want = OpAdd
	case len(new) == 0:
		want = OpRemove
	}
	oldSnapshot, newSnapshot := slices.Clone(old), slices.Clone(new)
	option := WithStringSliceDiff(old, new)
	got := NewDiff(WithKey("values"), option)
	require.Equal(t, want, got.Diff.Op)
	require.Equal(t, want != OpNoop, got.Summary().HasChanged)
	require.Equal(t, want == OpNoop, got.FormatChanged("") == "")
	require.Equal(t, got, NewDiff(WithKey("values"), option))
	require.Equal(t, got, NewDiff(WithKey("values"), WithStringSliceDiff(
		drawPermutation(t, old), drawPermutation(t, new),
	)))
	inverse := map[Op]Op{OpAdd: OpRemove, OpRemove: OpAdd, OpChange: OpChange, OpNoop: OpNoop}
	require.Equal(t, inverse[want], NewDiff(WithStringSliceDiff(new, old)).Diff.Op)
	require.Equal(t, oldSnapshot, old)
	require.Equal(t, newSnapshot, new)
	independent := NewDiff(WithKey("values"), option)
	got.Diff.Op, got.Diff.Diff = OpUnknown, "mutated"
	require.Equal(t, independent, NewDiff(WithKey("values"), option))
	require.Equal(t, oldSnapshot, old)
	require.Equal(t, newSnapshot, new)
}

func TestStringSliceDiffProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		value := hegel.OneOf[string](hegel.Text(), hegel.SampledFrom([]string{"", "a", "b", ", ", "a, b", "\x00", "\n", "\r\t", "é"}))
		old := hegel.Draw(t, hegel.Lists(value).MaxSize(8))
		new := hegel.Draw(t, hegel.Lists(value).MaxSize(8))
		switch hegel.Draw(t, hegel.Integers(0, 4)) {
		case 1:
			new = drawPermutation(t, old)
		case 2:
			new = append(slices.Clone(old), "")
		case 3:
			item := hegel.Draw(t, value)
			old, new = []string{item}, []string{item, item}
		case 4:
			old, new = []string{"a, b"}, []string{"a", "b"}
		}
		assertSliceDiff(t, old, new)
		self := NewDiff(WithStringSliceDiff(old, old))
		require.False(t, self.Summary().HasChanged)

		backing := hegel.Draw(t, hegel.Lists(value).MinSize(3).MaxSize(12))
		snapshot := slices.Clone(backing)
		end := hegel.Draw(t, hegel.Integers(1, len(backing)))
		assertSliceDiff(t, backing[:end], backing[1:])
		require.Equal(t, snapshot, backing, "overlapping slices and spare capacity remain unchanged")
	})
}

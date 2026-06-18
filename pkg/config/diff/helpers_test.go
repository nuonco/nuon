package diff

import (
	"testing"

	"github.com/stretchr/testify/suite"
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

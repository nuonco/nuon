package urlpolicy

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedirectCheck(t *testing.T) {
	newPolicy := func(t *testing.T, mode Mode) *Policy {
		policy, err := New(Config{Mode: mode, Categories: map[Category]CategoryConfig{
			CategoryStackTemplates: {Entries: []string{"https://templates.example.com"}},
		}})
		require.NoError(t, err)
		return policy
	}
	target := func(t *testing.T, raw string) *url.URL {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		return u
	}

	t.Run("an unrestricted category has no check", func(t *testing.T) {
		require.Nil(t, newPolicy(t, ModeEnforce).RedirectCheck(CategoryConfigFetch, "a file"))
		var none *Policy
		require.Nil(t, none.RedirectCheck(CategoryStackTemplates, "a stack template"))
	})

	t.Run("enforce refuses an off-list target", func(t *testing.T) {
		check := newPolicy(t, ModeEnforce).RedirectCheck(CategoryStackTemplates, "a stack template")
		require.NoError(t, check(target(t, "https://templates.example.com/stack.yaml")))

		err := check(target(t, "https://other.example.com/stack.yaml"))
		var violation Violation
		require.ErrorAs(t, err, &violation)
		require.Equal(t, CategoryStackTemplates, violation.Category)
		require.Equal(t, "redirect followed for a stack template", violation.Field)
	})

	t.Run("warn reports an off-list target and follows it", func(t *testing.T) {
		policy := newPolicy(t, ModeWarn)
		var got Violations
		policy.OnWarn(func(v Violations) { got = append(got, v...) })
		check := policy.RedirectCheck(CategoryStackTemplates, "a stack template")

		require.NoError(t, check(target(t, "https://templates.example.com/stack.yaml")))
		require.Empty(t, got)
		require.NoError(t, check(target(t, "https://other.example.com/stack.yaml")))
		require.Len(t, got, 1)
		require.Equal(t, "https://other.example.com/stack.yaml", got[0].Value)
	})
}

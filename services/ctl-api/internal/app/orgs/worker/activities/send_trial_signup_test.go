package activities

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildTrialSignup(t *testing.T) {
	org := &app.Org{Name: "acme"}

	t.Run("full name with source", func(t *testing.T) {
		account := &app.Account{
			Email:      "ada@example.com",
			Identities: []app.AccountIdentity{{Name: "Ada Lovelace King"}},
		}
		got := buildTrialSignup(account, org, "CLI")
		assert.Equal(t, "Ada", got.FirstName)
		assert.Equal(t, "Lovelace King", got.LastName)
		assert.Equal(t, "ada@example.com", got.Email)
		assert.Equal(t, "Created via CLI. Org: acme", got.Notes)
		assert.Equal(t, "trial-signup", got.Subject)
		assert.Empty(t, got.CompanyName)
		assert.Empty(t, got.JobTitle)
	})

	t.Run("single name falls back to ULN", func(t *testing.T) {
		account := &app.Account{
			Email:      "ada@example.com",
			Identities: []app.AccountIdentity{{Name: "Ada"}},
		}
		got := buildTrialSignup(account, org, "Dashboard")
		assert.Equal(t, "Ada", got.FirstName)
		assert.Equal(t, "ULN", got.LastName)
		assert.Equal(t, "Created via Dashboard. Org: acme", got.Notes)
	})

	t.Run("no identities", func(t *testing.T) {
		account := &app.Account{Email: "ada@example.com"}
		got := buildTrialSignup(account, org, "CLI")
		assert.Empty(t, got.FirstName)
		assert.Equal(t, "ULN", got.LastName)
	})

	t.Run("skips empty identity names", func(t *testing.T) {
		account := &app.Account{
			Email: "ada@example.com",
			Identities: []app.AccountIdentity{
				{Name: ""},
				{Name: "Grace Hopper"},
			},
		}
		got := buildTrialSignup(account, org, "CLI")
		assert.Equal(t, "Grace", got.FirstName)
		assert.Equal(t, "Hopper", got.LastName)
	})

	t.Run("empty source omits created via prefix", func(t *testing.T) {
		account := &app.Account{Email: "ada@example.com"}
		got := buildTrialSignup(account, org, "")
		assert.Equal(t, "Org: acme", got.Notes)
	})
}

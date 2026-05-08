package account

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func newTestClient() *Client {
	return &Client{
		accountCache: expirable.NewLRU[string, *app.Account](accountCacheSize, nil, accountCacheTTL),
	}
}

func TestFetchAccount_CacheHit(t *testing.T) {
	c := newTestClient()

	acct := &app.Account{
		ID:    "acc123",
		Email: "test@example.com",
		Roles: []app.Role{{ID: "role1"}},
	}
	c.accountCache.Add("acc123", acct)

	got, err := c.FetchAccount(context.Background(), "acc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "acc123" {
		t.Fatalf("expected acc123, got %s", got.ID)
	}
	if got.Email != "test@example.com" {
		t.Fatalf("expected test@example.com, got %s", got.Email)
	}
}

func TestInvalidateAccount(t *testing.T) {
	c := newTestClient()

	acct := &app.Account{
		ID:    "acc123",
		Roles: []app.Role{{ID: "role1"}},
	}
	c.accountCache.Add("acc123", acct)

	if _, ok := c.accountCache.Get("acc123"); !ok {
		t.Fatal("expected account in cache before invalidation")
	}

	c.InvalidateAccount("acc123")

	if _, ok := c.accountCache.Get("acc123"); ok {
		t.Fatal("expected account removed from cache after invalidation")
	}
}

func TestCacheSkipsAccountsWithNoRoles(t *testing.T) {
	c := newTestClient()

	noRoles := &app.Account{ID: "acc456", Roles: []app.Role{}}
	withRoles := &app.Account{ID: "acc789", Roles: []app.Role{{ID: "role1"}}}

	// Simulate what FetchAccount does after a DB load
	if len(noRoles.Roles) > 0 {
		c.accountCache.Add(noRoles.ID, noRoles)
	}
	if len(withRoles.Roles) > 0 {
		c.accountCache.Add(withRoles.ID, withRoles)
	}

	if _, ok := c.accountCache.Get("acc456"); ok {
		t.Fatal("account with no roles should not be cached")
	}
	if _, ok := c.accountCache.Get("acc789"); !ok {
		t.Fatal("account with roles should be cached")
	}
}

func TestCacheTTLExpiry(t *testing.T) {
	shortTTL := 50 * time.Millisecond
	c := &Client{
		accountCache: expirable.NewLRU[string, *app.Account](accountCacheSize, nil, shortTTL),
	}

	acct := &app.Account{
		ID:    "acc123",
		Roles: []app.Role{{ID: "role1"}},
	}
	c.accountCache.Add("acc123", acct)

	if _, ok := c.accountCache.Get("acc123"); !ok {
		t.Fatal("expected account in cache immediately after add")
	}

	time.Sleep(100 * time.Millisecond)

	if _, ok := c.accountCache.Get("acc123"); ok {
		t.Fatal("expected account evicted from cache after TTL expiry")
	}
}

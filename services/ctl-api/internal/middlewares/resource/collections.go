package resource

import "github.com/gin-gonic/gin"

// filteredCollections lists (method, route) pairs for collection endpoints that
// authorize deferred grantees by filtering their result set to entitled rows
// rather than gating the whole endpoint. These routes are let through here and
// apply grant-scope filtering in their handler. Routes are matched on the gin
// FullPath (the registered pattern, e.g. "/v1/installs"), not the literal path.
//
// Populated as list handlers adopt scope filtering. Empty means every deferred
// request without a resolved resource fails closed.
var filteredCollections = map[string]map[string]struct{}{}

func isFilteredCollection(ctx *gin.Context) bool {
	routes, ok := filteredCollections[ctx.Request.Method]
	if !ok {
		return false
	}
	_, ok = routes[ctx.FullPath()]
	return ok
}

package helpers

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ErrorTree is the cause graph rooted at a single composite_error.
type ErrorTree struct {
	Root     *app.CompositeError `json:"root"`
	Children []*ErrorTreeNode    `json:"children,omitempty"`
}

// ErrorTreeNode is a child of an ErrorTree (or another node), carrying the
// edge metadata (Idx, IsPrimary) and any further descendants.
type ErrorTreeNode struct {
	Error     *app.CompositeError `json:"error"`
	Idx       int                 `json:"idx"`
	IsPrimary bool                `json:"is_primary"`
	Children  []*ErrorTreeNode    `json:"children,omitempty"`
}

// Tree returns the cause graph rooted at id, walked breadth-first up to
// maxDepth levels. maxDepth ≤ 0 disables the depth bound (use with care).
//
// The walk uses two queries per level (one for edges, one for child rows),
// not per node. Total query count is O(depth), not O(nodes), so wide /
// deep graphs scale linearly. A recursive CTE is the long-term path —
// this implementation avoids it without paying the N+1 cost.
func (h *Helpers) Tree(ctx context.Context, rootID string, maxDepth int) (*ErrorTree, error) {
	if rootID == "" {
		return nil, errors.New("composite_errors.Tree: rootID is required")
	}

	// Load the root row directly rather than routing through h.Get, which
	// returns an error if the row's Type is no longer in the catalog. The
	// tree view should still render for historic / deprecated types — the
	// AfterFind hydration on each row already swallows hydration errors
	// for the same reason.
	var rootRow app.CompositeError
	if err := h.db.WithContext(ctx).Where("id = ?", rootID).First(&rootRow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "load composite_error root for tree")
	}
	root := &rootRow

	tree := &ErrorTree{Root: root}
	if maxDepth == 0 {
		// Depth 0 means "root only" — leave Children empty.
		return tree, nil
	}

	// nodesByID indexes every materialized ErrorTreeNode so the next level's
	// edges can attach themselves to the right parent without another lookup.
	// The root itself is keyed by its ID even though it's an *ErrorTree, not
	// an *ErrorTreeNode — we never read it back from this map (only its ID
	// is used as a parent search key on the next level).
	visited := map[string]bool{rootID: true}
	parentChildren := map[string][]*ErrorTreeNode{}

	currentLevel := []string{rootID}
	for depth := 1; len(currentLevel) > 0 && (maxDepth <= 0 || depth <= maxDepth); depth++ {
		nextLevel, err := h.expandLevel(ctx, currentLevel, visited, parentChildren)
		if err != nil {
			return nil, err
		}
		currentLevel = nextLevel
	}

	tree.Children = parentChildren[rootID]
	return tree, nil
}

// expandLevel loads edges for every parentID in the current level, hydrates
// the matching child rows in one bulk fetch, and appends ErrorTreeNodes to
// parentChildren. Returns the IDs of the next level's parents (the children
// just materialized) so the caller can drive the BFS loop.
func (h *Helpers) expandLevel(
	ctx context.Context,
	parentIDs []string,
	visited map[string]bool,
	parentChildren map[string][]*ErrorTreeNode,
) ([]string, error) {
	if len(parentIDs) == 0 {
		return nil, nil
	}

	var edges []app.CompositeErrorCause
	if err := h.db.WithContext(ctx).
		Where("parent_id IN ?", parentIDs).
		Order("parent_id asc, idx asc, created_at asc").
		Find(&edges).Error; err != nil {
		return nil, errors.Wrap(err, "load composite_error_causes")
	}
	if len(edges) == 0 {
		return nil, nil
	}

	childIDs := make([]string, 0, len(edges))
	for _, e := range edges {
		if !visited[e.ChildID] {
			visited[e.ChildID] = true
			childIDs = append(childIDs, e.ChildID)
		}
	}
	if len(childIDs) == 0 {
		return nil, nil
	}

	var rows []*app.CompositeError
	if err := h.db.WithContext(ctx).
		Where("id IN ?", childIDs).
		Find(&rows).Error; err != nil {
		return nil, errors.Wrap(err, "load composite_error rows for tree")
	}
	rowsByID := make(map[string]*app.CompositeError, len(rows))
	for _, r := range rows {
		rowsByID[r.ID] = r
	}

	touchedParents := make(map[string]struct{}, len(parentIDs))
	for _, e := range edges {
		row, ok := rowsByID[e.ChildID]
		if !ok {
			// Parent or child soft-deleted between edge load and row load,
			// or the row failed to AfterFind — skip gracefully.
			continue
		}
		parentChildren[e.ParentID] = append(parentChildren[e.ParentID], &ErrorTreeNode{
			Error:     row,
			Idx:       e.Idx,
			IsPrimary: e.IsPrimary,
		})
		touchedParents[e.ParentID] = struct{}{}
	}

	// Apply primary-first / idx ordering on the parents we just appended to.
	// Stable preserves load order on ties.
	for parentID := range touchedParents {
		kids := parentChildren[parentID]
		sort.SliceStable(kids, func(i, j int) bool {
			if kids[i].IsPrimary != kids[j].IsPrimary {
				return kids[i].IsPrimary
			}
			return kids[i].Idx < kids[j].Idx
		})
	}

	return childIDs, nil
}

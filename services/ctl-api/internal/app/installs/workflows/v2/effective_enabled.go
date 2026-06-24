package v2

// The effective-enabled / cascade logic lives in app.ComponentEnablementResolver
// so it is shared by the workflow step generator and the service-side enable
// validation. genCtx holds one resolver and exposes thin wrappers.

func (dg *genCtx) effectiveEnabled(compID string) bool {
	return dg.enablement.EffectiveEnabled(compID)
}

func (dg *genCtx) transitiveDependentsClosure(rootIDs []string) []string {
	return dg.enablement.TransitiveDependentsClosure(rootIDs)
}

func (dg *genCtx) topoSort(ids []string) []string {
	return dg.enablement.TopoSort(ids)
}

func (dg *genCtx) reverseTopoSort(ids []string) []string {
	return dg.enablement.ReverseTopoSort(ids)
}

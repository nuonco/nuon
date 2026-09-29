package v2

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

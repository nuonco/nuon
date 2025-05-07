package refs

func uniqueifyRefs(refs []Ref) []Ref {
	seen := make(map[Ref]struct{})
	unique := make([]Ref, 0, len(refs))

	for _, ref := range refs {
		if _, exists := seen[ref]; exists {
			continue
		}

		seen[ref] = struct{}{}
		unique = append(unique, ref)
	}

	return unique
}

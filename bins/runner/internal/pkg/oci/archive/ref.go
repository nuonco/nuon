package ociarchive

import "oras.land/oras-go/v2"

func (a *archive) Ref() oras.ReadOnlyTarget {
	return a.store
}

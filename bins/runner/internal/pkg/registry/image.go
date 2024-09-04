package registry

type Image struct {
	Image        string
	Tag          string
	Architecture string
}

// Name is the full name including the tag.
func (i *Image) Name() string {
	return i.Image + ":" + i.Tag
}

package registry

type Image struct {
	Image        string
	Tag          string
	Architecture string
}

func (i *Image) Name() string {
	return i.Image + ":" + i.Tag
}

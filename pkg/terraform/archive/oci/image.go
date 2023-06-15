package oci

import "fmt"

type Image struct {
	Registry string
	Repo     string
	Tag      string
}

func (i Image) RepoURL() string {
	return fmt.Sprintf("%s/%s", i.Registry, i.Repo)
}

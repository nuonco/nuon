package vcsv1

import (
	"reflect"

	"github.com/go-faker/faker/v4"
)

func init() {
	_ = faker.AddProvider("vcsConfig", fakeVcsConfig)
}

func fakeVcsConfig(v reflect.Value) (interface{}, error) {
	return &Config{
		Cfg: &Config_PublicGitConfig{
			PublicGitConfig: &PublicGitConfig{
				Repo:      "https://github.com/jonmorehouse/empty",
				Directory: "/",
				GitRef:    "main",
			},
		},
	}, nil
}

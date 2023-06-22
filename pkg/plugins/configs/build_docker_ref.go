package configs

//use "docker-ref" {
//image = "gcr.io/my-project/my-image"
//tag	= "abcd1234"
//}

// PublicDockerPullBuild is used to pull a public docker image
type DockerRefBuild struct {
	Plugin string `hcl:"plugin,label"`

	Image string `hcl:"image"`
	Tag   string `hcl:"tag"`
}

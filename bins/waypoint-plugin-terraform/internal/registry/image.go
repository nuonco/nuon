package registry

// TODO(jm): this needs to be converted to protobufs, to properly work with Waypoint
type Image struct {
	Image    string
	Tag      string
	Location struct {
		Registry struct {
			WaypointGenerated bool
		}
	}
}

// TODO(jm): this needs to use the *docker.Image proto from waypoint/builtin/docker, but we don't have a great way of
// vendoring that proto and then using it from code. Regardless, this needs to be a proto that matches that proto def.
type DockerImage struct {
	Image    string
	Tag      string
	Location struct {
		Registry struct {
			WaypointGenerated bool
		}
	}
}

// ImageMapper maps a terraform.Image to a docker.Image structure.
func ImageMapper(src *Image) *DockerImage {
	return &DockerImage{
		Image: src.Image,
		Tag:   src.Tag,
		Location: struct {
			Registry struct{ WaypointGenerated bool }
		}{},
	}
}

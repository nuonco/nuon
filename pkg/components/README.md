# go-components

Centralized go library for generating json representations of HCL files for
waypoint clients.

The main utility needed here is the ability to name the key for a json nested
block by it's value, which cannot be done OOB with `json.Marshal`. This library
also aims to synchronize component needs from `graphql-api` with
`workers-*` microservices so we're not copoying structs between the repos, and
`graphql` is setting some of the values for the HCL generation in `workers-*`
from the UI inputs.

## Usage


```go
HCL{
  Project: "0ezv19lly186a2iazihwf6ms3s",
	App: &App{
			Name: "mario",
			Build: &UseBlock{
				Use: &DockerRef{
					Image: "kennethreitz/httpbin",
					Tag:   "latest",
				},
			},
			Deploy: &UseBlock{
  			Use: &Kubernetes{
				Name: "kubernetes",
			},
		},
	},
}

hclJSON, err := hcl.ToJSON()
if err != nil {
	return "", err
}
wpReq := &gen.QueueJobRequest{
	Job: &gen.Job{
   // ...
	WaypointHcl: &gen.Hcl{
			Contents: hclJSON,
			Format:   gen.Hcl_JSON,
		},
	},
   // ...
}
```
which will compile into the following json:
```json
{
	"project": "0ezv19lly186a2iazihwf6ms3s",

	"app": {
		"mario": {
			"build": {
				"use": {
					"docker-ref": {
						"image": "kennethreitz/httpbin",
						"tag": "latest"
					}
				}
			},

			"deploy": {
				"use": {
					"kubernetes": {}
				}
			}
		}
	}
}
```

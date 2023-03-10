package configs

import "github.com/hashicorp/hcl/v2"

const (
	defaultWaypointConfigFilename string = "waypoint.hcl"
)

// TODO(jm): over time, we'd like to eventually build configs using HCL, but for now we're just using these types to
// help testing, and to have a way to keep adding functionality.
type waypointBuildBlock struct {
	Use []waypointBuildPluginBlock `hcl:"use,block"`

	Registry []waypointRegistryPluginBlock `hcl:"registry,block"`
}

// registry block configuration, for something like `use "aws-ecr"`
type waypointRegistryPluginBlock struct {
	Use struct {
		Name string `hcl:"name,label"`

		Repository string `hcl:"repository"`
		Tag        string `hcl:"tag"`

		Remain hcl.Body `hcl:",remain"`
	} `hcl:"use,block"`
}

// build block configuration, for something like `use "docker-pull"`
type waypointBuildPluginBlock struct {
	Name string `hcl:"name,label"`

	// TODO(jm): add other fields
	Remain hcl.Body `hcl:",remain"`
}

// deploy block configuration, for something like `use "kubernetes"`
type waypointDeployBlock struct {
	Use []struct {
		Name string `hcl:"name,label"`

		// different fields for deployment

		// TODO(jm): add other fields
		Remain hcl.Body `hcl:",remain"`
	} `hcl:"use,block"`
}

type waypointAppBlock struct {
	Name   string   `hcl:"name,label"`
	Remain hcl.Body `hcl:",remain"`

	Build  []waypointBuildBlock  `hcl:"build,block"`
	Deploy []waypointDeployBlock `hcl:"deploy,block"`
}

type waypointConfig struct {
	Project string             `hcl:"project"`
	App     []waypointAppBlock `hcl:"app,block"`
}

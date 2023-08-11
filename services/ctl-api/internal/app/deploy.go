package app

type Deploy struct {
	Model

	BuildID string
	Build   Build

	InstallID string
	Install   Install

	InstanceID string
}

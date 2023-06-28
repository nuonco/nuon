// Package k8s contains components for deploying to Kubernetes.
package platform

import (
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
)

const platformName = "kubernetes"

// Options are the SDK options to use for instantiation for
// the Kubernetes plugin.
var Options = []sdk.Option{
	sdk.WithComponents(&Platform{}, &Releaser{}, &ConfigSourcer{}, &TaskLauncher{}),
}

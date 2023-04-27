// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package registry

// Comment out the following functions to implement Authentication in
// your component

//func (r *Registry) ValidateAuthFunc() interface{} {
//	return r.validateAuth
//}
//
//// AuthFunc satisfies the Authenticator interface
//func (r *Registry) AuthFunc() interface{} {
//	return r.authenticate
//}

// A ValidateAuthFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// If an error is returned, Waypoint will attempt to call
// AuthFunc
//func (r *Registry) validateAuth(
//	ctx context.Context,
//	ui terminal.UI,
//) error {
//	s := ui.Status()
//	defer s.Close()
//	s.Update("Validate authentication")
//
//	return fmt.Errorf("Unable to Authenticate")
//}

// A AuthFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// Output parameters must be *component.AuthResult, error
//func (p *Registry) authenticate(
//	ctx context.Context,
//	ui terminal.UI,
//) (*component.AuthResult, error) {
//
//	ui.Output("Describe the manual authentication steps here")
//
//	return &component.AuthResult{Authenticated: false}, nil
//}

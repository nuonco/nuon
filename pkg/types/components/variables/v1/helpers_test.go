package variablesv1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func TestVariables_EnvVars(t *testing.T) {
	tests := map[string]struct {
		configsFn func() []*Variable
		assertFn  func(*testing.T, *EnvVars)
	}{
		"happy path": {
			configsFn: func() []*Variable {
				return []*Variable{
					{
						Actual: &Variable_TerraformVariable{
							TerraformVariable: &TerraformVariable{
								Name:  "terraform",
								Value: "terraform",
							},
						},
					},
					{
						Actual: &Variable_EnvVar{
							EnvVar: &EnvVar{
								Sensitive: true,
								Name:      "env-var",
								Value:     "env-var",
							},
						},
					},
				}
			},
			assertFn: func(t *testing.T, res *EnvVars) {
				assert.Len(t, res.Env, 1)
				assert.True(t, proto.Equal(res.Env[0], &EnvVar{
					Sensitive: true,
					Name:      "env-var",
					Value:     "env-var",
				}))
			},
		},
		"nil": {
			configsFn: func() []*Variable {
				return nil
			},
			assertFn: func(t *testing.T, res *EnvVars) {
				assert.NotNil(t, res.Env)
				assert.Len(t, res.Env, 0)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfgs := &Variables{
				Variables: test.configsFn(),
			}
			res := cfgs.EnvVars()
			test.assertFn(t, res)
		})
	}
}

func TestVariables_HelmValues(t *testing.T) {
	tests := map[string]struct {
		configsFn func() []*Variable
		assertFn  func(*testing.T, *HelmValues)
	}{
		"happy path": {
			configsFn: func() []*Variable {
				return []*Variable{
					{
						Actual: &Variable_TerraformVariable{
							TerraformVariable: &TerraformVariable{
								Name:  "terraform",
								Value: "terraform",
							},
						},
					},
					{
						Actual: &Variable_HelmValue{
							HelmValue: &HelmValue{
								Name:  "helm",
								Value: "helm",
							},
						},
					},
				}
			},
			assertFn: func(t *testing.T, res *HelmValues) {
				assert.Len(t, res.Values, 1)
				assert.True(t, proto.Equal(res.Values[0], &HelmValue{
					Name:  "helm",
					Value: "helm",
				}))
			},
		},
		"nil": {
			configsFn: func() []*Variable {
				return nil
			},
			assertFn: func(t *testing.T, res *HelmValues) {
				assert.NotNil(t, res.Values)
				assert.Len(t, res.Values, 0)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfgs := &Variables{
				Variables: test.configsFn(),
			}
			res := cfgs.HelmValues()
			test.assertFn(t, res)
		})
	}
}

func TestVariables_HelmValueMaps(t *testing.T) {
	tests := map[string]struct {
		configsFn func() []*Variable
		assertFn  func(*testing.T, []*HelmValuesMap)
	}{
		"happy path": {
			configsFn: func() []*Variable {
				return []*Variable{
					{
						Actual: &Variable_TerraformVariable{
							TerraformVariable: &TerraformVariable{
								Name:  "terraform",
								Value: "terraform",
							},
						},
					},
					{
						Actual: &Variable_HelmValuesMap{
							HelmValuesMap: &HelmValuesMap{
								Values: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"helm-values": nil,
									},
								},
							},
						},
					},
				}
			},
			assertFn: func(t *testing.T, res []*HelmValuesMap) {
				assert.Len(t, res, 1)
				assert.True(t, proto.Equal(res[0], &HelmValuesMap{
					Values: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"helm-values": nil,
						},
					},
				}))
			},
		},
		"nil": {
			configsFn: func() []*Variable {
				return nil
			},
			assertFn: func(t *testing.T, res []*HelmValuesMap) {
				assert.NotNil(t, res)
				assert.Len(t, res, 0)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfgs := &Variables{
				Variables: test.configsFn(),
			}
			res := cfgs.HelmValueMaps()
			test.assertFn(t, res)
		})
	}
}

func TestVariables_WaypointVariables(t *testing.T) {
	tests := map[string]struct {
		configsFn func() []*Variable
		assertFn  func(*testing.T, *WaypointVariables)
	}{
		"happy path": {
			configsFn: func() []*Variable {
				return []*Variable{
					{
						Actual: &Variable_TerraformVariable{
							TerraformVariable: &TerraformVariable{
								Name:  "terraform",
								Value: "terraform",
							},
						},
					},
					{
						Actual: &Variable_WaypointVariable{
							WaypointVariable: &WaypointVariable{
								Name:  "helm",
								Value: "helm",
							},
						},
					},
				}
			},
			assertFn: func(t *testing.T, res *WaypointVariables) {
				assert.Len(t, res.Variables, 1)
				assert.True(t, proto.Equal(res.Variables[0], &WaypointVariable{
					Name:  "helm",
					Value: "helm",
				}))
			},
		},
		"nil": {
			configsFn: func() []*Variable {
				return nil
			},
			assertFn: func(t *testing.T, res *WaypointVariables) {
				assert.NotNil(t, res.Variables)
				assert.Len(t, res.Variables, 0)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfgs := &Variables{
				Variables: test.configsFn(),
			}
			res := cfgs.WaypointVariables()
			test.assertFn(t, res)
		})
	}
}

func TestVariables_TerraformVariables(t *testing.T) {
	tests := map[string]struct {
		configsFn func() []*Variable
		assertFn  func(*testing.T, *TerraformVariables)
	}{
		"happy path": {
			configsFn: func() []*Variable {
				return []*Variable{
					{
						Actual: &Variable_TerraformVariable{
							TerraformVariable: &TerraformVariable{
								Name:  "terraform",
								Value: "terraform",
							},
						},
					},
					{
						Actual: &Variable_WaypointVariable{
							WaypointVariable: &WaypointVariable{
								Name:  "helm",
								Value: "helm",
							},
						},
					},
				}
			},
			assertFn: func(t *testing.T, res *TerraformVariables) {
				assert.Len(t, res.Variables, 1)
				assert.True(t, proto.Equal(res.Variables[0], &TerraformVariable{
					Name:  "terraform",
					Value: "terraform",
				}))
			},
		},
		"nil": {
			configsFn: func() []*Variable {
				return nil
			},
			assertFn: func(t *testing.T, res *TerraformVariables) {
				assert.NotNil(t, res)
				assert.Len(t, res.Variables, 0)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfgs := &Variables{
				Variables: test.configsFn(),
			}
			res := cfgs.TerraformVariables()
			test.assertFn(t, res)
		})
	}
}

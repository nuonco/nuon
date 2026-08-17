package airgap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap/statestore"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func latebindEnvelope(t *testing.T) *Envelope {
	t.Helper()
	createPlan := json.RawMessage(`{"sandbox_run_plan":{"install_id":"in1","aws_auth":{"region":"us-west-2","static":{"key":"old"},"assume_role":{"role_arn":"arn:old"},"use_default":false}}}`)
	applyPlan := json.RawMessage(`{"sandbox_run_plan":{"install_id":"in1","apply_plan_contents":"stale"}}`)
	deployPlan := json.RawMessage(`{"deploy_plan":{"install_id":"in1","apply_plan_contents":"stale","kubernetes_manifest":{"cluster_info":{"id":"old","endpoint":"https://old","ca_data":"old-ca","aws_auth":{"assume_role":{"role_arn":"arn:old"},"use_default":false}}}}}`)
	return &Envelope{
		Version:               "v0",
		InstallID:             "in1",
		CreatedAt:             time.Now().UTC(),
		ForceDefaultCloudAuth: true,
		Steps: []Step{
			{ID: "create", Name: "create", JobType: "sandbox-terraform", JobOperation: "create-apply-plan", JobGroup: "sandbox", CompositePlan: createPlan},
			{ID: "apply", Name: "apply", JobType: "sandbox-terraform", JobOperation: "apply-plan", JobGroup: "sandbox", DependsOn: []string{"create"}, PlanFromStep: "create", CompositePlan: applyPlan},
			{ID: "deploy", Name: "deploy", JobType: "kubernetes-manifest-deploy", JobOperation: "apply-plan", JobGroup: "deploy", DependsOn: []string{"apply"}, CompositePlan: deployPlan},
		},
	}
}

func TestRenderStepPlanChainsAndRebinds(t *testing.T) {
	envelope := latebindEnvelope(t)
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.GetJobPlanJSON(ctx, "apply")
	require.ErrorContains(t, err, "no execution result recorded")

	raw, err := client.GetJobPlanJSON(ctx, "create")
	require.NoError(t, err)
	var createPlan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &createPlan))
	auth := createPlan["sandbox_run_plan"].(map[string]any)["aws_auth"].(map[string]any)
	require.Equal(t, true, auth["use_default"])
	require.Nil(t, auth["assume_role"])
	require.Nil(t, auth["static"])

	tfplan := []byte("rendered-tfplan-bytes")
	var executionReq models.ServiceCreateRunnerJobExecutionRequest
	execution, err := client.CreateJobExecution(ctx, "create", &executionReq)
	require.NoError(t, err)
	_, err = client.CreateJobExecutionResult(ctx, "create", execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:            true,
		ContentsCompressed: base64.URLEncoding.EncodeToString(tfplan),
	})
	require.NoError(t, err)
	_, err = client.CreateJobExecutionOutputs(ctx, "create", execution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{
		Outputs: map[string]any{"cluster": map[string]any{"name": "fresh", "endpoint": "https://fresh", "certificate_authority_data": "fresh-ca"}},
	})
	require.NoError(t, err)
	_, err = client.UpdateJobExecution(ctx, "create", execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatusFinished})
	require.NoError(t, err)

	raw, err = client.GetJobPlanJSON(ctx, "apply")
	require.NoError(t, err)
	var applyPlan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &applyPlan))
	contents := applyPlan["sandbox_run_plan"].(map[string]any)["apply_plan_contents"].(string)
	decoded, err := base64.StdEncoding.DecodeString(contents)
	require.NoError(t, err)
	require.Equal(t, tfplan, decoded)

	raw, err = client.GetJobPlanJSON(ctx, "deploy")
	require.NoError(t, err)
	var deployPlan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &deployPlan))
	info := deployPlan["deploy_plan"].(map[string]any)["kubernetes_manifest"].(map[string]any)["cluster_info"].(map[string]any)
	require.Equal(t, "fresh", info["id"])
	require.Equal(t, "https://fresh", info["endpoint"])
	require.Equal(t, "fresh-ca", info["ca_data"])
	deployAuth := info["aws_auth"].(map[string]any)
	require.Equal(t, true, deployAuth["use_default"])
	require.Nil(t, deployAuth["assume_role"])
}

func TestStepPlanPersistedOnRender(t *testing.T) {
	envelope := latebindEnvelope(t)
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()

	_, found, err := store.ReadFile(statestore.StepPlanKey("create"))
	require.NoError(t, err)
	require.False(t, found, "plan must not exist before the runner renders it")

	rendered, err := client.GetJobPlanJSON(ctx, "create")
	require.NoError(t, err)

	persisted, found, err := store.ReadFile(statestore.StepPlanKey("create"))
	require.NoError(t, err)
	require.True(t, found)
	require.JSONEq(t, rendered, string(persisted))

	var plan map[string]any
	require.NoError(t, json.Unmarshal(persisted, &plan))
	auth := plan["sandbox_run_plan"].(map[string]any)["aws_auth"].(map[string]any)
	require.Equal(t, true, auth["use_default"], "persisted plan must be the late-bound render, not the vendor envelope")

	_, err = client.GetJobCompositePlan(ctx, "create")
	require.NoError(t, err)
	_, found, err = store.ReadFile(statestore.StepPlanKey("create"))
	require.NoError(t, err)
	require.True(t, found)

	_, err = client.GetJobPlanJSON(ctx, "apply")
	require.ErrorContains(t, err, "no execution result recorded")
	_, found, err = store.ReadFile(statestore.StepPlanKey("apply"))
	require.NoError(t, err)
	require.False(t, found, "failed renders must not persist a plan")
}

func TestRenderStepPlanRebindsInstallStackOutputs(t *testing.T) {
	createPlan := json.RawMessage(`{
		"sandbox_run_plan": {
			"install_id": "in1",
			"vars": {
				"vpc_id": "vpc-old",
				"private_subnet": "subnet-old-b",
				"region": "us-east-1",
				"description": "lives in vpc-old"
			},
			"state": {
				"install_stack": {
					"populated": true,
					"outputs": {
						"vpc_id": "vpc-old",
						"private_subnets": "subnet-old-a,subnet-old-b",
						"region": "us-east-1",
						"request_type": "Create"
					}
				}
			}
		}
	}`)
	envelope := &Envelope{
		Version:   "v0",
		InstallID: "in1",
		CreatedAt: time.Now().UTC(),
		Steps: []Step{
			{ID: "create", Name: "create", JobType: "sandbox-terraform", JobOperation: "create-apply-plan", JobGroup: "sandbox", CompositePlan: createPlan},
		},
	}
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)
	client.SetInstallStackOutputs(map[string]any{
		"vpc_id":          "vpc-new",
		"private_subnets": "subnet-new-a,subnet-new-b",
		"region":          "us-east-1",
		"request_type":    "Create",
		"extra_output":    "added-later",
	})

	raw, err := client.GetJobPlanJSON(context.Background(), "create")
	require.NoError(t, err)
	var plan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &plan))
	srp := plan["sandbox_run_plan"].(map[string]any)

	vars := srp["vars"].(map[string]any)
	require.Equal(t, "vpc-new", vars["vpc_id"])
	require.Equal(t, "subnet-new-b", vars["private_subnet"])
	require.Equal(t, "us-east-1", vars["region"])
	require.Equal(t, "lives in vpc-old", vars["description"], "substring occurrences must not be rewritten")

	outputs := srp["state"].(map[string]any)["install_stack"].(map[string]any)["outputs"].(map[string]any)
	require.Equal(t, "vpc-new", outputs["vpc_id"])
	require.Equal(t, "subnet-new-a,subnet-new-b", outputs["private_subnets"])
	require.Equal(t, "added-later", outputs["extra_output"])
}

func TestRenderStepPlanRebindsMapValuedStackOutputs(t *testing.T) {
	createPlan := json.RawMessage(`{
		"sandbox_run_plan": {
			"install_id": "in1",
			"vars": {
				"break_glass_iam_role_arn": "__NUON_AIRGAP_STACK_break_glass_role_arns_0__"
			},
			"state": {
				"install_stack": {
					"populated": true,
					"outputs": {
						"break_glass_role_arns": {
							"in1-sandbox-break-glass": "__NUON_AIRGAP_STACK_break_glass_role_arns_0__"
						},
						"custom_role_arns": {}
					}
				}
			}
		}
	}`)
	envelope := &Envelope{
		Version:   "v0",
		InstallID: "in1",
		CreatedAt: time.Now().UTC(),
		Steps: []Step{
			{ID: "create", Name: "create", JobType: "sandbox-terraform", JobOperation: "create-apply-plan", JobGroup: "sandbox", CompositePlan: createPlan},
		},
	}
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)
	client.SetInstallStackOutputs(map[string]any{
		"break_glass_role_arns": map[string]any{
			"in1-sandbox-break-glass": "arn:aws:iam::123:role/in1-sandbox-break-glass",
		},
		"custom_role_arns": map[string]any{},
	})

	raw, err := client.GetJobPlanJSON(context.Background(), "create")
	require.NoError(t, err)
	var plan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &plan))
	srp := plan["sandbox_run_plan"].(map[string]any)

	vars := srp["vars"].(map[string]any)
	require.Equal(t, "arn:aws:iam::123:role/in1-sandbox-break-glass", vars["break_glass_iam_role_arn"])

	outputs := srp["state"].(map[string]any)["install_stack"].(map[string]any)["outputs"].(map[string]any)
	require.Equal(t, map[string]any{
		"in1-sandbox-break-glass": "arn:aws:iam::123:role/in1-sandbox-break-glass",
	}, outputs["break_glass_role_arns"], "snapshot must stay map-shaped after rebinding")
	require.Equal(t, map[string]any{}, outputs["custom_role_arns"])
}

func TestForceDefaultCloudAuthKeepsExistingStackRoles(t *testing.T) {
	createPlan := json.RawMessage(`{
		"sandbox_run_plan": {
			"install_id": "in1",
			"aws_auth": {
				"region": "us-east-1",
				"assume_role": {"role_arn": "arn:aws:iam::111:role/in1-provision"},
				"use_default": false
			},
			"kubernetes_manifest": {
				"cluster_info": {
					"aws_auth": {
						"assume_role": {"role_arn": "arn:aws:iam::111:role/in1-cluster-access"},
						"use_default": false
					}
				}
			},
			"state": {
				"install_stack": {
					"populated": true,
					"outputs": {"iam_role_provision_arn": "arn:aws:iam::111:role/in1-provision"}
				}
			}
		}
	}`)
	envelope := &Envelope{
		Version:               "v0",
		InstallID:             "in1",
		CreatedAt:             time.Now().UTC(),
		ForceDefaultCloudAuth: true,
		Steps: []Step{
			{ID: "create", Name: "create", JobType: "sandbox-terraform", JobOperation: "create-apply-plan", JobGroup: "sandbox", CompositePlan: createPlan},
		},
	}
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)
	client.SetInstallStackOutputs(map[string]any{
		"iam_role_provision_arn": "arn:aws:iam::999:role/in1-provision",
	})

	raw, err := client.GetJobPlanJSON(context.Background(), "create")
	require.NoError(t, err)
	var plan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &plan))
	srp := plan["sandbox_run_plan"].(map[string]any)

	auth := srp["aws_auth"].(map[string]any)
	require.Equal(t, false, auth["use_default"], "rebound install-stack role must keep assume-role auth")
	assume := auth["assume_role"].(map[string]any)
	require.Equal(t, "arn:aws:iam::999:role/in1-provision", assume["role_arn"], "role ARN must be rebound to this environment's stack role")

	clusterAuth := srp["kubernetes_manifest"].(map[string]any)["cluster_info"].(map[string]any)["aws_auth"].(map[string]any)
	require.Equal(t, true, clusterAuth["use_default"], "roles not created by the stack must still fall back to ambient credentials")
	require.Nil(t, clusterAuth["assume_role"])
}

func sandboxRebindEnvelope(t *testing.T) *Envelope {
	t.Helper()
	sandboxApply := json.RawMessage(`{"sandbox_run_plan":{"install_id":"in1"}}`)
	deployPlan := json.RawMessage(`{
		"deploy_plan": {
			"install_id": "in1",
			"terraform": {
				"vars": {
					"zone_id": "ZOLDPUBLIC000000000",
					"internal_zone_id": "ZOLDINTERNAL0000000",
					"repo_url": "unrelated.example.com/repo",
					"enabled": "true"
				},
				"state": {
					"sandbox": {
						"outputs": {
							"nuon_dns": {
								"enabled": true,
								"public_domain": {"name": "in1.nuon.run", "zone_id": "ZOLDPUBLIC000000000", "nameservers": ["ns-old-1", "ns-old-2"]},
								"internal_domain": {"name": "in1.internal.nuon.run", "zone_id": "ZOLDINTERNAL0000000"}
							},
							"vpc": {"id": "vpc-reference0001"}
						}
					},
					"install": {
						"sandbox": {
							"outputs": {
								"nuon_dns": {
									"enabled": true,
									"public_domain": {"name": "in1.nuon.run", "zone_id": "ZOLDPUBLIC000000000", "nameservers": ["ns-old-1", "ns-old-2"]},
									"internal_domain": {"name": "in1.internal.nuon.run", "zone_id": "ZOLDINTERNAL0000000"}
								},
								"vpc": {"id": "vpc-reference0001"}
							}
						}
					}
				}
			}
		}
	}`)
	return &Envelope{
		Version:   "v0",
		InstallID: "in1",
		CreatedAt: time.Now().UTC(),
		Steps: []Step{
			{ID: "sbx-apply", Name: "sbx-apply", JobType: "sandbox-terraform", JobOperation: "apply-plan", JobGroup: "sandbox", CompositePlan: sandboxApply},
			{ID: "deploy", Name: "deploy", JobType: "terraform-deploy", JobOperation: "create-apply-plan", JobGroup: "deploy", DependsOn: []string{"sbx-apply"}, CompositePlan: deployPlan},
		},
	}
}

var freshSandboxOutputs = map[string]any{
	"nuon_dns": map[string]any{
		"enabled":         true,
		"public_domain":   map[string]any{"name": "in1.nuon.run", "zone_id": "ZNEWPUBLIC000000000", "nameservers": []any{"ns-new-1", "ns-new-2"}},
		"internal_domain": map[string]any{"name": "in1.internal.nuon.run", "zone_id": "ZNEWINTERNAL0000000"},
	},
	"vpc": map[string]any{"id": "vpc-actual000002"},
}

func finishSandboxApply(t *testing.T, client *Client) {
	t.Helper()
	ctx := context.Background()
	var executionReq models.ServiceCreateRunnerJobExecutionRequest
	execution, err := client.CreateJobExecution(ctx, "sbx-apply", &executionReq)
	require.NoError(t, err)
	_, err = client.CreateJobExecutionResult(ctx, "sbx-apply", execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{Success: true})
	require.NoError(t, err)
	_, err = client.CreateJobExecutionOutputs(ctx, "sbx-apply", execution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{Outputs: freshSandboxOutputs})
	require.NoError(t, err)
	_, err = client.UpdateJobExecution(ctx, "sbx-apply", execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatusFinished})
	require.NoError(t, err)
}

func assertSandboxRebound(t *testing.T, raw string) {
	t.Helper()
	var plan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &plan))
	tf := plan["deploy_plan"].(map[string]any)["terraform"].(map[string]any)

	vars := tf["vars"].(map[string]any)
	require.Equal(t, "ZNEWPUBLIC000000000", vars["zone_id"])
	require.Equal(t, "ZNEWINTERNAL0000000", vars["internal_zone_id"])
	require.Equal(t, "unrelated.example.com/repo", vars["repo_url"], "unrelated values must not be rewritten")
	require.Equal(t, "true", vars["enabled"], "short generic values must not be rewritten")

	state := tf["state"].(map[string]any)
	for _, snapshot := range []map[string]any{
		state["sandbox"].(map[string]any)["outputs"].(map[string]any),
		state["install"].(map[string]any)["sandbox"].(map[string]any)["outputs"].(map[string]any),
	} {
		dns := snapshot["nuon_dns"].(map[string]any)
		public := dns["public_domain"].(map[string]any)
		require.Equal(t, "ZNEWPUBLIC000000000", public["zone_id"])
		require.Equal(t, []any{"ns-new-1", "ns-new-2"}, public["nameservers"])
		require.Equal(t, "ZNEWINTERNAL0000000", dns["internal_domain"].(map[string]any)["zone_id"])
		require.Equal(t, "vpc-actual000002", snapshot["vpc"].(map[string]any)["id"])
	}
}

func TestRenderStepPlanRebindsSandboxOutputs(t *testing.T) {
	envelope := sandboxRebindEnvelope(t)
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)

	finishSandboxApply(t, client)

	raw, err := client.GetJobPlanJSON(context.Background(), "deploy")
	require.NoError(t, err)
	assertSandboxRebound(t, raw)
}

func TestRenderStepPlanRebindsEmbeddedPlaceholderTokens(t *testing.T) {
	sandboxApply := json.RawMessage(`{"sandbox_run_plan":{"install_id":"in1"}}`)
	deployPlan := json.RawMessage(`{
		"deploy_plan": {
			"install_id": "in1",
			"terraform": {
				"vars": {
					"domain_name": "*.__NUON_AIRGAP_SANDBOX_nuon_dns_public_domain_name__",
					"zone_id": "__NUON_AIRGAP_SANDBOX_nuon_dns_public_domain_zone_id__"
				},
				"state": {
					"sandbox": {
						"outputs": {
							"nuon_dns": {
								"public_domain": {
									"name": "__NUON_AIRGAP_SANDBOX_nuon_dns_public_domain_name__",
									"zone_id": "__NUON_AIRGAP_SANDBOX_nuon_dns_public_domain_zone_id__"
								}
							}
						}
					}
				}
			}
		}
	}`)
	envelope := &Envelope{
		Version:   "v0",
		InstallID: "in1",
		CreatedAt: time.Now().UTC(),
		Steps: []Step{
			{ID: "sbx-apply", Name: "sbx-apply", JobType: "sandbox-terraform", JobOperation: "apply-plan", JobGroup: "sandbox", CompositePlan: sandboxApply},
			{ID: "deploy", Name: "deploy", JobType: "terraform-deploy", JobOperation: "create-apply-plan", JobGroup: "deploy", DependsOn: []string{"sbx-apply"}, CompositePlan: deployPlan},
		},
	}
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)

	finishSandboxApply(t, client)

	raw, err := client.GetJobPlanJSON(context.Background(), "deploy")
	require.NoError(t, err)
	var plan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &plan))
	vars := plan["deploy_plan"].(map[string]any)["terraform"].(map[string]any)["vars"].(map[string]any)
	require.Equal(t, "*.in1.nuon.run", vars["domain_name"], "tokens embedded in larger strings must be rewritten")
	require.Equal(t, "ZNEWPUBLIC000000000", vars["zone_id"])
}

func TestRenderStepPlanRebindsSnapshotlessPlansFromReferenceSnapshots(t *testing.T) {
	sandboxApply := json.RawMessage(`{"sandbox_run_plan":{"install_id":"in1"}}`)
	deployPlan := json.RawMessage(`{
		"deploy_plan": {
			"install_id": "in1",
			"terraform": {
				"state": {
					"sandbox": {
						"outputs": {
							"ecr": {
								"registry_id": "504178855485",
								"registry_url": "504178855485.dkr.ecr.us-east-1.amazonaws.com",
								"repository_url": "504178855485.dkr.ecr.us-east-1.amazonaws.com/in1"
							}
						}
					},
					"install_stack": {
						"outputs": {
							"account_id": "504178855485",
							"maintenance_iam_role_arn": "arn:aws:iam::504178855485:role/in1-maintenance"
						}
					}
				}
			}
		}
	}`)
	ociPlan := json.RawMessage(`{
		"sync_oci_plan": {
			"dst_registry": {
				"RegistryType": "ecr",
				"Region": "us-east-1",
				"ECRAuth": {
					"assume_role": {"role_arn": "arn:aws:iam::504178855485:role/in1-maintenance"},
					"use_default": false
				},
				"Repository": "504178855485.dkr.ecr.us-east-1.amazonaws.com/in1",
				"LoginServer": "504178855485.dkr.ecr.us-east-1.amazonaws.com"
			},
			"dst_tag": "dpl123"
		}
	}`)
	envelope := &Envelope{
		Version:               "v0",
		InstallID:             "in1",
		CreatedAt:             time.Now().UTC(),
		ForceDefaultCloudAuth: true,
		Steps: []Step{
			{ID: "sbx-apply", Name: "sbx-apply", JobType: "sandbox-terraform", JobOperation: "apply-plan", JobGroup: "sandbox", CompositePlan: sandboxApply},
			{ID: "oci", Name: "oci", JobType: "oci-sync", JobOperation: "sync", JobGroup: "sync", DependsOn: []string{"sbx-apply"}, CompositePlan: ociPlan},
			{ID: "deploy", Name: "deploy", JobType: "terraform-deploy", JobOperation: "create-apply-plan", JobGroup: "deploy", DependsOn: []string{"oci"}, CompositePlan: deployPlan},
		},
	}
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)
	client.SetInstallStackOutputs(map[string]any{
		"account_id":               "423623870300",
		"maintenance_iam_role_arn": "arn:aws:iam::423623870300:role/in1-maintenance",
	})

	ctx := context.Background()
	var executionReq models.ServiceCreateRunnerJobExecutionRequest
	execution, err := client.CreateJobExecution(ctx, "sbx-apply", &executionReq)
	require.NoError(t, err)
	_, err = client.CreateJobExecutionResult(ctx, "sbx-apply", execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{Success: true})
	require.NoError(t, err)
	_, err = client.CreateJobExecutionOutputs(ctx, "sbx-apply", execution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{Outputs: map[string]any{
		"ecr": map[string]any{
			"registry_id":    "423623870300",
			"registry_url":   "423623870300.dkr.ecr.us-east-1.amazonaws.com",
			"repository_url": "423623870300.dkr.ecr.us-east-1.amazonaws.com/in1",
		},
	}})
	require.NoError(t, err)
	_, err = client.UpdateJobExecution(ctx, "sbx-apply", execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatusFinished})
	require.NoError(t, err)

	raw, err := client.GetJobPlanJSON(ctx, "oci")
	require.NoError(t, err)
	var plan map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &plan))
	dst := plan["sync_oci_plan"].(map[string]any)["dst_registry"].(map[string]any)
	require.Equal(t, "423623870300.dkr.ecr.us-east-1.amazonaws.com/in1", dst["Repository"])
	require.Equal(t, "423623870300.dkr.ecr.us-east-1.amazonaws.com", dst["LoginServer"])
	auth := dst["ECRAuth"].(map[string]any)
	require.Equal(t, false, auth["use_default"], "rebound maintenance role exists in the target account and must be kept")
	require.Equal(t, "arn:aws:iam::423623870300:role/in1-maintenance", auth["assume_role"].(map[string]any)["role_arn"])
	require.Equal(t, "dpl123", plan["sync_oci_plan"].(map[string]any)["dst_tag"])
}

func TestRenderStepPlanRebindsSandboxOutputsAfterResume(t *testing.T) {
	envelope := sandboxRebindEnvelope(t)
	require.NoError(t, envelope.Validate())

	dir := t.TempDir()
	store, err := statestore.NewDisk(dir)
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)
	finishSandboxApply(t, client)

	resumedStore, err := statestore.NewDisk(dir)
	require.NoError(t, err)
	resumed, err := NewClient(envelope, resumedStore, zap.NewNop())
	require.NoError(t, err)

	raw, err := resumed.GetJobPlanJSON(context.Background(), "deploy")
	require.NoError(t, err)
	assertSandboxRebound(t, raw)
}

func TestRebindClusterInfoSkipsComponentBoundClusters(t *testing.T) {
	cluster := map[string]any{"name": "real-cluster", "endpoint": "https://real", "certificate_authority_data": "real-ca"}
	plan := map[string]any{
		"sandbox_bound": map[string]any{
			"cluster_info": map[string]any{"id": "__NUON_AIRGAP_CLUSTER_id__", "endpoint": "__NUON_AIRGAP_CLUSTER_endpoint__", "ca_data": "__NUON_AIRGAP_CLUSTER_ca_data__"},
		},
		"context_bound": map[string]any{
			"cluster_info": map[string]any{
				"id":       ComponentOutputPlaceholder("eks", "cluster.name"),
				"endpoint": ComponentOutputPlaceholder("eks", "cluster.endpoint"),
				"ca_data":  ComponentOutputPlaceholder("eks", "cluster.certificate_authority_data"),
			},
		},
	}

	rebindClusterInfo(plan, cluster)

	sandboxBound := plan["sandbox_bound"].(map[string]any)["cluster_info"].(map[string]any)
	require.Equal(t, "real-cluster", sandboxBound["id"])
	require.Equal(t, "https://real", sandboxBound["endpoint"])
	require.Equal(t, "real-ca", sandboxBound["ca_data"])

	contextBound := plan["context_bound"].(map[string]any)["cluster_info"].(map[string]any)
	require.Equal(t, ComponentOutputPlaceholder("eks", "cluster.name"), contextBound["id"])
	require.Equal(t, ComponentOutputPlaceholder("eks", "cluster.endpoint"), contextBound["endpoint"])
	require.Equal(t, ComponentOutputPlaceholder("eks", "cluster.certificate_authority_data"), contextBound["ca_data"])
}

func TestRenderStepPlanBindsComponentOutputs(t *testing.T) {
	certToken := ComponentOutputPlaceholder("certificate", "public_domain_certificate_arn")
	lambdaToken := ComponentOutputPlaceholder("lambda_function", "lambda_function.lambda_function_arn")
	producerPlan := json.RawMessage(`{"deploy_plan":{"install_id":"in1"}}`)
	consumerPlan := json.RawMessage(`{"deploy_plan":{"install_id":"in1","terraform_deploy":{"variables":{"domain_name_certificate_arn":"` + certToken + ` ","lambda_arn":"` + lambdaToken + `"}}}}`)
	envelope := &Envelope{
		Version:   "v0",
		InstallID: "in1",
		CreatedAt: time.Now().UTC(),
		Steps: []Step{
			{ID: "deploy-certificate-apply", Name: "certificate apply", JobType: "terraform-deploy", JobOperation: "apply-plan", JobGroup: "deploy", CompositePlan: producerPlan},
			{ID: "deploy-api-plan", Name: "api create-apply-plan", JobType: "terraform-deploy", JobOperation: "create-apply-plan", JobGroup: "deploy", DependsOn: []string{"deploy-certificate-apply"}, CompositePlan: consumerPlan},
		},
		OutputBindings: []OutputBinding{
			{Token: certToken, ComponentName: "certificate", StepID: "deploy-certificate-apply", OutputPath: "public_domain_certificate_arn"},
			{Token: lambdaToken, ComponentName: "lambda_function", StepID: "deploy-certificate-apply", OutputPath: "lambda_function.lambda_function_arn"},
		},
	}
	require.NoError(t, envelope.Validate())

	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	client, err := NewClient(envelope, store, zap.NewNop())
	require.NoError(t, err)
	ctx := context.Background()

	_, err = client.GetJobPlanJSON(ctx, "deploy-api-plan")
	require.ErrorContains(t, err, "not available yet")
	require.ErrorContains(t, err, "certificate.public_domain_certificate_arn")

	var executionReq models.ServiceCreateRunnerJobExecutionRequest
	execution, err := client.CreateJobExecution(ctx, "deploy-certificate-apply", &executionReq)
	require.NoError(t, err)
	_, err = client.CreateJobExecutionOutputs(ctx, "deploy-certificate-apply", execution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{
		Outputs: map[string]any{
			"public_domain_certificate_arn": "arn:aws:acm:us-west-2:123:certificate/demo",
			"lambda_function":               map[string]any{"lambda_function_arn": "arn:aws:lambda:us-west-2:123:function:demo"},
		},
	})
	require.NoError(t, err)
	_, err = client.UpdateJobExecution(ctx, "deploy-certificate-apply", execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatusFinished})
	require.NoError(t, err)

	raw, err := client.GetJobPlanJSON(ctx, "deploy-api-plan")
	require.NoError(t, err)
	var rendered map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &rendered))
	variables := rendered["deploy_plan"].(map[string]any)["terraform_deploy"].(map[string]any)["variables"].(map[string]any)
	require.Equal(t, "arn:aws:acm:us-west-2:123:certificate/demo ", variables["domain_name_certificate_arn"])
	require.Equal(t, "arn:aws:lambda:us-west-2:123:function:demo", variables["lambda_arn"])
}

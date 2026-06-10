package arm

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) getRunnerLinkedDeployment(inp *stacks.TemplateInput) (map[string]any, map[string]ARMParameter, error) {
	templateURL := inp.RunnerNestedStackTemplateURL
	if templateURL == "" {
		return t.getDefaultRunnerDeployment(inp), nil, nil
	}

	// Custom runner template — fetch and inspect declared parameters.
	// Unlike the generic custom-nested-stack path we do NOT hoist arbitrary
	// params. The runner template is Nuon-owned plumbing; every parameter
	// is either baked by us or has a safe default in the template itself.
	armTmpl, err := fetchARMTemplate(templateURL)
	if err != nil {
		return nil, nil, fmt.Errorf("runner linked deployment: %w", err)
	}

	// All values Nuon can supply. Only injected if the template declares
	// the parameter — otherwise ARM rejects unknown params.
	managedParams := map[string]any{
		"nuonInstallID":       inp.Install.ID,
		"nuonOrgID":           inp.Runner.OrgID,
		"nuonAppID":           inp.Install.AppID,
		"location":            "[parameters('location')]",
		"runnerId":            inp.Runner.ID,
		"runnerApiUrl":        t.cfg.RunnerAPIURL,
		"runnerInitScriptUrl": inp.RunnerInitScriptURL,
		"runnerSubnetId":      "[reference('vnetDeployment').outputs.runnerSubnetId.value]",
		"customData":          t.buildRunnerCustomData(inp),
		"commonTags":          "[variables('commonTags')]",
	}

	deploymentParams := map[string]any{}
	for paramName := range armTmpl.Parameters {
		if val, ok := managedParams[paramName]; ok {
			deploymentParams[paramName] = map[string]any{"value": val}
		}
		// Parameters not in managedParams are left to their template
		// defaults. If the template declares a required param we don't
		// know about, ARM will surface a clear deployment error.
	}

	deployment := map[string]any{
		"type":       "Microsoft.Resources/deployments",
		"apiVersion": "2022-09-01",
		"name":       "runnerDeployment",
		"dependsOn":  []string{"vnetDeployment"},
		"properties": map[string]any{
			"mode": "Incremental",
			"templateLink": map[string]any{
				"uri": templateURL,
			},
			"parameters": deploymentParams,
		},
	}

	// Nothing hoisted — runner params are never customer-facing.
	return deployment, nil, nil
}

func (t *Templates) getDefaultRunnerDeployment(inp *stacks.TemplateInput) map[string]any {
	customData := t.buildRunnerCustomData(inp)

	deployment := map[string]any{
		"type":       "Microsoft.Resources/deployments",
		"apiVersion": "2022-09-01",
		"name":       "runnerDeployment",
		"dependsOn":  []string{"vnetDeployment"},
		"properties": map[string]any{
			"mode": "Incremental",
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"parameters": map[string]any{
				"nuonInstallID":  map[string]any{"value": "[parameters('nuonInstallID')]"},
				"location":       map[string]any{"value": "[parameters('location')]"},
				"runnerSubnetId": map[string]any{"value": "[reference('vnetDeployment').outputs.runnerSubnetId.value]"},
				"customData":     map[string]any{"value": customData},
				"commonTags":     map[string]any{"value": "[variables('commonTags')]"},
			},
			"template": t.getDefaultRunnerTemplate(),
		},
	}

	return deployment
}

func (t *Templates) getDefaultRunnerTemplate() map[string]any {
	return map[string]any{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": map[string]any{
			"nuonInstallID":  map[string]any{"type": "string"},
			"location":       map[string]any{"type": "string"},
			"runnerSubnetId": map[string]any{"type": "string"},
			"customData":     map[string]any{"type": "string"},
			"commonTags":     map[string]any{"type": "object"},
		},
		"resources": []any{
			map[string]any{
				"type":       "Microsoft.Compute/virtualMachineScaleSets",
				"apiVersion": "2023-03-01",
				"name":       "[format('{0}-vmss', parameters('nuonInstallID'))]",
				"location":   "[parameters('location')]",
				"tags":       "[parameters('commonTags')]",
				"sku": map[string]any{
					"name":     "Standard_D2s_v3",
					"tier":     "Standard",
					"capacity": 1,
				},
				"identity": map[string]any{
					"type": "SystemAssigned",
				},
				"properties": map[string]any{
					"overprovision": false,
					"upgradePolicy": map[string]any{
						"mode": "Manual",
					},
					// Self-heal the runner instance when it goes unhealthy (e.g.
					// the mng process powers the VM off on shutdown). This is the
					// Azure analog to the AWS ASG EC2 health check, which is what
					// brings a shut-down AWS runner back automatically. Health is
					// sourced from the Application Health extension below.
					"automaticRepairsPolicy": map[string]any{
						"enabled":      true,
						"gracePeriod":  "PT10M",
						"repairAction": "Replace",
					},
					"virtualMachineProfile": map[string]any{
						"osProfile": map[string]any{
							"computerNamePrefix": "[parameters('nuonInstallID')]",
							"adminUsername":      "nuon",
							"customData":         "[base64(parameters('customData'))]",
							"linuxConfiguration": map[string]any{
								"disablePasswordAuthentication": true,
								"ssh": map[string]any{
									"publicKeys": []map[string]any{
										{
											"path":    "/home/nuon/.ssh/authorized_keys",
											"keyData": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDmwMWT2029b4Oem5zSKRVDCBcjoVTfsUXlbGdfGeq8tzTPwqQLGqqVJDSkVb7kIjpbRv7fpB9tJERenhixW4SmYogfMlkvOy9sw+v46chmmgDmqy5Tv7MZB5SCwGVKYHv4EcwACM+GkA5jWO9poMwQM2FIEe4QAI/YaIchGf5HlfyjB/Yh7TZkuCdQ4GdTr3zwfa4DRjFThVDIobtKLjOri0u/Hcux1gduuh1gMYqTQ6oZvAGYAgWnQOiZ7rTrQvei8+SZRwFJohXPFmLjBaqmKMHs1+fu50PBA38Jp+Eey2ghvsab0HNG0eQ0icjhmHEkJZOEZ8R2/WufAON3NtapBVlOB+aCpeeRcO9wusf5kFEr3ytoRf/p8wf397efpCvYLfw9bMmxfnyzMEb1+SoFk8xLaYeyFbJDpvBvg0+m+vmwdKhquikJVII7/r0GCkaW4e3L43aBEiBip6UTFoYep/cpeN1qq8oTrUV8kMH1rPAIpZCls0LWrJJ2OqvcYJnQYWfHZ/uT/r7B6Fu8IOlyDSdwXzy3+NGaUROPj9UWT1wtWr0xyJFdE9N82noGzhmhRlhi1tYefNt/eszG2qlVg507vKIyvmfkR5VOxA51m9fw/Cgfck/KLy3XJWoXbri2eSraHomN9jEjOCerFFvtEKXViGsl4Xj0Z3B7y3ZA9Q== nuon-azure-vm-dummy@nuon.co",
										},
									},
								},
							},
						},
						"storageProfile": map[string]any{
							"imageReference": map[string]any{
								"publisher": "Canonical",
								"offer":     "0001-com-ubuntu-server-jammy",
								"sku":       "22_04-lts-gen2",
								"version":   "latest",
							},
							"osDisk": map[string]any{
								"createOption": "FromImage",
								"managedDisk": map[string]any{
									"storageAccountType": "Standard_LRS",
								},
								"diskSizeGB": 30,
							},
						},
						"networkProfile": map[string]any{
							"networkInterfaceConfigurations": []map[string]any{
								{
									"name": "[format('{0}-nic', parameters('nuonInstallID'))]",
									"properties": map[string]any{
										"primary": true,
										"ipConfigurations": []map[string]any{
											{
												"name": "[format('{0}-ipc', parameters('nuonInstallID'))]",
												"properties": map[string]any{
													"subnet": map[string]any{
														"id": "[parameters('runnerSubnetId')]",
													},
												},
											},
										},
									},
								},
							},
						},
						// Reports instance health to the VMSS by probing the mng
						// process's /livez endpoint. When the runner is down
						// (or the VM is powered off on shutdown) the probe fails,
						// the instance is marked unhealthy, and automaticRepairsPolicy
						// replaces it.
						"extensionProfile": map[string]any{
							"extensions": []map[string]any{
								{
									"name": "ApplicationHealth",
									"properties": map[string]any{
										"publisher":               "Microsoft.ManagedServices",
										"type":                    "ApplicationHealthLinux",
										"typeHandlerVersion":      "1.0", // Binary Health States (v1.0): a 200 from the probe
										"autoUpgradeMinorVersion": true,
										"settings": map[string]any{
											"protocol":       "http",
											"port":           9999,
											"requestPath":    "/livez",
											"numberOfProbes": 3, // Require 3 consecutive failing probes
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"outputs": map[string]any{
			"vmssName": map[string]any{
				"type":  "string",
				"value": "[format('{0}-vmss', parameters('nuonInstallID'))]",
			},
			"vmssPrincipalId": map[string]any{
				"type":  "string",
				"value": "[reference(resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), '2023-03-01', 'full').identity.principalId]",
			},
		},
	}
}

func (t *Templates) buildRunnerCustomData(inp *stacks.TemplateInput) string {
	return fmt.Sprintf(`#!/bin/bash

RUNNER_ID=%s
RUNNER_API_URL=%s
RUNNER_PLATFORM=azure
CONTAINER_IMAGE_URL=%s
CONTAINER_IMAGE_TAG=%s
AWS_REGION=%s

# Remove any existing Docker packages
apt-get remove -y docker docker-engine docker.io containerd runc

# Update package index and install prerequisites
apt-get update
apt-get install -y ca-certificates curl gnupg lsb-release jq

# Add Docker's official GPG key
mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up the repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Force unmask and start Docker service
rm -f /etc/systemd/system/docker.service
rm -f /etc/systemd/system/docker.socket
systemctl daemon-reload
systemctl unmask docker.service
systemctl unmask docker.socket
systemctl enable docker
systemctl start docker

# Ensure docker group exists and set up runner user
groupadd -f docker
mkdir -p /opt/nuon/runner
useradd runner -G docker -c "" -d /opt/nuon/runner -m
chown -R runner:runner /opt/nuon/runner

cat << EOF > /opt/nuon/runner/env
RUNNER_ID=$RUNNER_ID
RUNNER_API_URL=$RUNNER_API_URL
RUNNER_PLATFORM=$RUNNER_PLATFORM
ARM_USE_MSI=true
HOST_IP=$(curl -s https://checkip.amazonaws.com)
EOF

# Determine runner binary version from public-settings endpoint
RUNNER_BINARY_VERSION="${CONTAINER_IMAGE_TAG:-latest}"
echo "determining runner binary version"
echo "> $RUNNER_API_URL/v1/runners/$RUNNER_ID/public-settings"
for i in $(seq 1 30); do
  runner_binary_version=$(curl -s "$RUNNER_API_URL/v1/runners/$RUNNER_ID/public-settings" | jq -r '.binary_version // empty')
  if [ -n "$runner_binary_version" ] && [ "$runner_binary_version" != "null" ]; then
    RUNNER_BINARY_VERSION="$runner_binary_version"
    echo "determined runner binary version: $RUNNER_BINARY_VERSION"
    break
  fi
  echo "attempt $i/30: failed to determine runner binary version, retrying in 2s"
  sleep 2
done

# Install the runner binary on the host for mng mode using install.sh
mkdir -p /opt/nuon/runner/bin
curl -fsSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/runner/install.sh > /tmp/install-runner.sh
chmod +x /tmp/install-runner.sh
yes | /tmp/install-runner.sh $RUNNER_BINARY_VERSION /opt/nuon/runner/bin
rm /tmp/install-runner.sh

# Write initial image config for the mng monitor
cat << EOF > /opt/nuon/runner/image
CONTAINER_IMAGE_URL=$CONTAINER_IMAGE_URL
CONTAINER_IMAGE_TAG=$CONTAINER_IMAGE_TAG
EOF

chown -R runner:runner /opt/nuon/runner

# Create systemd unit file for mng mode
cat << 'EOF' > /etc/systemd/system/nuon-runner-mng.service
[Unit]
Description=Nuon Runner Management Service
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
User=root
EnvironmentFile=/opt/nuon/runner/env
ExecStart=/opt/nuon/runner/bin/runner mng
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
EOF

# Reload systemd and start the mng service
systemctl daemon-reload
systemctl enable --now nuon-runner-mng
`, inp.Runner.ID, t.cfg.RunnerAPIURL, inp.Settings.ContainerImageURL, inp.Settings.ContainerImageTag, inp.Install.AzureAccount.Location)
}

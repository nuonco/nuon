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

	// Custom runner template - fetch, extract params, build linked deployment
	armTmpl, err := fetchARMTemplate(templateURL)
	if err != nil {
		return nil, nil, fmt.Errorf("runner linked deployment: %w", err)
	}

	params, hoistedParams := extractARMParameters(armTmpl, ReservedParamNames)

	nuonParams := map[string]string{
		"nuonInstallID":       inp.Install.ID,
		"nuonOrgID":           inp.Runner.OrgID,
		"nuonAppID":           inp.Install.AppID,
		"location":            "[parameters('location')]",
		"runnerId":            inp.Runner.ID,
		"runnerApiUrl":        t.cfg.RunnerAPIURL,
		"runnerInitScriptUrl": inp.RunnerInitScriptURL,
	}

	deploymentParams := map[string]any{}
	for paramName := range params {
		if val, ok := nuonParams[paramName]; ok {
			deploymentParams[paramName] = map[string]any{"value": val}
		} else if paramName == "runnerSubnetId" {
			// Wire from VNet deployment output
			deploymentParams[paramName] = map[string]any{"value": "[reference('vnetDeployment').outputs.runnerSubnetId.value]"}
		} else {
			deploymentParams[paramName] = map[string]any{"value": fmt.Sprintf("[parameters('%s')]", paramName)}
		}
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

	return deployment, hoistedParams, nil
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
		},
		"resources": []any{
			map[string]any{
				"type":       "Microsoft.Compute/virtualMachineScaleSets",
				"apiVersion": "2023-03-01",
				"name":       "[format('{0}-vmss', parameters('nuonInstallID'))]",
				"location":   "[parameters('location')]",
				"sku": map[string]any{
					"name":     "Standard_D2as_v5",
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
					"virtualMachineProfile": map[string]any{
						"osProfile": map[string]any{
							"computerNamePrefix": "[take(parameters('nuonInstallID'), 9)]",
							"adminUsername":      "nuonadmin",
							"customData":         "[base64(parameters('customData'))]",
							"linuxConfiguration": map[string]any{
								"disablePasswordAuthentication": true,
								"ssh": map[string]any{
									"publicKeys": []map[string]any{
										{
											"path":    "/home/nuonadmin/.ssh/authorized_keys",
											"keyData": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDNnSz9UjUE3hh8TxnJfY1Xg2n6e6hH0rWk0E9YWnKtLQRP8U7VqEMjlLXWZ9gqkqbfLBDFm5MaRp5MT8cJyUW3VKafMFZIcmIkUmhGW2Y70PJEIFy1jHGYghkmdVnApkm4Zk2iNJMR0FqFz+xm7yKMfjOkHKCf3tfn2zn1Y3S3VRpjPj7i1p5r5VCyVF3NpuZxE1dpfOMO/5SjJGq+C5AOhXM7dcP5HAg4HskmPPpJhfSz0lGi/n0NKTFzKnl1jP3fHY7L6AIjy0ePj+vNqEBhzpSK0VZMJW+X6kfT5USMd6BSh1Rp7R0m2yfivFCfFB3Gl+E9coHtjCR63ZJFRs3p7aiFSpq8fXwqb/v5bVip6Y3etfSnTGAP9/VxVnXIljCO1vJaRpPqw2gE9OnXYwJ6X2fxFLi0rkxT1kXvwr+JOhM14rDYSJA2iz11BvztjnD6wxIPFkTxaBmPK2c6/J6h5XJLN8TuZHGBKrT5MQbPPAWCIwH9T0aSD5VTb0=",
										},
									},
								},
							},
						},
						"storageProfile": map[string]any{
							"imageReference": map[string]any{
								"publisher": "Canonical",
								"offer":     "ubuntu-24_04-lts",
								"sku":       "server",
								"version":   "latest",
							},
							"osDisk": map[string]any{
								"createOption": "FromImage",
								"managedDisk": map[string]any{
									"storageAccountType": "Premium_LRS",
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

# Install the runner binary on the host for mng mode using install.sh
# Uses CONTAINER_IMAGE_TAG (semver) as the version, with automatic latest.txt fallback
RUNNER_BINARY_VERSION="${CONTAINER_IMAGE_TAG:-latest}"
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

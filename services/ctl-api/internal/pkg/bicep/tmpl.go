package bicep

// Generated file. DO NOT EDIT
const tmpl = `
{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "metadata": {
    "_generator": {
      "name": "bicep",
      "version": "0.36.1.42791",
      "templateHash": "14798628880930657627"
    }
  },
  "parameters": {
    "nuonInstallID": {
      "type": "string",
      "defaultValue": "{{.Install.ID}}",
      "metadata": {
        "description": "The Nuon Install ID; prefixed to resource names."
      }
    },
    "nuonOrgID": {
      "type": "string",
      "defaultValue": "{{.Runner.OrgID}}",
      "metadata": {
        "description": "The Nuon Org ID. Used in tags."
      }
    },
    "nuonAppID": {
      "type": "string",
      "defaultValue": "{{.Install.AppID}}",
      "metadata": {
        "description": "The Nuon App ID. Used in tags."
      }
    },
    "vnetCIDR": {
      "type": "string",
      "defaultValue": "10.128.0.0/16",
      "metadata": {
        "description": "Please enter the IP range (CIDR notation) for this VNet."
      }
    },
    "publicSubnet1CIDR": {
      "type": "string",
      "defaultValue": "10.128.0.0/26",
      "metadata": {
        "description": "Please enter the IP range (CIDR notation) for the public subnet"
      }
    },
    "publicSubnet2CIDR": {
      "type": "string",
      "defaultValue": "10.128.0.64/26",
      "metadata": {
        "description": "Please enter the IP range (CIDR notation) for the public subnet in the second zone (optional)"
      }
    },
    "publicSubnet3CIDR": {
      "type": "string",
      "defaultValue": "10.128.0.128/26",
      "metadata": {
        "description": "Please enter the IP range (CIDR notation) for the public subnet in the third zone (optional)"
      }
    },
    "runnerSubnetCIDR": {
      "type": "string",
      "defaultValue": "10.128.128.0/24",
      "metadata": {
        "description": "Please enter the IP range (CIDR notation) for the dedicated private subnet for the runner."
      }
    },
    "privateSubnet1CIDR": {
      "type": "string",
      "defaultValue": "10.128.130.0/24",
      "metadata": {
        "description": "Please enter the IP range (CIDR notation) for the private subnet"
      }
    },
    "privateSubnet2CIDR": {
      "type": "string",
      "defaultValue": "10.128.132.0/24",
      "metadata": {
        "description": "Please enter the IP range (CIDR notation) for the private subnet in the second zone (optional)"
      }
    },
    "privateSubnet3CIDR": {
      "type": "string",
      "defaultValue": "10.128.134.0/24",
      "metadata": {
        "description": "Please enter the IP range (CIDR notation) for the private subnet in the third zone (optional)"
      }
    },
    "location": {
      "type": "string",
      "defaultValue": "{{.Install.AzureAccount.Location}}",
      "metadata": {
        "description": "The location for all resources."
      }
    },
    "vmAdminUsername": {
      "type": "string",
      "defaultValue": "nuon_admin",
      "metadata": {
        "description": "Admin username for the VM"
      }
    },
    "vmAdminPassword": {
      "type": "securestring",
      "defaultValue": "Password!123",
      "metadata": {
        "description": "Admin password for the VM"
      }
    },
    "secrets": {
      "type": "array",
      "defaultValue": [],
      "metadata": {
        "description": "List of secrets to store in Azure Key Vault"
      }
    }
  },
  "variables": {
    "commonTags": {
      "install_nuon_co_id": "[parameters('nuonInstallID')]",
      "org_nuon_co_id": "[parameters('nuonOrgID')]",
      "app_nuon_co_id": "[parameters('nuonAppID')]"
    },
    "createPublicSubnet2": "[not(empty(parameters('publicSubnet2CIDR')))]",
    "createPublicSubnet3": "[not(empty(parameters('publicSubnet3CIDR')))]",
    "createPrivateSubnet2": "[not(empty(parameters('privateSubnet2CIDR')))]",
    "createPrivateSubnet3": "[not(empty(parameters('privateSubnet3CIDR')))]",
    "customData": "#!/bin/bash\n\nRUNNER_ID={{.Runner.ID}}\nRUNNER_API_TOKEN={{.APIToken}}\nRUNNER_API_URL={{.Settings.RunnerAPIURL}}\nAWS_REGION={{.Install.AzureAccount.Location}}\n\n# Remove any existing Docker packages\napt-get remove -y docker docker-engine docker.io containerd runc\n\n# Update package index and install prerequisites\napt-get update\napt-get install -y ca-certificates curl gnupg lsb-release\n\n# Add Docker's official GPG key\nmkdir -p /etc/apt/keyrings\ncurl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg\n\n# Set up the repository\necho \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\" | tee /etc/apt/sources.list.d/docker.list > /dev/null\n\n# Install Docker Engine\napt-get update\napt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin\n\n# Force unmask and start Docker service\nrm -f /etc/systemd/system/docker.service\nrm -f /etc/systemd/system/docker.socket\nsystemctl daemon-reload\nsystemctl unmask docker.service\nsystemctl unmask docker.socket\nsystemctl enable docker\nsystemctl start docker\n\n# Ensure docker group exists and set up runner user\ngroupadd -f docker\nmkdir -p /opt/nuon/runner\nuseradd runner -G docker -c \"\" -d /opt/nuon/runner -m\nchown -R runner:runner /opt/nuon/runner\n\ncat << EOF > /opt/nuon/runner/env\nRUNNER_ID=$RUNNER_ID\nRUNNER_API_TOKEN=$RUNNER_API_TOKEN\nRUNNER_API_URL=$RUNNER_API_URL\nARM_USE_MSI=true\n# FIXME(sdboyer) this hack must be fixed - userdata is only run on instance creation, and ip can change on each boot\nHOST_IP=$(curl -s https://checkip.amazonaws.com)\nEOF\n\n# this ⤵ is wrapped w/ single quotes to prevent variable expansion.\ncat << 'EOF' > /opt/nuon/runner/get_image_tag.sh\n#!/bin/bash\n\nset -u\n\n# source this file to get some env vars\n. /opt/nuon/runner/env\n\n# Fetch runner settings from the API\necho \"Fetching runner settings from $RUNNER_API_URL/v1/runners/$RUNNER_ID/settings\"\nRUNNER_SETTINGS=$(curl -s -H \"Authorization: Bearer $RUNNER_API_TOKEN\" \"$RUNNER_API_URL/v1/runners/$RUNNER_ID/settings\")\n\n# Extract container image URL and tag from the response\nCONTAINER_IMAGE_URL=$(echo \"$RUNNER_SETTINGS\" | grep -o '\"container_image_url\":\"[^\"]*\"' | cut -d '\"' -f 4)\nCONTAINER_IMAGE_TAG=$(echo \"$RUNNER_SETTINGS\" | grep -o '\"container_image_tag\":\"[^\"]*\"' | cut -d '\"' -f 4)\n\n# echo into a file for easier retrieval; re-create the file to avoid duplicate values.\nrm -f /opt/nuon/runner/image\necho \"CONTAINER_IMAGE_URL=$CONTAINER_IMAGE_URL\" >> /opt/nuon/runner/image\necho \"CONTAINER_IMAGE_TAG=$CONTAINER_IMAGE_TAG\" >> /opt/nuon/runner/image\n\n# export so we can get these values by sourcing this file\nexport CONTAINER_IMAGE_URL=$CONTAINER_IMAGE_URL\nexport CONTAINER_IMAGE_TAG=$CONTAINER_IMAGE_TAG\n\necho \"Using container image: $CONTAINER_IMAGE_URL:$CONTAINER_IMAGE_TAG\"\nEOF\n\nchmod +x /opt/nuon/runner/get_image_tag.sh\n/opt/nuon/runner/get_image_tag.sh\n\n# Create systemd unit file for runner\ncat << 'EOF' > /etc/systemd/system/nuon-runner.service\n[Unit]\nDescription=Nuon Runner Service\nAfter=docker.service\nRequires=docker.service\n\n[Service]\nTimeoutStartSec=0\nUser=runner\nExecStartPre=-/bin/sh -c '/usr/bin/docker stop $(/usr/bin/docker ps -a -q --filter=\"name=%n\")'\nExecStartPre=-/bin/sh -c '/usr/bin/docker rm $(/usr/bin/docker ps -a -q --filter=\"name=%n\")'\nExecStartPre=-/bin/sh -c \"yes | /usr/bin/docker system prune\"\nExecStartPre=-/bin/sh /opt/nuon/runner/get_image_tag.sh\nEnvironmentFile=/opt/nuon/runner/image\nEnvironmentFile=/opt/nuon/runner/env\nExecStartPre=echo \"Using container image: ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG}\"\nExecStartPre=/usr/bin/docker pull ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG}\nExecStart=/usr/bin/docker run --network host -v /tmp/nuon-runner:/tmp --rm --name %n -p 5000:5000 --memory \"3750g\" --cpus=\"1.75\" --env-file /opt/nuon/runner/env ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG} run\nRestart=always\nRestartSec=5\n\n[Install]\nWantedBy=default.target\nEOF\n\n# Enable SSH password authentication\nsed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config\nsed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config\nsystemctl restart sshd\n\n# Reload systemd and start the service (no SELinux on Ubuntu)\nsystemctl daemon-reload\nsystemctl enable --now nuon-runner\n"
  },
  "resources": [
    {
      "type": "Microsoft.Network/networkSecurityGroups",
      "apiVersion": "2023-04-01",
      "name": "[format('{0}-public-nsg', parameters('nuonInstallID'))]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "properties": {
        "securityRules": [
          {
            "name": "Allow-All-Inbound",
            "properties": {
              "description": "Allow all inbound traffic from any source",
              "protocol": "*",
              "sourcePortRange": "*",
              "destinationPortRange": "*",
              "sourceAddressPrefix": "*",
              "destinationAddressPrefix": "*",
              "access": "Allow",
              "priority": 200,
              "direction": "Inbound"
            }
          }
        ]
      }
    },
    {
      "type": "Microsoft.Network/networkSecurityGroups",
      "apiVersion": "2023-04-01",
      "name": "[format('{0}-private-nsg', parameters('nuonInstallID'))]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "properties": {
        "securityRules": []
      }
    },
    {
      "type": "Microsoft.Network/routeTables",
      "apiVersion": "2023-04-01",
      "name": "[format('{0}-private-routetable', parameters('nuonInstallID'))]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "properties": {
        "disableBgpRoutePropagation": false
      }
    },
    {
      "type": "Microsoft.Network/virtualNetworks",
      "apiVersion": "2023-04-01",
      "name": "[format('{0}-vnet', parameters('nuonInstallID'))]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "properties": {
        "addressSpace": {
          "addressPrefixes": [
            "[parameters('vnetCIDR')]"
          ]
        },
        "subnets": [
          {
            "name": "[format('{0}-public-subnet-zone1', parameters('nuonInstallID'))]",
            "properties": {
              "addressPrefix": "[parameters('publicSubnet1CIDR')]",
              "privateEndpointNetworkPolicies": "Disabled",
              "privateLinkServiceNetworkPolicies": "Enabled",
              "networkSecurityGroup": {
                "id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]"
              },
              "natGateway": {
                "id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]"
              }
            }
          },
          {
            "name": "[format('{0}-public-subnet-zone2', parameters('nuonInstallID'))]",
            "properties": {
              "addressPrefix": "[parameters('publicSubnet2CIDR')]",
              "privateEndpointNetworkPolicies": "Disabled",
              "privateLinkServiceNetworkPolicies": "Enabled",
              "networkSecurityGroup": {
                "id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]"
              },
              "natGateway": {
                "id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]"
              }
            }
          },
          {
            "name": "[format('{0}-public-subnet-zone3', parameters('nuonInstallID'))]",
            "properties": {
              "addressPrefix": "[parameters('publicSubnet3CIDR')]",
              "privateEndpointNetworkPolicies": "Disabled",
              "privateLinkServiceNetworkPolicies": "Enabled",
              "networkSecurityGroup": {
                "id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]"
              },
              "natGateway": {
                "id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]"
              }
            }
          },
          {
            "name": "[format('{0}-private-runner-subnet', parameters('nuonInstallID'))]",
            "properties": {
              "addressPrefix": "[parameters('runnerSubnetCIDR')]",
              "privateEndpointNetworkPolicies": "Disabled",
              "privateLinkServiceNetworkPolicies": "Enabled",
              "networkSecurityGroup": {
                "id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]"
              },
              "natGateway": {
                "id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]"
              },
              "serviceEndpoints": [
                {
                  "service": "Microsoft.KeyVault"
                },
                {
                  "service": "Microsoft.ContainerRegistry"
                }
              ]
            }
          },
          {
            "name": "[format('{0}-private-subnet-zone1', parameters('nuonInstallID'))]",
            "properties": {
              "addressPrefix": "[parameters('privateSubnet1CIDR')]",
              "privateEndpointNetworkPolicies": "Disabled",
              "privateLinkServiceNetworkPolicies": "Enabled",
              "networkSecurityGroup": {
                "id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]"
              },
              "natGateway": {
                "id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]"
              },
              "serviceEndpoints": [
                {
                  "service": "Microsoft.KeyVault"
                },
                {
                  "service": "Microsoft.ContainerRegistry"
                }
              ]
            }
          },
          {
            "name": "[format('{0}-private-subnet-zone2', parameters('nuonInstallID'))]",
            "properties": {
              "addressPrefix": "[parameters('privateSubnet2CIDR')]",
              "privateEndpointNetworkPolicies": "Disabled",
              "privateLinkServiceNetworkPolicies": "Enabled",
              "networkSecurityGroup": {
                "id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]"
              },
              "natGateway": {
                "id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]"
              },
              "serviceEndpoints": [
                {
                  "service": "Microsoft.KeyVault"
                },
                {
                  "service": "Microsoft.ContainerRegistry"
                }
              ]
            }
          },
          {
            "name": "[format('{0}-private-subnet-zone3', parameters('nuonInstallID'))]",
            "properties": {
              "addressPrefix": "[parameters('privateSubnet3CIDR')]",
              "privateEndpointNetworkPolicies": "Disabled",
              "privateLinkServiceNetworkPolicies": "Enabled",
              "networkSecurityGroup": {
                "id": "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]"
              },
              "natGateway": {
                "id": "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]"
              },
              "serviceEndpoints": [
                {
                  "service": "Microsoft.KeyVault"
                },
                {
                  "service": "Microsoft.ContainerRegistry"
                }
              ]
            }
          }
        ]
      },
      "dependsOn": [
        "[resourceId('Microsoft.Network/natGateways', format('{0}-natgw', parameters('nuonInstallID')))]",
        "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-private-nsg', parameters('nuonInstallID')))]",
        "[resourceId('Microsoft.Network/networkSecurityGroups', format('{0}-public-nsg', parameters('nuonInstallID')))]"
      ]
    },
    {
      "type": "Microsoft.KeyVault/vaults",
      "apiVersion": "2023-02-01",
      "name": "[take(format('{0}kv', parameters('nuonInstallID')), 24)]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "properties": {
        "enabledForDeployment": true,
        "enabledForTemplateDeployment": true,
        "enabledForDiskEncryption": true,
        "tenantId": "[subscription().tenantId]",
        "accessPolicies": [],
        "sku": {
          "name": "standard",
          "family": "A"
        },
        "networkAcls": {
          "defaultAction": "Deny",
          "bypass": "AzureServices",
          "ipRules": [],
          "virtualNetworkRules": [
            {
              "id": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-runner-subnet', parameters('nuonInstallID')))]"
            },
            {
              "id": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone1', parameters('nuonInstallID')))]"
            }
          ]
        }
      },
      "dependsOn": [
        "[resourceId('Microsoft.Network/virtualNetworks', format('{0}-vnet', parameters('nuonInstallID')))]"
      ]
    },
    {
      "copy": {
        "name": "keyVaultSecrets",
        "count": "[length(parameters('secrets'))]"
      },
      "type": "Microsoft.KeyVault/vaults/secrets",
      "apiVersion": "2023-02-01",
      "name": "[format('{0}/{1}', take(format('{0}kv', parameters('nuonInstallID')), 24), parameters('secrets')[copyIndex()].name)]",
      "properties": {
        "value": "[parameters('secrets')[copyIndex()].value]",
        "contentType": "[if(contains(parameters('secrets')[copyIndex()], 'contentType'), parameters('secrets')[copyIndex()].contentType, null())]",
        "attributes": {
          "enabled": "[if(contains(parameters('secrets')[copyIndex()], 'enabled'), parameters('secrets')[copyIndex()].enabled, true())]"
        }
      },
      "dependsOn": [
        "[resourceId('Microsoft.KeyVault/vaults', take(format('{0}kv', parameters('nuonInstallID')), 24))]"
      ]
    },
    {
      "type": "Microsoft.Network/publicIPAddresses",
      "apiVersion": "2023-04-01",
      "name": "[format('{0}-natgw-pip', parameters('nuonInstallID'))]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "sku": {
        "name": "Standard"
      },
      "properties": {
        "publicIPAllocationMethod": "Static"
      }
    },
    {
      "type": "Microsoft.Network/natGateways",
      "apiVersion": "2023-04-01",
      "name": "[format('{0}-natgw', parameters('nuonInstallID'))]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "sku": {
        "name": "Standard"
      },
      "properties": {
        "publicIpAddresses": [
          {
            "id": "[resourceId('Microsoft.Network/publicIPAddresses', format('{0}-natgw-pip', parameters('nuonInstallID')))]"
          }
        ],
        "idleTimeoutInMinutes": 4
      },
      "dependsOn": [
        "[resourceId('Microsoft.Network/publicIPAddresses', format('{0}-natgw-pip', parameters('nuonInstallID')))]"
      ]
    },
    {
      "type": "Microsoft.Compute/virtualMachineScaleSets",
      "apiVersion": "2023-03-01",
      "name": "[format('{0}-vmss', parameters('nuonInstallID'))]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "sku": {
        "name": "Standard_B2s",
        "tier": "Standard",
        "capacity": 1
      },
      "identity": {
        "type": "SystemAssigned"
      },
      "properties": {
        "upgradePolicy": {
          "mode": "Manual"
        },
        "virtualMachineProfile": {
          "osProfile": {
            "computerNamePrefix": "[parameters('nuonInstallID')]",
            "adminUsername": "[parameters('vmAdminUsername')]",
            "adminPassword": "[parameters('vmAdminPassword')]",
            "customData": "[base64(variables('customData'))]",
            "linuxConfiguration": {
              "disablePasswordAuthentication": false
            }
          },
          "storageProfile": {
            "imageReference": {
              "publisher": "Canonical",
              "offer": "0001-com-ubuntu-server-jammy",
              "sku": "22_04-lts-gen2",
              "version": "latest"
            },
            "osDisk": {
              "createOption": "FromImage",
              "managedDisk": {
                "storageAccountType": "Standard_LRS"
              }
            }
          },
          "networkProfile": {
            "networkInterfaceConfigurations": [
              {
                "name": "nic",
                "properties": {
                  "primary": true,
                  "ipConfigurations": [
                    {
                      "name": "ipconfig",
                      "properties": {
                        "subnet": {
                          "id": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-runner-subnet', parameters('nuonInstallID')))]"
                        }
                      }
                    }
                  ]
                }
              }
            ]
          }
        }
      },
      "dependsOn": [
        "[resourceId('Microsoft.Network/virtualNetworks', format('{0}-vnet', parameters('nuonInstallID')))]"
      ]
    },
    {
      "type": "Microsoft.Authorization/roleAssignments",
      "apiVersion": "2022-04-01",
      "name": "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'Contributor')]",
      "properties": {
        "roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b24988ac-6180-42a0-ab88-20f7382dd24c')]",
        "principalId": "[reference(resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), '2023-03-01', 'full').identity.principalId]",
        "principalType": "ServicePrincipal"
      },
      "dependsOn": [
        "[resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID')))]"
      ]
    },
    {
      "type": "Microsoft.Authorization/roleAssignments",
      "apiVersion": "2022-04-01",
      "name": "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'RoleBasedAccessControlAdministrator')]",
      "properties": {
        "roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'f58310d9-a9f6-439a-9e8d-f62e7b41a168')]",
        "principalId": "[reference(resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), '2023-03-01', 'full').identity.principalId]",
        "principalType": "ServicePrincipal"
      },
      "dependsOn": [
        "[resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID')))]"
      ]
    },
    {
      "type": "Microsoft.Authorization/roleAssignments",
      "apiVersion": "2022-04-01",
      "name": "[guid(resourceGroup().id, resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), 'AzureKubernetesServiceRBACClusterAdmin')]",
      "properties": {
        "roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b1ff04bb-8a4e-4dc4-8eb5-8693973ce19b')]",
        "principalId": "[reference(resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), '2023-03-01', 'full').identity.principalId]",
        "principalType": "ServicePrincipal"
      },
      "dependsOn": [
        "[resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID')))]"
      ]
    },
    {
      "type": "Microsoft.Resources/deploymentScripts",
      "apiVersion": "2023-08-01",
      "name": "[format('{0}-phone-home-script', parameters('nuonInstallID'))]",
      "location": "[parameters('location')]",
      "tags": "[variables('commonTags')]",
      "kind": "AzureCLI",
      "properties": {
        "azCliVersion": "2.40.0",
        "timeout": "PT30M",
        "retentionInterval": "PT1H",
        "environmentVariables": [
          {
            "name": "SUBSCRIPTION_ID",
            "value": "[subscription().subscriptionId]"
          },
          {
            "name": "SUBSCRIPTION_TENANT_ID",
            "value": "[subscription().tenantId]"
          },
          {
            "name": "RESOURCE_GROUP_ID",
            "value": "[resourceGroup().id]"
          },
          {
            "name": "RESOURCE_GROUP_NAME",
            "value": "[resourceGroup().name]"
          },
          {
            "name": "RESOURCE_GROUP_LOCATION",
            "value": "[resourceGroup().location]"
          },
          {
            "name": "VNET_ID",
            "value": "[resourceId('Microsoft.Network/virtualNetworks', format('{0}-vnet', parameters('nuonInstallID')))]"
          },
          {
            "name": "VNET_NAME",
            "value": "[format('{0}-vnet', parameters('nuonInstallID'))]"
          },
          {
            "name": "PUBLIC_SUBNET_1_ID",
            "value": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone1', parameters('nuonInstallID')))]"
          },
          {
            "name": "PUBLIC_SUBNET_1_NAME",
            "value": "[format('{0}-public-subnet-zone1', parameters('nuonInstallID'))]"
          },
          {
            "name": "PUBLIC_SUBNET_2_ID",
            "value": "[if(variables('createPublicSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone2', parameters('nuonInstallID'))), '')]"
          },
          {
            "name": "PUBLIC_SUBNET_2_NAME",
            "value": "[if(variables('createPublicSubnet2'), format('{0}-public-subnet-zone2', parameters('nuonInstallID')), '')]"
          },
          {
            "name": "PUBLIC_SUBNET_3_ID",
            "value": "[if(variables('createPublicSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone3', parameters('nuonInstallID'))), '')]"
          },
          {
            "name": "PUBLIC_SUBNET_3_NAME",
            "value": "[if(variables('createPublicSubnet3'), format('{0}-public-subnet-zone3', parameters('nuonInstallID')), '')]"
          },
          {
            "name": "PRIVATE_SUBNET_1_ID",
            "value": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone1', parameters('nuonInstallID')))]"
          },
          {
            "name": "PRIVATE_SUBNET_1_NAME",
            "value": "[format('{0}-private-subnet-zone1', parameters('nuonInstallID'))]"
          },
          {
            "name": "PRIVATE_SUBNET_2_ID",
            "value": "[if(variables('createPrivateSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone2', parameters('nuonInstallID'))), '')]"
          },
          {
            "name": "PRIVATE_SUBNET_2_NAME",
            "value": "[if(variables('createPrivateSubnet2'), format('{0}-private-subnet-zone2', parameters('nuonInstallID')), '')]"
          },
          {
            "name": "PRIVATE_SUBNET_3_ID",
            "value": "[if(variables('createPrivateSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone3', parameters('nuonInstallID'))), '')]"
          },
          {
            "name": "PRIVATE_SUBNET_3_NAME",
            "value": "[if(variables('createPrivateSubnet3'), format('{0}-private-subnet-zone3', parameters('nuonInstallID')), '')]"
          },
          {
            "name": "PUBLIC_SUBNET_IDS_CSV",
            "value": "[join(filter(createArray(resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone1', parameters('nuonInstallID'))), if(variables('createPublicSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone2', parameters('nuonInstallID'))), ''), if(variables('createPublicSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone3', parameters('nuonInstallID'))), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"
          },
          {
            "name": "PUBLIC_SUBNET_NAMES_CSV",
            "value": "[join(filter(createArray(format('{0}-public-subnet-zone1', parameters('nuonInstallID')), if(variables('createPublicSubnet2'), format('{0}-public-subnet-zone2', parameters('nuonInstallID')), ''), if(variables('createPublicSubnet3'), format('{0}-public-subnet-zone3', parameters('nuonInstallID')), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"
          },
          {
            "name": "PRIVATE_SUBNET_IDS_CSV",
            "value": "[join(filter(createArray(resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone1', parameters('nuonInstallID'))), if(variables('createPrivateSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone2', parameters('nuonInstallID'))), ''), if(variables('createPrivateSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone3', parameters('nuonInstallID'))), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"
          },
          {
            "name": "PRIVATE_SUBNET_NAMES_CSV",
            "value": "[join(filter(createArray(format('{0}-private-subnet-zone1', parameters('nuonInstallID')), if(variables('createPrivateSubnet2'), format('{0}-private-subnet-zone2', parameters('nuonInstallID')), ''), if(variables('createPrivateSubnet3'), format('{0}-private-subnet-zone3', parameters('nuonInstallID')), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"
          }
        ],
        "scriptContent": "      #!/bin/bash\n      \n      # Construct the JSON payload with stack outputs\n      #\n      # Including the credentails object for backwards compatibility.\n      # We used to need this when the org runner did the sandbox provision,\n      # but the independent runner obviates the need for this.\n      #\n      # The provision workflow still looks for auth credentials,\n      # because it needs the role ARNs to use for different jobs.\n      # Azure resource groups obviate the need for multiple roles,\n      # so we don't need to return anything.\n\n      # Create arrays for public and private subnets (filtering out empty values)\n      PUBLIC_SUBNETS=(\"$PUBLIC_SUBNET_1_ID\")\n      PUBLIC_SUBNET_NAMES=(\"$PUBLIC_SUBNET_1_NAME\")\n      if [ -n \"$PUBLIC_SUBNET_2_ID\" ]; then \n        PUBLIC_SUBNETS+=(\"$PUBLIC_SUBNET_2_ID\")\n        PUBLIC_SUBNET_NAMES+=(\"$PUBLIC_SUBNET_2_NAME\")\n      fi\n      if [ -n \"$PUBLIC_SUBNET_3_ID\" ]; then \n        PUBLIC_SUBNETS+=(\"$PUBLIC_SUBNET_3_ID\")\n        PUBLIC_SUBNET_NAMES+=(\"$PUBLIC_SUBNET_3_NAME\")\n      fi\n\n      PRIVATE_SUBNETS=(\"$PRIVATE_SUBNET_1_ID\")\n      PRIVATE_SUBNET_NAMES=(\"$PRIVATE_SUBNET_1_NAME\")\n      if [ -n \"$PRIVATE_SUBNET_2_ID\" ]; then \n        PRIVATE_SUBNETS+=(\"$PRIVATE_SUBNET_2_ID\")\n        PRIVATE_SUBNET_NAMES+=(\"$PRIVATE_SUBNET_2_NAME\")\n      fi\n      if [ -n \"$PRIVATE_SUBNET_3_ID\" ]; then \n        PRIVATE_SUBNETS+=(\"$PRIVATE_SUBNET_3_ID\")\n        PRIVATE_SUBNET_NAMES+=(\"$PRIVATE_SUBNET_3_NAME\")\n      fi\n\n      PAYLOAD=$(cat << EOF\n{\n  \"resource_group_id\": \"$RESOURCE_GROUP_ID\",\n  \"resource_group_name\": \"$RESOURCE_GROUP_NAME\",\n  \"resource_group_location\": \"$RESOURCE_GROUP_LOCATION\",\n  \"network_id\": \"$VNET_ID\",\n  \"network_name\": \"$VNET_NAME\",\n  \"public_subnet_ids\": \"$PUBLIC_SUBNET_IDS_CSV\",\n  \"public_subnet_names\": \"$PUBLIC_SUBNET_NAMES_CSV\",\n  \"private_subnet_ids\": \"$PRIVATE_SUBNET_IDS_CSV\",\n  \"private_subnet_names\": \"$PRIVATE_SUBNET_NAMES_CSV\",\n  \"subscription_id\": \"$SUBSCRIPTION_ID\",\n  \"subscription_tenant_id\": \"$SUBSCRIPTION_TENANT_ID\"\n}\nEOF\n)\n      \n      # Send the phone home request\n      curl -X POST \\\n        \"{{.CloudFormationStackVersion.PhoneHomeURL}}\" \\\n        -H \"Content-Type: application/json\" \\\n        -H \"Accept: application/json\" \\\n        -d \"$PAYLOAD\" \\\n        --fail \\\n        --silent \\\n        --show-error\n      \n      if [ $? -eq 0 ]; then\n        echo \"Phone home request sent successfully\"\n      else\n        echo \"Failed to send phone home request\"\n        exit 1\n      fi\n    "
      },
      "dependsOn": [
        "[resourceId('Microsoft.Network/virtualNetworks', format('{0}-vnet', parameters('nuonInstallID')))]"
      ]
    },
    {
      "type": "Microsoft.Resources/deployments",
      "apiVersion": "2022-09-01",
      "name": "custom-role-deployment",
      "subscriptionId": "[subscription().subscriptionId]",
      "location": "[resourceGroup().location]",
      "properties": {
        "expressionEvaluationOptions": {
          "scope": "inner"
        },
        "mode": "Incremental",
        "parameters": {
          "nuonInstallID": {
            "value": "[parameters('nuonInstallID')]"
          },
          "principalID": {
            "value": "[reference(resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID'))), '2023-03-01', 'full').identity.principalId]"
          }
        },
        "template": {
          "$schema": "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
          "contentVersion": "1.0.0.0",
          "metadata": {
            "_generator": {
              "name": "bicep",
              "version": "0.36.1.42791",
              "templateHash": "3348984103542313038"
            }
          },
          "parameters": {
            "nuonInstallID": {
              "type": "string"
            },
            "principalID": {
              "type": "string"
            }
          },
          "resources": [
            {
              "type": "Microsoft.Authorization/roleDefinitions",
              "apiVersion": "2022-04-01",
              "name": "[guid(subscription().id, format('{0}-runner-resource-provider-register-role', parameters('nuonInstallID')))]",
              "properties": {
                "roleName": "[format('{0}-runner-resource-provider-register-role', parameters('nuonInstallID'))]",
                "description": "Custom role to allow assuming other trusted roles",
                "assignableScopes": [
                  "[subscription().id]"
                ],
                "permissions": [
                  {
                    "actions": [
                      "*/register/action"
                    ],
                    "notActions": [],
                    "dataActions": [],
                    "notDataActions": []
                  }
                ]
              }
            },
            {
              "type": "Microsoft.Authorization/roleAssignments",
              "apiVersion": "2022-04-01",
              "name": "[guid(subscription().id, parameters('principalID'), 'CustomRunnerRole')]",
              "properties": {
                "roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, format('{0}-runner-resource-provider-register-role', parameters('nuonInstallID'))))]",
                "principalId": "[parameters('principalID')]",
                "principalType": "ServicePrincipal"
              },
              "dependsOn": [
                "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, format('{0}-runner-resource-provider-register-role', parameters('nuonInstallID'))))]"
              ]
            }
          ]
        }
      },
      "dependsOn": [
        "[resourceId('Microsoft.Compute/virtualMachineScaleSets', format('{0}-vmss', parameters('nuonInstallID')))]"
      ]
    }
  ],
  "outputs": {
    "vnetId": {
      "type": "string",
      "value": "[resourceId('Microsoft.Network/virtualNetworks', format('{0}-vnet', parameters('nuonInstallID')))]"
    },
    "vnetName": {
      "type": "string",
      "value": "[format('{0}-vnet', parameters('nuonInstallID'))]"
    },
    "publicSubnet1Id": {
      "type": "string",
      "value": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone1', parameters('nuonInstallID')))]"
    },
    "publicSubnet1Name": {
      "type": "string",
      "value": "[format('{0}-public-subnet-zone1', parameters('nuonInstallID'))]"
    },
    "publicSubnet2Id": {
      "type": "string",
      "value": "[if(variables('createPublicSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone2', parameters('nuonInstallID'))), '')]"
    },
    "publicSubnet2Name": {
      "type": "string",
      "value": "[if(variables('createPublicSubnet2'), format('{0}-public-subnet-zone2', parameters('nuonInstallID')), '')]"
    },
    "publicSubnet3Id": {
      "type": "string",
      "value": "[if(variables('createPublicSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone3', parameters('nuonInstallID'))), '')]"
    },
    "publicSubnet3Name": {
      "type": "string",
      "value": "[if(variables('createPublicSubnet3'), format('{0}-public-subnet-zone3', parameters('nuonInstallID')), '')]"
    },
    "privateSubnet1Id": {
      "type": "string",
      "value": "[resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone1', parameters('nuonInstallID')))]"
    },
    "privateSubnet1Name": {
      "type": "string",
      "value": "[format('{0}-private-subnet-zone1', parameters('nuonInstallID'))]"
    },
    "privateSubnet2Id": {
      "type": "string",
      "value": "[if(variables('createPrivateSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone2', parameters('nuonInstallID'))), '')]"
    },
    "privateSubnet2Name": {
      "type": "string",
      "value": "[if(variables('createPrivateSubnet2'), format('{0}-private-subnet-zone2', parameters('nuonInstallID')), '')]"
    },
    "privateSubnet3Id": {
      "type": "string",
      "value": "[if(variables('createPrivateSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone3', parameters('nuonInstallID'))), '')]"
    },
    "privateSubnet3Name": {
      "type": "string",
      "value": "[if(variables('createPrivateSubnet3'), format('{0}-private-subnet-zone3', parameters('nuonInstallID')), '')]"
    },
    "publicSubnetIds": {
      "type": "string",
      "value": "[join(filter(createArray(resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone1', parameters('nuonInstallID'))), if(variables('createPublicSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone2', parameters('nuonInstallID'))), ''), if(variables('createPublicSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-public-subnet-zone3', parameters('nuonInstallID'))), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"
    },
    "publicSubnetNames": {
      "type": "string",
      "value": "[join(filter(createArray(format('{0}-public-subnet-zone1', parameters('nuonInstallID')), if(variables('createPublicSubnet2'), format('{0}-public-subnet-zone2', parameters('nuonInstallID')), ''), if(variables('createPublicSubnet3'), format('{0}-public-subnet-zone3', parameters('nuonInstallID')), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"
    },
    "privateSubnetIds": {
      "type": "string",
      "value": "[join(filter(createArray(resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone1', parameters('nuonInstallID'))), if(variables('createPrivateSubnet2'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone2', parameters('nuonInstallID'))), ''), if(variables('createPrivateSubnet3'), resourceId('Microsoft.Network/virtualNetworks/subnets', format('{0}-vnet', parameters('nuonInstallID')), format('{0}-private-subnet-zone3', parameters('nuonInstallID'))), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"
    },
    "privateSubnetNames": {
      "type": "string",
      "value": "[join(filter(createArray(format('{0}-private-subnet-zone1', parameters('nuonInstallID')), if(variables('createPrivateSubnet2'), format('{0}-private-subnet-zone2', parameters('nuonInstallID')), ''), if(variables('createPrivateSubnet3'), format('{0}-private-subnet-zone3', parameters('nuonInstallID')), '')), lambda('x', not(empty(lambdaVariables('x'))))), ',')]"
    }
  }
}
`

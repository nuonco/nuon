
@description('The Nuon Install ID; prefixed to resource names.')
param nuonInstallID string = '{{.Install.ID}}'

@description('The Nuon Org ID. Used in tags.')
param nuonOrgID string = '{{.Runner.OrgID}}'

@description('The Nuon App ID. Used in tags.')
param nuonAppID string = '{{.Install.AppID}}'

@description('Please enter the IP range (CIDR notation) for this VNet.')
param vnetCIDR string = '10.128.0.0/16'

@description('Please enter the IP range (CIDR notation) for the public subnet')
param publicSubnet1CIDR string = '10.128.0.0/26'

@description('Please enter the IP range (CIDR notation) for the public subnet in the second zone (optional)')
param publicSubnet2CIDR string = '10.128.0.64/26'

@description('Please enter the IP range (CIDR notation) for the public subnet in the third zone (optional)')
param publicSubnet3CIDR string = '10.128.0.128/26'

@description('Please enter the IP range (CIDR notation) for the dedicated private subnet for the runner.')
param runnerSubnetCIDR string = '10.128.128.0/24'

@description('Please enter the IP range (CIDR notation) for the private subnet')
param privateSubnet1CIDR string = '10.128.130.0/24'

@description('Please enter the IP range (CIDR notation) for the private subnet in the second zone (optional)')
param privateSubnet2CIDR string = '10.128.132.0/24'

@description('Please enter the IP range (CIDR notation) for the private subnet in the third zone (optional)')
param privateSubnet3CIDR string = '10.128.134.0/24'

@description('The location for all resources.')
param location string = '{{.Install.AzureAccount.Location}}'

@description('Admin username for the VM')
param vmAdminUsername string = 'nuon_admin'

@description('Admin password for the VM')
@secure()
param vmAdminPassword string = 'Password!123'

@description('List of secrets to store in Azure Key Vault')
param secrets array = []

var commonTags = {
  install_nuon_co_id: nuonInstallID
  org_nuon_co_id: nuonOrgID
  app_nuon_co_id: nuonAppID
}

var createPublicSubnet2 = !empty(publicSubnet2CIDR)
var createPublicSubnet3 = !empty(publicSubnet3CIDR) 
var createPrivateSubnet2 = !empty(privateSubnet2CIDR)
var createPrivateSubnet3 = !empty(privateSubnet3CIDR)

resource publicNsg 'Microsoft.Network/networkSecurityGroups@2023-04-01' = {
  name: '${nuonInstallID}-public-nsg'
  location: location
  tags: commonTags
  properties: {
    securityRules: [
      {
        name: 'Allow-All-Inbound'
        properties: {
          description: 'Allow all inbound traffic from any source'
          protocol: '*'
          sourcePortRange: '*'
          destinationPortRange: '*'
          sourceAddressPrefix: '*'
          destinationAddressPrefix: '*'
          access: 'Allow'
          priority: 200
          direction: 'Inbound'
        }
      }
    ]
  }
}

resource privateNsg 'Microsoft.Network/networkSecurityGroups@2023-04-01' = {
  name: '${nuonInstallID}-private-nsg'
  location: location
  tags: commonTags
  properties: {
    securityRules: []
  }
}

resource privateRouteTable 'Microsoft.Network/routeTables@2023-04-01' = {
  name: '${nuonInstallID}-private-routetable'
  location: location
  tags: commonTags
  properties: {
    disableBgpRoutePropagation: false
  }
}

resource vnet 'Microsoft.Network/virtualNetworks@2023-04-01' = {
  name: '${nuonInstallID}-vnet'
  location: location
  tags: commonTags
  properties: {
    addressSpace: {
      addressPrefixes: [
        vnetCIDR
      ]
    }
    subnets: [
      {
        name: '${nuonInstallID}-public-subnet-zone1'
        properties: {
          addressPrefix: publicSubnet1CIDR
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
          networkSecurityGroup: {
            id: publicNsg.id
          }
          natGateway: {
            id: natGateway.id
          }
        }
      }
      {
        name: '${nuonInstallID}-public-subnet-zone2'
        properties: {
          addressPrefix: publicSubnet2CIDR
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
          networkSecurityGroup: {
            id: publicNsg.id
          }
          natGateway: {
            id: natGateway.id
          }
        }
      }
      {
        name: '${nuonInstallID}-public-subnet-zone3'
        properties: {
          addressPrefix: publicSubnet3CIDR
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
          networkSecurityGroup: {
            id: publicNsg.id
          }
          natGateway: {
            id: natGateway.id
          }
        }
      }
      {
        name: '${nuonInstallID}-private-runner-subnet'
        properties: {
          addressPrefix: runnerSubnetCIDR
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
          networkSecurityGroup: {
            id: privateNsg.id
          }
          natGateway: {
            id: natGateway.id
          }
          serviceEndpoints: [
            {
              service: 'Microsoft.KeyVault'
            }
            {
              service: 'Microsoft.ContainerRegistry'
            }
          ]
        }
      }
      {
        name: '${nuonInstallID}-private-subnet-zone1'
        properties: {
          addressPrefix: privateSubnet1CIDR
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
          networkSecurityGroup: {
            id: privateNsg.id
          }
          natGateway: {
            id: natGateway.id
          }
          serviceEndpoints: [
            {
              service: 'Microsoft.KeyVault'
            }
            {
              service: 'Microsoft.ContainerRegistry'
            }
          ]
        }
      }
      {
        name: '${nuonInstallID}-private-subnet-zone2'
        properties: {
          addressPrefix: privateSubnet2CIDR
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
          networkSecurityGroup: {
            id: privateNsg.id
          }
          natGateway: {
            id: natGateway.id
          }
          serviceEndpoints: [
            {
              service: 'Microsoft.KeyVault'
            }
            {
              service: 'Microsoft.ContainerRegistry'
            }
          ]
        }
      }
      {
        name: '${nuonInstallID}-private-subnet-zone3'
        properties: {
          addressPrefix: privateSubnet3CIDR
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
          networkSecurityGroup: {
            id: privateNsg.id
          }
          natGateway: {
            id: natGateway.id
          }
          serviceEndpoints: [
            {
              service: 'Microsoft.KeyVault'
            }
            {
              service: 'Microsoft.ContainerRegistry'
            }
          ]
        }
      }
    ]
  }
}

resource keyVault 'Microsoft.KeyVault/vaults@2023-02-01' = {
  name: take('${nuonInstallID}kv', 24)
  location: location
  tags: commonTags
  properties: {
    enabledForDeployment: true
    enabledForTemplateDeployment: true
    enabledForDiskEncryption: true
    tenantId: subscription().tenantId
    accessPolicies: []
    sku: {
      name: 'standard'
      family: 'A'
    }
    networkAcls: {
      defaultAction: 'Deny'
      bypass: 'AzureServices'
      ipRules: []
      virtualNetworkRules: [
        {
          id: resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-runner-subnet')
        }
        {
          id: resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone1')
        }
      ]
    }
  }
}

resource keyVaultSecrets 'Microsoft.KeyVault/vaults/secrets@2023-02-01' = [for secret in secrets: {
  name: '${keyVault.name}/${secret.name}'
  properties: {
    value: secret.value
    contentType: contains(secret, 'contentType') ? secret.contentType : null
    attributes: {
      enabled: contains(secret, 'enabled') ? secret.enabled : true
    }
  }
}]


resource natGatewayPublicIP 'Microsoft.Network/publicIPAddresses@2023-04-01' = {
  name: '${nuonInstallID}-natgw-pip'
  location: location
  tags: commonTags
  sku: {
    name: 'Standard'
  }
  properties: {
    publicIPAllocationMethod: 'Static'
  }
}

resource natGateway 'Microsoft.Network/natGateways@2023-04-01' = {
  name: '${nuonInstallID}-natgw'
  location: location
  tags: commonTags
  sku: {
    name: 'Standard'
  }
  properties: {
    publicIpAddresses: [
      {
        id: natGatewayPublicIP.id
      }
    ]
    idleTimeoutInMinutes: 4
  }
}

var customData = '''
#!/bin/bash

RUNNER_ID={{.Runner.ID}}
RUNNER_API_TOKEN={{.APIToken}}
RUNNER_API_URL={{.Settings.RunnerAPIURL}}
AWS_REGION={{.Install.AzureAccount.Location}}

# Remove any existing Docker packages
apt-get remove -y docker docker-engine docker.io containerd runc

# Update package index and install prerequisites
apt-get update
apt-get install -y ca-certificates curl gnupg lsb-release

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
RUNNER_API_TOKEN=$RUNNER_API_TOKEN
RUNNER_API_URL=$RUNNER_API_URL
ARM_USE_MSI=true
# FIXME(sdboyer) this hack must be fixed - userdata is only run on instance creation, and ip can change on each boot
HOST_IP=$(curl -s https://checkip.amazonaws.com)
EOF

# this ⤵ is wrapped w/ single quotes to prevent variable expansion.
cat << 'EOF' > /opt/nuon/runner/get_image_tag.sh
#!/bin/bash

set -u

# source this file to get some env vars
. /opt/nuon/runner/env

# Fetch runner settings from the API
echo "Fetching runner settings from $RUNNER_API_URL/v1/runners/$RUNNER_ID/settings"
RUNNER_SETTINGS=$(curl -s -H "Authorization: Bearer $RUNNER_API_TOKEN" "$RUNNER_API_URL/v1/runners/$RUNNER_ID/settings")

# Extract container image URL and tag from the response
CONTAINER_IMAGE_URL=$(echo "$RUNNER_SETTINGS" | grep -o '"container_image_url":"[^"]*"' | cut -d '"' -f 4)
CONTAINER_IMAGE_TAG=$(echo "$RUNNER_SETTINGS" | grep -o '"container_image_tag":"[^"]*"' | cut -d '"' -f 4)

# echo into a file for easier retrieval; re-create the file to avoid duplicate values.
rm -f /opt/nuon/runner/image
echo "CONTAINER_IMAGE_URL=$CONTAINER_IMAGE_URL" >> /opt/nuon/runner/image
echo "CONTAINER_IMAGE_TAG=$CONTAINER_IMAGE_TAG" >> /opt/nuon/runner/image

# export so we can get these values by sourcing this file
export CONTAINER_IMAGE_URL=$CONTAINER_IMAGE_URL
export CONTAINER_IMAGE_TAG=$CONTAINER_IMAGE_TAG

echo "Using container image: $CONTAINER_IMAGE_URL:$CONTAINER_IMAGE_TAG"
EOF

chmod +x /opt/nuon/runner/get_image_tag.sh
/opt/nuon/runner/get_image_tag.sh

# Create systemd unit file for runner
cat << 'EOF' > /etc/systemd/system/nuon-runner.service
[Unit]
Description=Nuon Runner Service
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
User=runner
ExecStartPre=-/bin/sh -c '/usr/bin/docker stop $(/usr/bin/docker ps -a -q --filter="name=%n")'
ExecStartPre=-/bin/sh -c '/usr/bin/docker rm $(/usr/bin/docker ps -a -q --filter="name=%n")'
ExecStartPre=-/bin/sh -c "yes | /usr/bin/docker system prune"
ExecStartPre=-/bin/sh /opt/nuon/runner/get_image_tag.sh
EnvironmentFile=/opt/nuon/runner/image
EnvironmentFile=/opt/nuon/runner/env
ExecStartPre=echo "Using container image: ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG}"
ExecStartPre=/usr/bin/docker pull ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG}
ExecStart=/usr/bin/docker run --network host -v /tmp/nuon-runner:/tmp --rm --name %n -p 5000:5000 --memory "3750g" --cpus="1.75" --env-file /opt/nuon/runner/env ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG} run
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
EOF

# Enable SSH password authentication
sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config
systemctl restart sshd

# Reload systemd and start the service (no SELinux on Ubuntu)
systemctl daemon-reload
systemctl enable --now nuon-runner
'''

// Virtual Machine Scale Set
resource vmss 'Microsoft.Compute/virtualMachineScaleSets@2023-03-01' = {
  name: '${nuonInstallID}-vmss'
  location: location
  tags: commonTags
  sku: {
    name: 'Standard_B2s'
    tier: 'Standard'
    capacity: 1
  }
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    upgradePolicy: {
      mode: 'Manual'
    }
    virtualMachineProfile: {
      osProfile: {
        computerNamePrefix: nuonInstallID
        adminUsername: vmAdminUsername
        adminPassword: vmAdminPassword
        customData: base64(customData)
        linuxConfiguration: {
          disablePasswordAuthentication: false
        }
      }
      storageProfile: {
        imageReference: {
          publisher: 'Canonical'
          offer: '0001-com-ubuntu-server-jammy'
          sku: '22_04-lts-gen2'
          version: 'latest'
        }
        osDisk: {
          createOption: 'FromImage'
          managedDisk: {
            storageAccountType: 'Standard_LRS'
          }
        }
      }
      networkProfile: {
        networkInterfaceConfigurations: [
          {
            name: 'nic'
            properties: {
              primary: true
              ipConfigurations: [
                {
                  name: 'ipconfig'
                  properties: {
                    subnet: {
                      id: resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-runner-subnet')
                    }
                  }
                }
              ]
            }
          }
        ]
      }
    }
  }
}

resource vmssContributorRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, vmss.id, 'Contributor')
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b24988ac-6180-42a0-ab88-20f7382dd24c')
    principalId: vmss.identity.principalId
    principalType: 'ServicePrincipal'
  }
}

resource vmssRbacAdminRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, vmss.id, 'RoleBasedAccessControlAdministrator')
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'f58310d9-a9f6-439a-9e8d-f62e7b41a168')
    principalId: vmss.identity.principalId
    principalType: 'ServicePrincipal'
  }
}

resource vmssAksRbacClusterAdminRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, vmss.id, 'AzureKubernetesServiceRBACClusterAdmin')
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b1ff04bb-8a4e-4dc4-8eb5-8693973ce19b')
    principalId: vmss.identity.principalId
    principalType: 'ServicePrincipal'
  }
}

module customRoleModule 'custom-role.bicep' = {
  name: 'custom-role-deployment'
  scope: subscription()
  params: {
    nuonInstallID: nuonInstallID
    principalID: vmss.identity.principalId
  }
}

resource phoneHomeScript 'Microsoft.Resources/deploymentScripts@2023-08-01' = {
  name: '${nuonInstallID}-phone-home-script'
  location: location
  tags: commonTags
  kind: 'AzureCLI'
  properties: {
    azCliVersion: '2.40.0'
    timeout: 'PT30M'
    retentionInterval: 'PT1H'
    environmentVariables: [
      {
        name: 'SUBSCRIPTION_ID'
        value: subscription().subscriptionId
      }
      {
        name: 'SUBSCRIPTION_TENANT_ID'
        value: subscription().tenantId
      }
      {
        name: 'RESOURCE_GROUP_ID'
        value: resourceGroup().id
      }
      {
        name: 'RESOURCE_GROUP_NAME'
        value: resourceGroup().name
      }
      {
        name: 'RESOURCE_GROUP_LOCATION'
        value: resourceGroup().location
      }
      {
        name: 'VNET_ID'
        value: vnet.id
      }
      {
        name: 'VNET_NAME'
        value: vnet.name
      }
      {
        name: 'PUBLIC_SUBNET_1_ID'
        value: resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone1')
      }
      {
        name: 'PUBLIC_SUBNET_1_NAME'
        value: '${nuonInstallID}-public-subnet-zone1'
      }
      {
        name: 'PUBLIC_SUBNET_2_ID'
        value: createPublicSubnet2 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone2') : ''
      }
      {
        name: 'PUBLIC_SUBNET_2_NAME'
        value: createPublicSubnet2 ? '${nuonInstallID}-public-subnet-zone2' : ''
      }
      {
        name: 'PUBLIC_SUBNET_3_ID'
        value: createPublicSubnet3 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone3') : ''
      }
      {
        name: 'PUBLIC_SUBNET_3_NAME'
        value: createPublicSubnet3 ? '${nuonInstallID}-public-subnet-zone3' : ''
      }
      {
        name: 'PRIVATE_SUBNET_1_ID'
        value: resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone1')
      }
      {
        name: 'PRIVATE_SUBNET_1_NAME'
        value: '${nuonInstallID}-private-subnet-zone1'
      }
      {
        name: 'PRIVATE_SUBNET_2_ID'
        value: createPrivateSubnet2 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone2') : ''
      }
      {
        name: 'PRIVATE_SUBNET_2_NAME'
        value: createPrivateSubnet2 ? '${nuonInstallID}-private-subnet-zone2' : ''
      }
      {
        name: 'PRIVATE_SUBNET_3_ID'
        value: createPrivateSubnet3 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone3') : ''
      }
      {
        name: 'PRIVATE_SUBNET_3_NAME'
        value: createPrivateSubnet3 ? '${nuonInstallID}-private-subnet-zone3' : ''
      }
      {
        name: 'PUBLIC_SUBNET_IDS_CSV'
        value: join(filter([
          resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone1')
          createPublicSubnet2 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone2') : ''
          createPublicSubnet3 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone3') : ''
        ], x => !empty(x)), ',')
      }
      {
        name: 'PUBLIC_SUBNET_NAMES_CSV'
        value: join(filter([
          '${nuonInstallID}-public-subnet-zone1'
          createPublicSubnet2 ? '${nuonInstallID}-public-subnet-zone2' : ''
          createPublicSubnet3 ? '${nuonInstallID}-public-subnet-zone3' : ''
        ], x => !empty(x)), ',')
      }
      {
        name: 'PRIVATE_SUBNET_IDS_CSV'
        value: join(filter([
          resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone1')
          createPrivateSubnet2 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone2') : ''
          createPrivateSubnet3 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone3') : ''
        ], x => !empty(x)), ',')
      }
      {
        name: 'PRIVATE_SUBNET_NAMES_CSV'
        value: join(filter([
          '${nuonInstallID}-private-subnet-zone1'
          createPrivateSubnet2 ? '${nuonInstallID}-private-subnet-zone2' : ''
          createPrivateSubnet3 ? '${nuonInstallID}-private-subnet-zone3' : ''
        ], x => !empty(x)), ',')
      }
    ]
    scriptContent: '''
      #!/bin/bash
      
      # Construct the JSON payload with stack outputs
      #
      # Including the credentails object for backwards compatibility.
      # We used to need this when the org runner did the sandbox provision,
      # but the independent runner obviates the need for this.
      #
      # The provision workflow still looks for auth credentials,
      # because it needs the role ARNs to use for different jobs.
      # Azure resource groups obviate the need for multiple roles,
      # so we don't need to return anything.

      # Create arrays for public and private subnets (filtering out empty values)
      PUBLIC_SUBNETS=("$PUBLIC_SUBNET_1_ID")
      PUBLIC_SUBNET_NAMES=("$PUBLIC_SUBNET_1_NAME")
      if [ -n "$PUBLIC_SUBNET_2_ID" ]; then 
        PUBLIC_SUBNETS+=("$PUBLIC_SUBNET_2_ID")
        PUBLIC_SUBNET_NAMES+=("$PUBLIC_SUBNET_2_NAME")
      fi
      if [ -n "$PUBLIC_SUBNET_3_ID" ]; then 
        PUBLIC_SUBNETS+=("$PUBLIC_SUBNET_3_ID")
        PUBLIC_SUBNET_NAMES+=("$PUBLIC_SUBNET_3_NAME")
      fi

      PRIVATE_SUBNETS=("$PRIVATE_SUBNET_1_ID")
      PRIVATE_SUBNET_NAMES=("$PRIVATE_SUBNET_1_NAME")
      if [ -n "$PRIVATE_SUBNET_2_ID" ]; then 
        PRIVATE_SUBNETS+=("$PRIVATE_SUBNET_2_ID")
        PRIVATE_SUBNET_NAMES+=("$PRIVATE_SUBNET_2_NAME")
      fi
      if [ -n "$PRIVATE_SUBNET_3_ID" ]; then 
        PRIVATE_SUBNETS+=("$PRIVATE_SUBNET_3_ID")
        PRIVATE_SUBNET_NAMES+=("$PRIVATE_SUBNET_3_NAME")
      fi

      PAYLOAD=$(cat << EOF
{
  "resource_group_id": "$RESOURCE_GROUP_ID",
  "resource_group_name": "$RESOURCE_GROUP_NAME",
  "resource_group_location": "$RESOURCE_GROUP_LOCATION",
  "network_id": "$VNET_ID",
  "network_name": "$VNET_NAME",
  "public_subnet_ids": "$PUBLIC_SUBNET_IDS_CSV",
  "public_subnet_names": "$PUBLIC_SUBNET_NAMES_CSV",
  "private_subnet_ids": "$PRIVATE_SUBNET_IDS_CSV",
  "private_subnet_names": "$PRIVATE_SUBNET_NAMES_CSV",
  "subscription_id": "$SUBSCRIPTION_ID",
  "subscription_tenant_id": "$SUBSCRIPTION_TENANT_ID"
}
EOF
)
      
      # Send the phone home request
      curl -X POST \
        "{{.CloudFormationStackVersion.PhoneHomeURL}}" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -d "$PAYLOAD" \
        --fail \
        --silent \
        --show-error
      
      if [ $? -eq 0 ]; then
        echo "Phone home request sent successfully"
      else
        echo "Failed to send phone home request"
        exit 1
      fi
    '''
  }
  dependsOn: [
    vnet
  ]
}

// Outputs
output vnetId string = vnet.id
output vnetName string = vnet.name
output publicSubnet1Id string = resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone1')
output publicSubnet1Name string = '${nuonInstallID}-public-subnet-zone1'
output publicSubnet2Id string = createPublicSubnet2 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone2') : ''
output publicSubnet2Name string = createPublicSubnet2 ? '${nuonInstallID}-public-subnet-zone2' : ''
output publicSubnet3Id string = createPublicSubnet3 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone3') : ''
output publicSubnet3Name string = createPublicSubnet3 ? '${nuonInstallID}-public-subnet-zone3' : ''
output privateSubnet1Id string = resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone1')
output privateSubnet1Name string = '${nuonInstallID}-private-subnet-zone1'
output privateSubnet2Id string = createPrivateSubnet2 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone2') : ''
output privateSubnet2Name string = createPrivateSubnet2 ? '${nuonInstallID}-private-subnet-zone2' : ''
output privateSubnet3Id string = createPrivateSubnet3 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone3') : ''
output privateSubnet3Name string = createPrivateSubnet3 ? '${nuonInstallID}-private-subnet-zone3' : ''

// Comma-separated list outputs
output publicSubnetIds string = join(filter([
  resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone1')
  createPublicSubnet2 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone2') : ''
  createPublicSubnet3 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-public-subnet-zone3') : ''
], x => !empty(x)), ',')

output publicSubnetNames string = join(filter([
  '${nuonInstallID}-public-subnet-zone1'
  createPublicSubnet2 ? '${nuonInstallID}-public-subnet-zone2' : ''
  createPublicSubnet3 ? '${nuonInstallID}-public-subnet-zone3' : ''
], x => !empty(x)), ',')

output privateSubnetIds string = join(filter([
  resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone1')
  createPrivateSubnet2 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone2') : ''
  createPrivateSubnet3 ? resourceId('Microsoft.Network/virtualNetworks/subnets', vnet.name, '${nuonInstallID}-private-subnet-zone3') : ''
], x => !empty(x)), ',')

output privateSubnetNames string = join(filter([
  '${nuonInstallID}-private-subnet-zone1'
  createPrivateSubnet2 ? '${nuonInstallID}-private-subnet-zone2' : ''
  createPrivateSubnet3 ? '${nuonInstallID}-private-subnet-zone3' : ''
], x => !empty(x)), ',')

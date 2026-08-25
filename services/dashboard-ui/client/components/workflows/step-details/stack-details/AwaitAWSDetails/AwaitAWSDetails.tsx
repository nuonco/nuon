import { useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Expand } from '@/components/common/Expand'
import { Link } from '@/components/common/Link'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { CreateOIDCTrustPolicyButton } from '@/components/oidc-trust-policies'
import { CreateServiceAccountTokenModalContainer } from '@/components/service-accounts/ServiceAccountToken'
import { useConfig } from '@/hooks/use-config'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOIDCTrustPolicies } from '@/hooks/use-oidc-trust-policies'
import { useStackServiceAccount } from '@/hooks/use-stack-service-account'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from '../types'

interface IAwaitAWSDetails extends IStackDetails {
  orgId: string
  installId?: string
  installAwsRegion?: string
  tfProvider?: boolean
}

const telemetryExportConfigFilename = 'telemetry-export-config.yaml'
const telemetryExportConfig = `version: v1

telemetry:
  logs:
    audit:
      enabled: true

exporters:
  otlphttp:
    endpoint: https://otlp.example.com
    headers:
      Authorization: Bearer <token>`

// The tfvars envelope ctl-api stores in `terraform_contents` is a JSON
// document of shape `{"inputs_tfvars": "<hcl>", "secrets_tfvars": "<hcl>"}`.
// Mirrors the GCP parser at AwaitGCPDetails.tsx.
interface TfvarsEnvelope {
  inputs: string
  providerInputs: string
  secrets: string
}

function parseTfvars(contents: unknown): TfvarsEnvelope {
  const empty: TfvarsEnvelope = { inputs: '', providerInputs: '', secrets: '' }
  if (!contents) return empty

  let raw: unknown = contents
  if (typeof raw === 'string') {
    try {
      raw = JSON.parse(raw)
    } catch {
      try {
        raw = JSON.parse(atob(raw as string))
      } catch {
        return empty
      }
    }
  }

  if (typeof raw === 'object' && raw !== null) {
    const rec = raw as Record<string, unknown>
    const inputs = String(rec.inputs_tfvars ?? '')
    const secrets = String(rec.secrets_tfvars ?? '')
    const legacy = String(rec.tfvars ?? '')
    return {
      inputs: inputs || (secrets ? '' : legacy),
      providerInputs: String(rec.provider_tfvars ?? ''),
      secrets,
    }
  }

  return empty
}

export const AwaitAWSDetails = ({
  stack,
  orgId,
  installId,
  installAwsRegion,
  tfProvider = false,
  loading,
}: IAwaitAWSDetails) => {
  const version = stack?.versions?.at(0)
  // The new TerraformContents fields aren't in the regenerated OpenAPI types
  // yet; bridge with a local widening cast.
  const versionExt = version as
    | (typeof version & {
        terraform_contents?: unknown
        terraform_checksum?: string
      })
    | undefined

  const tfvars = useMemo(
    () => parseTfvars(versionExt?.terraform_contents),
    [versionExt?.terraform_contents]
  )
  const hasTerraform = tfvars.inputs.length > 0 || tfvars.secrets.length > 0

  if (loading) {
    return (
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Setup your install stack
        </Text>
        <Card>
          <Text weight="strong">CloudFormation template</Text>
          <Code loading />
        </Card>
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Deploy with AWS CLI
          </Text>
          <Card>
            <Text>Create stack</Text>
            <Code loading />
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Setup your install stack
      </Text>

      {hasTerraform ? (
        <Tabs
          initActiveTab={version?.template_url ? 'cloudformation' : 'terraform'}
          tabLabels={{
            cloudformation: 'CloudFormation',
            tfmodule: 'TF Module',
          }}
          tabs={{
            cloudformation: (
              <CloudFormationTab
                version={version}
                installId={installId}
                installAwsRegion={installAwsRegion}
              />
            ),
            terraform: (
              <TerraformTab
                inputsTfvars={tfvars.inputs}
                installAwsRegion={installAwsRegion}
                secretsTfvars={tfvars.secrets}
                installId={installId}
              />
            ),
            // Gated on the org feature until the module and provider releases it
            // depends on are published.
            ...(tfProvider
              ? {
                  tfmodule: (
                    <TFModuleTab
                      orgId={orgId}
                      installId={installId}
                      installAwsRegion={installAwsRegion}
                    />
                  ),
                }
              : {}),
          }}
        />
      ) : (
        <CloudFormationTab
          version={version}
          installId={installId}
          installAwsRegion={installAwsRegion}
        />
      )}
    </div>
  )
}

interface ICloudFormationTab {
  version: NonNullable<IStackDetails['stack']['versions']>[number] | undefined
  installId?: string
  installAwsRegion?: string
}

const CloudFormationTab = ({
  version,
  installId,
  installAwsRegion,
}: ICloudFormationTab) => {
  const templateUrl = version?.template_url

  if (!templateUrl) {
    return (
      <div className="flex flex-col gap-4 pt-4">
        <Card>
          <Text variant="subtext" theme="neutral">
            CloudFormation is unavailable for this install. Set up the stack
            template bucket to enable it.
          </Text>
        </Card>
      </div>
    )
  }

  const isS3Template =
    templateUrl?.includes('s3.amazonaws.com') || templateUrl?.includes('.s3.')
  const region =
    (version as { region?: string } | undefined)?.region ||
    version?.quick_link_url?.match(/region=([^&#]+)/)?.[1] ||
    installAwsRegion ||
    ''
  const stackName =
    version?.quick_link_url?.match(/stackName=([^&]+)/)?.[1] ||
    `nuon-${installId || 'install'}`
  // Prefer the server-supplied quick link; otherwise derive one from the
  // template URL so older stack versions (created before the backend change)
  // still surface the one-click launch.
  const quickLink =
    version?.quick_link_url ||
    (templateUrl
      ? region
        ? `https://${region}.console.aws.amazon.com/cloudformation/home?region=${region}#/stacks/quickcreate?templateUrl=${encodeURIComponent(templateUrl)}&stackName=${stackName}`
        : `https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?templateUrl=${encodeURIComponent(templateUrl)}&stackName=${stackName}`
      : '')
  // CLI commands and console links work whether the user already chose a
  // region or hasn't yet — when unknown, render a `<YOUR_REGION>` placeholder
  // and let the user substitute at run-time.
  const regionForCmd = region || '<YOUR_REGION>'
  const consoleUrl = region
    ? `https://console.aws.amazon.com/cloudformation/home?region=${region}#/stacks/events?filteringText=${stackName}&filteringStatus=active&viewNested=true`
    : `https://console.aws.amazon.com/cloudformation/home#/stacks?filteringText=${stackName}`

  return (
    <div className="flex flex-col gap-4 pt-4">
      {quickLink ? (
        <Card>
          <span className="flex justify-between items-center">
            <Text>Quick launch in AWS console</Text>
            <ClickToCopyButton textToCopy={quickLink} />
          </span>
          <Link href={quickLink} target="_blank" rel="noopener noreferrer">
            <Code>{quickLink}</Code>
          </Link>
        </Card>
      ) : (
        <Card>
          <Text variant="subtext" theme="neutral">
            Open the AWS console in your preferred region, then create a
            CloudFormation stack from the template URL below. The CLI snippets
            further down include a <code>&lt;YOUR_REGION&gt;</code> placeholder
            you can substitute.
          </Text>
        </Card>
      )}

      {templateUrl ? (
        <Card>
          <span className="flex justify-between items-center">
            <Text weight="strong">CloudFormation template</Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={templateUrl} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() => window.open(templateUrl, '_blank')}
              >
                Download
              </Button>
            </span>
          </span>
          <Link href={templateUrl} target="_blank" rel="noopener noreferrer">
            <Code>{templateUrl}</Code>
          </Link>
        </Card>
      ) : null}

      {quickLink ? <Divider dividerWord="or" /> : null}

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Deploy with AWS CLI
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Create stack</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={
                isS3Template
                  ? `aws cloudformation create-stack --stack-name ${stackName} --template-url ${templateUrl} --capabilities CAPABILITY_NAMED_IAM --region ${regionForCmd}`
                  : `curl -sLo template.json "${templateUrl}" && aws cloudformation create-stack --stack-name ${stackName} --template-body file://template.json --capabilities CAPABILITY_NAMED_IAM --region ${regionForCmd}`
              }
            />
          </span>
          <Code className="text-xs whitespace-pre-wrap break-all">
            {isS3Template
              ? `aws cloudformation create-stack \\\n  --stack-name ${stackName} \\\n  --template-url ${templateUrl} \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${regionForCmd}`
              : `curl -sLo template.json "${templateUrl}" \\\n  && aws cloudformation create-stack \\\n  --stack-name ${stackName} \\\n  --template-body file://template.json \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${regionForCmd}`}
          </Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Update existing stack</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={
                isS3Template
                  ? `aws cloudformation update-stack --stack-name ${stackName} --template-url ${templateUrl} --capabilities CAPABILITY_NAMED_IAM --region ${regionForCmd}`
                  : `curl -sLo template.json "${templateUrl}" && aws cloudformation update-stack --stack-name ${stackName} --template-body file://template.json --capabilities CAPABILITY_NAMED_IAM --region ${regionForCmd}`
              }
            />
          </span>
          <Code className="text-xs whitespace-pre-wrap break-all">
            {isS3Template
              ? `aws cloudformation update-stack \\\n  --stack-name ${stackName} \\\n  --template-url ${templateUrl} \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${regionForCmd}`
              : `curl -sLo template.json "${templateUrl}" \\\n  && aws cloudformation update-stack \\\n  --stack-name ${stackName} \\\n  --template-body file://template.json \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region ${regionForCmd}`}
          </Code>
        </Card>
      </div>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Verify your stack
        </Text>
        <Card>
          <Text variant="subtext" theme="neutral">
            After running the create or update command above, open the AWS
            CloudFormation console to monitor your stack progress.
          </Text>
          <Button
            variant="secondary"
            onClick={() => window.open(consoleUrl, '_blank')}
          >
            Open in AWS console
          </Button>
        </Card>
      </div>

      <AWSTelemetryExportInstructions
        installAwsRegion={regionForCmd}
        installId={installId}
      />
    </div>
  )
}

interface ITerraformTab {
  inputsTfvars: string
  installAwsRegion?: string
  secretsTfvars: string
  installId?: string
}

const TerraformTab = ({
  inputsTfvars,
  installAwsRegion,
  secretsTfvars,
  installId,
}: ITerraformTab) => {
  const inputsFile = inputsTfvars

  const cloneCmd = `git clone https://github.com/nuonco/install-stacks.git
cd install-stacks/aws`

  const backendSnippet = `terraform {
  backend "s3" {
    bucket = "<your-state-bucket>"
    key    = "nuon/${installId}/terraform.tfstate"
    region = "<your-state-bucket-region>"
  }
}`

  const applyCmd = `terraform init && terraform apply`

  return (
    <div className="flex flex-col gap-4 pt-4">
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Clone the install stack module
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Clone and enter the AWS module directory</Text>
            <ClickToCopyButton textToCopy={cloneCmd} />
          </span>
          <Code variant="preformated">{cloneCmd}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          2. Configure remote state (recommended)
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Create a <code>backend.tf</code> file to store Terraform state in
              S3
            </Text>
            <ClickToCopyButton textToCopy={backendSnippet} />
          </span>
          <Code variant="preformated">{backendSnippet}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Save the install configuration
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>inputs.auto.tfvars</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={inputsFile} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(inputsFile, 'inputs.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{inputsFile}</Code>
        </Card>
        {secretsTfvars ? (
          <Card>
            <span className="flex justify-between items-center">
              <Text>
                Save this as <code>secrets.auto.tfvars</code>
              </Text>
              <span className="flex gap-2 items-center">
                <ClickToCopyButton textToCopy={secretsTfvars} />
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() =>
                    createFileDownload(secretsTfvars, 'secrets.auto.tfvars')
                  }
                >
                  Download
                </Button>
              </span>
            </span>
            <Code variant="preformated">{secretsTfvars}</Code>
          </Card>
        ) : null}
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Apply with Terraform
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Run Terraform</Text>
            <ClickToCopyButton textToCopy={applyCmd} />
          </span>
          <Code variant="preformated">{applyCmd}</Code>
        </Card>
      </div>

      <AWSTelemetryExportInstructions
        installAwsRegion={installAwsRegion}
        installId={installId}
      />
    </div>
  )
}

const AWSTelemetryExportInstructions = ({
  installAwsRegion,
  installId,
}: {
  installAwsRegion?: string
  installId?: string
}) => {
  const secretID = `nuon/${installId || '<install-id>'}/telemetry-export-config`
  const uploadCmd = `aws secretsmanager put-secret-value \\
  --secret-id "${secretID}" \\
  --secret-string file://${telemetryExportConfigFilename} \\
  --region "${installAwsRegion || '<aws-region>'}"`

  return (
    <Expand
      id="telemetry-export"
      heading={
        <Text variant="base" weight="strong">
          Configure telemetry export (optional)
        </Text>
      }
    >
      <div className="flex flex-col gap-4 p-2">
        <Text variant="subtext">
          To export runner audit logs and other telemetry to your own backend,
          update the <code>telemetry-export-config</code> secret in AWS Secrets
          Manager after the stack is provisioned. See the{' '}
          <Link
            href="https://docs.nuon.co/guides/export-runner-audit-logs"
            isExternal
          >
            telemetry export reference
          </Link>{' '}
          for available settings.
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>{telemetryExportConfigFilename}</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={telemetryExportConfig} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(
                    telemetryExportConfig,
                    telemetryExportConfigFilename
                  )
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{telemetryExportConfig}</Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Update the telemetry export secret</Text>
            <ClickToCopyButton textToCopy={uploadCmd} />
          </span>
          <Code variant="preformated">{uploadCmd}</Code>
        </Card>
      </div>
    </Expand>
  )
}

interface ITFModuleTab {
  orgId: string
  installId?: string
  installAwsRegion?: string
}

const padTo = (name: string, width: number) => name.padEnd(width, ' ')

// `inputs = { ... }`, indented for the module block. The module's `inputs` and
// `secrets` maps only accept keys the app declares, so the snippet lists exactly the
// customer-facing ones. Empty when the app declares no customer inputs — the module
// treats an omitted map as "use control-plane values".
const buildInputsBlock = (
  inputs: Array<{ name?: string; default?: string }>
): string => {
  if (inputs.length === 0) return ''

  const nameWidth = Math.max(
    ...inputs.map((input) => (input.name ?? '').length)
  )
  const lines = inputs.map(
    (input) =>
      `    ${padTo(input.name ?? '', nameWidth)} = "${input.default ?? ''}"`
  )

  return `\n\n  inputs = {\n${lines.join('\n')}\n  }`
}

// `secrets = { ... }`. Auto-generated secrets are minted by the stack itself, so only
// the customer-supplied ones belong in the snippet. Values come from the root-level
// `variable` blocks, which Terraform populates from `TF_VAR_*`.
const buildSecretsBlock = (secrets: Array<{ name?: string }>): string => {
  if (secrets.length === 0) return ''

  const width = Math.max(...secrets.map((secret) => (secret.name ?? '').length))
  const lines = secrets.map(
    (secret) =>
      `    ${padTo(secret.name ?? '', width)} = { value = var.${secret.name ?? ''} }`
  )

  return `\n\n  secrets = {\n${lines.join('\n')}\n  }`
}

// Root-level `variable` blocks for each customer secret. Terraform fills them from
// `TF_VAR_<name>`, so no real value is ever written to main.tf.
const buildSecretVariablesBlock = (
  secrets: Array<{ name?: string; description?: string }>
): string => {
  if (secrets.length === 0) return ''

  const blocks = secrets.map((secret) => {
    const description = secret.description?.trim()
    // Widths match what `terraform fmt` would produce for the attributes present.
    const width = description ? 'description'.length : 'sensitive'.length
    const attrs = [
      `  ${padTo('type', width)} = string`,
      `  ${padTo('sensitive', width)} = true`,
      ...(description
        ? [`  ${padTo('description', width)} = "${description}"`]
        : []),
    ]
    return `variable "${secret.name ?? ''}" {\n${attrs.join('\n')}\n}`
  })

  return `\n\n${blocks.join('\n\n')}`
}

// `export TF_VAR_<name>='<placeholder>'` lines, shown alongside both auth methods
// since secrets are needed regardless of how the module authenticates.
const buildSecretExports = (secrets: Array<{ name?: string }>): string =>
  secrets
    .map(
      (secret) =>
        `export TF_VAR_${secret.name ?? ''}='<${(secret.name ?? '').replace(/_/g, '-')}-value>'`
    )
    .join('\n')

// OIDC auth is hidden until the experience is polished and fully tested; flip
// this to restore the Static token / OIDC toggle.
const OIDC_AUTH_ENABLED = false

// Directions for the published nuonco/stack/aws module, which reads its whole config
// from the API. Distinct from TerraformTab, which clones install-stacks and is driven
// by generated tfvars.
const TFModuleTab = ({ orgId, installId, installAwsRegion }: ITFModuleTab) => {
  const queryClient = useQueryClient()
  const { addModal } = useSurfaces()
  const [authMethod, setAuthMethod] = useState<'token' | 'oidc'>('token')
  // Only fetched once the OIDC pane is opened; most customers never need this list.
  const { data: trustPolicies } = useOIDCTrustPolicies({
    enabled: OIDC_AUTH_ENABLED && authMethod === 'oidc',
  })
  const {
    data: serviceAccount,
    isLoading,
    isError,
  } = useStackServiceAccount({
    installId,
    orgId,
    enabled: true,
  })

  // Defaults to a day rather than the modal's usual year: this credential is pasted
  // in by hand, so it is more exposed than one held by a service. Every other
  // duration is still selectable.
  const openCreateToken = () =>
    addModal(
      <CreateServiceAccountTokenModalContainer
        accountId={serviceAccount?.account_id ?? ''}
        identity={serviceAccount?.email ?? 'this install stack'}
        defaultDuration="24h"
        tokenName={`stack-${installId}`}
        onCreated={() =>
          queryClient.invalidateQueries({
            queryKey: ['stack-service-account', installId],
          })
        }
      />
    )

  // Derived from the timestamp rather than has_live_token, so the label stays honest
  // if only one of the two arrives.
  const liveUntil =
    serviceAccount?.has_live_token && serviceAccount.expires_at
      ? new Date(serviceAccount.expires_at).toLocaleString()
      : null

  const region = installAwsRegion ?? '<your-install-region>'

  // While the app config is still resolving, the snippet renders without the inputs
  // and secrets blocks rather than blocking the rest of the directions.
  const { appConfig } = useInstallAppConfig()

  const customerInputs = useMemo(() => {
    const declared = appConfig?.input?.inputs ?? []
    const grouped = (appConfig?.input?.input_groups ?? []).flatMap(
      (group) => group.app_inputs ?? []
    )
    const seen = new Set<string>()
    return [...declared, ...grouped].filter((input) => {
      if (!input.name || input.source !== 'customer') return false
      if (seen.has(input.name)) return false
      seen.add(input.name)
      return true
    })
  }, [appConfig?.input?.inputs, appConfig?.input?.input_groups])

  // Secrets carry no vendor/customer flag: everything not auto-generated is the
  // customer's to provide.
  const customerSecrets = useMemo(
    () =>
      (appConfig?.secrets?.secrets ?? []).filter(
        (secret) => !!secret.name && !secret.auto_generate
      ),
    [appConfig?.secrets?.secrets]
  )

  const inputsBlock = useMemo(
    () => buildInputsBlock(customerInputs),
    [customerInputs]
  )
  const secretsBlock = useMemo(
    () => buildSecretsBlock(customerSecrets),
    [customerSecrets]
  )
  const secretVariablesBlock = useMemo(
    () => buildSecretVariablesBlock(customerSecrets),
    [customerSecrets]
  )
  const secretExports = useMemo(
    () => buildSecretExports(customerSecrets),
    [customerSecrets]
  )

  const mainTf = `terraform {
  required_providers {
    aws   = { source = "hashicorp/aws" }
    stack = { source = "nuonco/stack" }
  }
}

provider "aws" {
  region = "${region}"
}

provider "stack" {}

module "aws_stack" {
  source  = "nuonco/stack/aws"
  version = "~> 1.0"

  install_id = "${installId ?? '<install-id>'}"${inputsBlock}${secretsBlock}
}${secretVariablesBlock}`

  // Always a placeholder: the token value is shown once, in the create modal.
  const authCmd = `export NUON_API_TOKEN='<api-token>'`
  const applyCmd = `terraform init && terraform apply`

  return (
    <div className="flex flex-col gap-4 pt-4">
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Create your Terraform configuration
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>main.tf</code>
            </Text>
            <ClickToCopyButton textToCopy={mainTf} />
          </span>
          <Code variant="preformated">{mainTf}</Code>
        </Card>
        {secretExports ? (
          <Card>
            <span className="flex justify-between items-center">
              <Text>Export the app&apos;s secret values.</Text>
              <ClickToCopyButton textToCopy={secretExports} />
            </span>
            <Code variant="preformated">{secretExports}</Code>
          </Card>
        ) : null}
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <span className="flex justify-between items-center gap-4">
          <Text variant="base" weight="strong">
            2. Authenticate
          </Text>
          {OIDC_AUTH_ENABLED ? (
            <ToggleButton<'token' | 'oidc'>
              value={authMethod}
              onChange={setAuthMethod}
              options={[
                { value: 'token', label: 'Static token' },
                { value: 'oidc', label: 'OIDC' },
              ]}
            />
          ) : null}
        </span>
        {OIDC_AUTH_ENABLED && authMethod === 'oidc' ? (
          <OIDCAuthPane
            installId={installId}
            policyNames={(trustPolicies ?? [])
              .map((policy) => policy.name ?? '')
              .filter(Boolean)}
          />
        ) : (
          <StaticTokenAuthPane
            authCmd={authCmd}
            isLoading={isLoading}
            isError={isError}
            liveUntil={liveUntil}
            canCreate={!!serviceAccount?.account_id}
            onCreateToken={openCreateToken}
          />
        )}
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Apply
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Initialize the module and create the stack</Text>
            <ClickToCopyButton textToCopy={applyCmd} />
          </span>
          <Code variant="preformated">{applyCmd}</Code>
        </Card>
      </div>
    </div>
  )
}

const StaticTokenAuthPane = ({
  authCmd,
  isLoading,
  isError,
  liveUntil,
  canCreate,
  onCreateToken,
}: {
  authCmd: string
  isLoading: boolean
  isError: boolean
  liveUntil: string | null
  canCreate: boolean
  onCreateToken: () => void
}) => {
  return (
    <>
      {isError ? (
        <Text variant="subtext" theme="neutral">
          This install stack has no service account yet. Reprovision the install
          to create one, then reload this page.
        </Text>
      ) : null}
      <Card>
        <span className="flex justify-between items-center">
          <Text>
            This token authorizes the module to read its configuration. Treat it
            as a secret.
          </Text>
          <ClickToCopyButton textToCopy={authCmd} />
        </span>
        <Code variant="preformated">{authCmd}</Code>
        {/* Nothing to offer until the service account resolves; the guidance above
              already explains why it might not. */}
        {isLoading || isError || !canCreate ? null : (
          <span className="flex justify-between items-center gap-4">
            <Text variant="subtext" theme="neutral">
              {liveUntil
                ? `A token is active until ${liveUntil}. Creating another does not revoke it unless you ask it to.`
                : 'No token yet. Create one to get its value — it is shown only once.'}
            </Text>
            <Button size="sm" variant="secondary" onClick={onCreateToken}>
              Create token
            </Button>
          </span>
        )}
      </Card>
    </>
  )
}

// The alternative to handing a customer a token at all: Actions mints an ID token per
// run and the control plane trades it for a short-lived Nuon token.
const OIDCAuthPane = ({
  installId,
  policyNames,
}: {
  installId?: string
  policyNames: string[]
}) => {
  const config = useConfig()
  const policyName = `stack-${installId ?? 'install'}`
  const existing = policyNames.includes(policyName)

  // The runner API, not the public one the OIDC settings page prefills: the SDK
  // requests its ID token with the URL it talks to and the control plane compares the
  // audience literally, so prefilling this means the two agree with no extra config.
  const audience = config.runnerApiUrl ?? ''

  const workflowSnippet = `permissions:
  id-token: write
  contents: read

env:
  NUON_ORG_ID: \${{ vars.NUON_ORG_ID }}`

  return (
    <>
      <Card>
        <span className="flex justify-between items-center gap-4">
          <Text>
            A trust policy tells Nuon which repository and branch may exchange
            an OIDC token for access to this org.
          </Text>
          {/* Without an audience the policy would be created with a blank one and
              reject every token, so this offers nothing rather than something
              broken. */}
          {audience ? (
            <CreateOIDCTrustPolicyButton
              variant="secondary"
              size="sm"
              lockPreset
              repoSource="manual"
              githubAudience={audience}
              defaultRole="org_admin"
              defaultName={policyName}
              reservedNames={policyNames}
            >
              {existing ? 'Create another' : 'Create trust policy'}
            </CreateOIDCTrustPolicyButton>
          ) : null}
        </span>
        {audience ? null : (
          <Text variant="subtext" theme="neutral">
            This control plane has not published its runner API URL, which the
            policy needs as its audience. Set <code>NUON_RUNNER_API_URL</code>{' '}
            on the dashboard and reload.
          </Text>
        )}
        {existing ? (
          <Text variant="subtext" theme="neutral">
            A policy named <code>{policyName}</code> already exists. Check it
            covers the repository and branch running this Terraform before
            creating another.
          </Text>
        ) : null}
      </Card>

      <Card>
        <span className="flex justify-between items-center">
          <Text>
            Add this to the workflow that applies the Terraform. No{' '}
            <code>NUON_API_TOKEN</code>.
          </Text>
          <ClickToCopyButton textToCopy={workflowSnippet} />
        </span>
        <Code variant="preformated">{workflowSnippet}</Code>
      </Card>

      <Text variant="subtext" theme="neutral">
        Without <code>id-token: write</code> the workflow cannot mint an ID
        token and the provider falls back to looking for a static token.
      </Text>
    </>
  )
}

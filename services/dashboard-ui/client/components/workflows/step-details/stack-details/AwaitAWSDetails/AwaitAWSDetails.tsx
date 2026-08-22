import { useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Expand } from '@/components/common/Expand'
import { Link } from '@/components/common/Link'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { useStackToken } from '@/hooks/use-stack-token'
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
            // Additive: the legacy terraform directions above are unchanged.
            // Gated on the org feature until the module and provider releases
            // it depends on are published.
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

// Directions for the published nuonco/stack/aws module, which reads its whole
// configuration from the API. Distinct from the TerraformTab above, which
// clones install-stacks and is driven by generated tfvars.
const TFModuleTab = ({ orgId, installId, installAwsRegion }: ITFModuleTab) => {
  const {
    data: token,
    isLoading,
    isError,
  } = useStackToken({
    installId,
    orgId,
    enabled: true,
  })

  const region = installAwsRegion ?? '<your-install-region>'

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
  version = "~> 0.2"

  install_id = "${installId ?? '<install-id>'}"
}`

  const backendSnippet = `terraform {
  backend "s3" {
    bucket = "<your-state-bucket>"
    key    = "nuon/${installId}/terraform.tfstate"
    region = "<your-state-bucket-region>"
  }
}`

  const authCmd = `export NUON_API_TOKEN='${token?.api_token ?? '<api-token>'}'`
  const applyCmd = `terraform init && terraform apply`

  return (
    <div className="flex flex-col gap-4 pt-4">
      <Text variant="subtext" theme="neutral">
        The <code>nuonco/stack/aws</code> module reads this install&apos;s
        configuration from the Nuon API, so the only input is the install ID.
        Everything else — runner details, IAM permissions, roles, inputs, and
        secrets — comes from the control plane.
      </Text>

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
          3. Authenticate
        </Text>
        {isError ? (
          <Text variant="subtext" theme="neutral">
            No API token has been issued for this install stack yet. Reprovision
            the install to mint one, then reload this page.
          </Text>
        ) : null}
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              This token authorizes the module to read its configuration. Treat
              it as a secret.
            </Text>
            {isLoading ? null : <ClickToCopyButton textToCopy={authCmd} />}
          </span>
          {isLoading ? (
            <Code loading />
          ) : (
            <Code variant="preformated">{authCmd}</Code>
          )}
        </Card>
        {token?.expires_at ? (
          <Text variant="subtext" theme="neutral">
            This token expires {new Date(token.expires_at).toLocaleString()}.
            Reload this page for a fresh one.
          </Text>
        ) : null}
        <Text variant="subtext" theme="neutral">
          In CI, prefer OIDC instead: grant{' '}
          <code>permissions: id-token: write</code>, set <code>org_id</code> on
          the <code>stack</code> provider, and omit the token entirely. That
          mints a credential per run, so nothing long-lived is stored.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Apply
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

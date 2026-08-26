import { useMemo } from 'react'
import { strToU8, zipSync } from 'fflate'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Expand } from '@/components/common/Expand'
import { Link } from '@/components/common/Link'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from '../types'

interface StackEnvelope {
  inputs: string
  providerInputs: string
  secrets: string
  spaceliftAdminTf: string
  spaceliftBlueprintYaml: string
}

function parseEnvelope(contents: unknown): StackEnvelope {
  const empty: StackEnvelope = {
    inputs: '',
    providerInputs: '',
    secrets: '',
    spaceliftAdminTf: '',
    spaceliftBlueprintYaml: '',
  }
  if (!contents) return empty

  let raw = contents
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
      spaceliftAdminTf: String(rec.spacelift_admin_tf ?? ''),
      spaceliftBlueprintYaml: String(rec.spacelift_blueprint_yaml ?? ''),
    }
  }

  return empty
}

// Unlike `terraform apply`, which prompts for missing required vars, Spacelift
// just uploads the file as-is — so surface these at the top as visible blanks.
function withSpaceliftGCPPlaceholders(tfvars: string): string {
  const rest = tfvars
    .split('\n')
    .filter((line) => !/^\s*gcp_project_id\s*=/.test(line))
    .filter((line) => !/^\s*gcp_region\s*=/.test(line))
    .join('\n')

  return `gcp_project_id = ""\ngcp_region     = ""\n\n${rest}`
}

interface IAwaitGCPDetails extends IStackDetails {
  installId?: string
  gcpProjectId?: string
  spaceliftEnabled?: boolean
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

export const AwaitGCPDetails = ({
  stack,
  installId,
  gcpProjectId,
  spaceliftEnabled,
  tfProvider = false,
  loading,
}: IAwaitGCPDetails) => {
  const version = stack?.versions?.at(0)
  const envelope = useMemo(
    () => parseEnvelope(version?.contents),
    [version?.contents]
  )
  const hasSpacelift =
    !!spaceliftEnabled &&
    (envelope.spaceliftAdminTf.length > 0 ||
      envelope.spaceliftBlueprintYaml.length > 0)
  const telemetryExportSecretID = `${installId || '<install-id>'}-telemetry-export-config`
  const projectID = gcpProjectId || '<gcp-project-id>'
  const createTelemetryExportSecretVersionCmd = `gcloud secrets versions add "${telemetryExportSecretID}" \\
  --data-file="${telemetryExportConfigFilename}" \\
  --project="${projectID}"`

  if (loading) {
    return (
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Setup your install stack
        </Text>
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            1. Clone the install stack module
          </Text>
          <Card>
            <Code loading />
          </Card>
        </div>
        <Divider />
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            2. Configure remote state (recommended)
          </Text>
          <Card>
            <Code loading />
          </Card>
        </div>
        <Divider />
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            3. Save the install configuration
          </Text>
          <Card>
            <Code loading />
          </Card>
        </div>
        <Divider />
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            4. Apply with Terraform
          </Text>
          <Card>
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

      {hasSpacelift ? (
        <Tabs
          initActiveTab="terraform"
          tabs={{
            terraform: (
              <TerraformTab
                inputsTfvars={envelope.inputs}
                providerTfvars={envelope.providerInputs}
                tfProvider={tfProvider}
                secretsTfvars={envelope.secrets}
                installId={installId}
              />
            ),
            spacelift: (
              <SpaceliftTab
                adminTf={envelope.spaceliftAdminTf}
                blueprintYaml={envelope.spaceliftBlueprintYaml}
                inputsTfvars={envelope.inputs}
                secretsTfvars={envelope.secrets}
                installId={installId}
              />
            ),
          }}
        />
      ) : (
        <TerraformTab
          inputsTfvars={envelope.inputs}
          providerTfvars={envelope.providerInputs}
          tfProvider={tfProvider}
          secretsTfvars={envelope.secrets}
          installId={installId}
        />
      )}

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
            update the <code>telemetry-export-config</code> secret in Secret
            Manager after the stack is applied. See the{' '}
            <Link
              href="https://docs.nuon.co/guides/export-runner-audit-logs"
              isExternal
              variant="inline"
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
              <ClickToCopyButton
                className="w-fit self-end"
                textToCopy={createTelemetryExportSecretVersionCmd}
              />
            </span>
            <Code variant="preformated">
              {createTelemetryExportSecretVersionCmd}
            </Code>
          </Card>
        </div>
      </Expand>
    </div>
  )
}

interface ITerraformTab {
  inputsTfvars: string
  providerTfvars: string
  tfProvider: boolean
  secretsTfvars: string
  installId?: string
}

const TerraformTab = ({
  inputsTfvars,
  providerTfvars,
  tfProvider,
  secretsTfvars,
  installId,
}: ITerraformTab) => {
  const inputsFile = tfProvider ? providerTfvars : inputsTfvars

  const cloneCmd = tfProvider
    ? `git clone -b ja/stack-sdk https://github.com/nuonco/install-stacks.git
cd install-stacks/gcp`
    : `git clone https://github.com/nuonco/install-stacks.git
cd install-stacks/gcp`

  const backendSnippet = `terraform {
  backend "gcs" {
    bucket = "<your-state-bucket>"
    prefix = "nuon/${installId}"
  }
}`

  const applyCmd = `terraform init && terraform apply`

  return (
    <div className="flex flex-col gap-4 pt-4">
      {tfProvider ? (
        <Text variant="subtext" theme="neutral">
          This module reads its configuration from the Nuon API via the{' '}
          <code>stack</code> Terraform provider, so the tfvars stay slim. The
          provider isn&apos;t published to the Terraform registry yet — add a
          dev override in <code>~/.terraformrc</code> pointing{' '}
          <code>nuonco/stack</code> at your local build before running{' '}
          <code>terraform init</code>.
        </Text>
      ) : null}

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Clone the install stack module
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Clone and enter the GCP module directory</Text>
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
              GCS
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
    </div>
  )
}

interface ISpaceliftTab {
  adminTf: string
  blueprintYaml: string
  inputsTfvars: string
  secretsTfvars: string
  installId?: string
}

const SpaceliftTab = ({
  adminTf,
  blueprintYaml,
  inputsTfvars,
  secretsTfvars,
  installId,
}: ISpaceliftTab) => {
  return (
    <div className="flex flex-col gap-4 pt-4">
      <Text variant="subtext" theme="neutral">
        Manage the install stack in Spacelift, using either a{' '}
        <Link
          href="https://docs.spacelift.io/concepts/blueprint/"
          target="_blank"
          rel="noopener noreferrer"
          variant="inline"
        >
          Blueprint
        </Link>{' '}
        or{' '}
        <Link
          href="https://registry.terraform.io/providers/spacelift-io/spacelift/latest/docs"
          target="_blank"
          rel="noopener noreferrer"
          variant="inline"
        >
          Terraform
        </Link>{' '}
        . If the customer is managing Spacelift through the Web UI, use the
        Blueprint. If the customer is managing Spacelift using Terraform, use
        the Terraform config.
      </Text>

      <Tabs
        initActiveTab="blueprint"
        tabs={{
          blueprint: <BlueprintSubTab blueprintYaml={blueprintYaml} />,
          terraform: (
            <TerraformSubTab
              adminTf={adminTf}
              inputsTfvars={inputsTfvars}
              secretsTfvars={secretsTfvars}
              installId={installId}
            />
          ),
        }}
      />
    </div>
  )
}

interface ITerraformSubTab {
  adminTf: string
  inputsTfvars: string
  secretsTfvars: string
  installId?: string
}

const TerraformSubTab = ({
  adminTf,
  inputsTfvars,
  secretsTfvars,
  installId,
}: ITerraformSubTab) => {
  const spaceliftInputsTfvars = useMemo(
    () => withSpaceliftGCPPlaceholders(inputsTfvars),
    [inputsTfvars]
  )

  return (
    <div className="flex flex-col gap-4 pt-4">
      <div className="flex flex-col gap-4">
        <span className="flex justify-between items-center">
          <Text variant="base" weight="strong">
            1. Add these files to your Spacelift terraform project
          </Text>
          <Button
            size="sm"
            variant="secondary"
            onClick={() => {
              const zipped = zipSync({
                'spacelift.tf': strToU8(adminTf),
                'inputs.auto.tfvars': strToU8(spaceliftInputsTfvars),
                'secrets.auto.tfvars': strToU8(secretsTfvars),
              })
              createFileDownload(
                new Blob([zipped], { type: 'application/zip' }),
                `nuon-spacelift-${installId ?? 'stack'}.zip`,
                'application/zip'
              )
            }}
          >
            Download all (.zip)
          </Button>
        </span>
        <Text variant="subtext" theme="neutral">
          Unpack all three into one directory, then edit before applying: set{' '}
          <code>space_id</code> in <code>spacelift.tf</code> to the Spacelift
          space this stack should live in (already have your own GCP integration
          for this stack? set <code>attach_gcp_service_account</code> to{' '}
          <code>false</code> instead), fill in <code>inputs.auto.tfvars</code>
          (including <code>gcp_project_id</code>/<code>gcp_region</code> at the
          top), and replace the placeholders in <code>secrets.auto.tfvars</code>{' '}
          with your real secrets.
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>spacelift.tf</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={adminTf} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() => createFileDownload(adminTf, 'spacelift.tf')}
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{adminTf}</Code>
        </Card>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>inputs.auto.tfvars</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={spaceliftInputsTfvars} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(
                    spaceliftInputsTfvars,
                    'inputs.auto.tfvars'
                  )
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{spaceliftInputsTfvars}</Code>
        </Card>
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
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          2. Authenticate the provider to Spacelift
        </Text>
        <Text variant="subtext" theme="neutral">
          Create a{' '}
          <Link
            href="https://docs.spacelift.io/integrations/api#spacelift-api-key"
            target="_blank"
            rel="noopener noreferrer"
            variant="inline"
          >
            Spacelift API key
          </Link>{' '}
          and export it as <code>SPACELIFT_API_KEY_ENDPOINT</code>,{' '}
          <code>SPACELIFT_API_KEY_ID</code>, and{' '}
          <code>SPACELIFT_API_KEY_SECRET</code> before applying.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Apply with Terraform
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Run in the directory with all three files</Text>
            <ClickToCopyButton textToCopy="terraform init && terraform apply" />
          </span>
          <Code variant="preformated">terraform init && terraform apply</Code>
        </Card>
        <Text variant="subtext" theme="neutral">
          This creates the install stack, mounts your tfvars, and (unless you
          set <code>attach_gcp_service_account</code> to <code>false</code>)
          attaches Spacelift&apos;s native GCP integration, no manual{' '}
          <strong>Settings → Integrations</strong> step needed.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Grant access, then run the install stack
        </Text>
        <Text variant="subtext" theme="neutral">
          If you attached the GCP integration (the default), grant the printed{' '}
          <code>gcp_service_account_email</code> output an IAM role on your
          target GCP project. If you attached your own integration instead,
          confirm its identity already has that access. Then trigger the
          stack&apos;s first run, it&apos;s set to auto-deploy, so it plans and
          applies your runner without further approval.
        </Text>
      </div>
    </div>
  )
}

interface IBlueprintSubTab {
  blueprintYaml: string
}

const BlueprintSubTab = ({ blueprintYaml }: IBlueprintSubTab) => {
  return (
    <div className="flex flex-col gap-4 pt-4">
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          1. Create the blueprint
        </Text>
        <Text variant="subtext" theme="neutral">
          In Spacelift, go to <strong>Blueprints → Create blueprint</strong> and
          paste this YAML as the template body.
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Blueprint template body</Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={blueprintYaml} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(blueprintYaml, 'blueprint.yaml')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{blueprintYaml}</Code>
        </Card>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          2. Publish it
        </Text>
        <Text variant="subtext" theme="neutral">
          Click <strong>Publish</strong> to move the blueprint from draft to
          published. Publishing is one-way — to change a published blueprint you
          clone it, edit, and publish again.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          3. Create a stack and fill in the inputs
        </Text>
        <Text variant="subtext" theme="neutral">
          On the published blueprint, click <strong>Create stack</strong> and
          fill in the inputs and secrets. This creates the stack but
          doesn&apos;t run it.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Configure credentials
        </Text>
        <Text variant="subtext" theme="neutral">
          If you haven't already, configure the{' '}
          <Link
            href="https://docs.spacelift.io/integrations/cloud-providers"
            target="_blank"
            rel="noopener noreferrer"
            variant="inline"
          >
            Spacelift Integration
          </Link>{' '}
          for the cloud you want to deploy to.
        </Text>
      </div>

      <Divider />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          4. Run the stack
        </Text>
        <Text variant="subtext" theme="neutral">
          Click <strong>Trigger</strong> on the stack page to run the stack.
        </Text>
      </div>
    </div>
  )
}

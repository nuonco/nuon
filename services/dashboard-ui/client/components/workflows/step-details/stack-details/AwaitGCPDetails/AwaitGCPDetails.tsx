import { useMemo } from 'react'
import { strToU8, zipSync } from 'fflate'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from '../types'

interface StackEnvelope {
  inputs: string
  secrets: string
  spaceliftAdminTf: string
  spaceliftBlueprintYaml: string
}

function parseEnvelope(contents: unknown): StackEnvelope {
  const empty: StackEnvelope = {
    inputs: '',
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
    return {
      inputs: String(rec.inputs_tfvars ?? ''),
      secrets: String(rec.secrets_tfvars ?? ''),
      spaceliftAdminTf: String(rec.spacelift_admin_tf ?? ''),
      spaceliftBlueprintYaml: String(rec.spacelift_blueprint_yaml ?? ''),
    }
  }

  return empty
}

interface IAwaitGCPDetails extends IStackDetails {
  installId?: string
  spaceliftEnabled?: boolean
}

export const AwaitGCPDetails = ({
  stack,
  installId,
  spaceliftEnabled,
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
          secretsTfvars={envelope.secrets}
          installId={installId}
        />
      )}
    </div>
  )
}

interface ITerraformTab {
  inputsTfvars: string
  secretsTfvars: string
  installId?: string
}

const TerraformTab = ({
  inputsTfvars,
  secretsTfvars,
  installId,
}: ITerraformTab) => {
  const cloneCmd = `git clone https://github.com/nuonco/install-stacks.git
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
              <ClickToCopyButton textToCopy={inputsTfvars} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(inputsTfvars, 'inputs.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{inputsTfvars}</Code>
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
        >
          Blueprint
        </Link>{' '}
        or{' '}
        <Link
          href="https://registry.terraform.io/providers/spacelift-io/spacelift/latest/docs"
          target="_blank"
          rel="noopener noreferrer"
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
                'inputs.auto.tfvars': strToU8(inputsTfvars),
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
              <ClickToCopyButton textToCopy={inputsTfvars} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(inputsTfvars, 'inputs.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{inputsTfvars}</Code>
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

export const AwaitGCPDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="275px" />

      <Card>
        <Skeleton height="17px" width="250px" />
        <Skeleton height="52px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="300px" />

      <Card>
        <Skeleton height="17px" width="300px" />
        <Skeleton height="72px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="250px" />

      <Card>
        <Skeleton height="17px" width="200px" />
        <Skeleton height="100px" width="100%" />
      </Card>

      <Divider />

      <Skeleton height="24px" width="200px" />

      <Card>
        <Skeleton height="17px" width="150px" />
        <Skeleton height="52px" width="100%" />
      </Card>
    </>
  )
}

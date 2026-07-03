import { useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
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
}

export const AwaitGCPDetails = ({ stack, installId }: IAwaitGCPDetails) => {
  const version = stack?.versions?.at(0)
  const envelope = useMemo(
    () => parseEnvelope(version?.contents),
    [version?.contents]
  )
  const hasSpacelift =
    envelope.spaceliftAdminTf.length > 0 ||
    envelope.spaceliftBlueprintYaml.length > 0

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
}

const SpaceliftTab = ({ adminTf, blueprintYaml }: ISpaceliftTab) => {
  return (
    <div className="flex flex-col gap-4 pt-4">
      <Text variant="subtext" theme="neutral">
        Run the install stack in Spacelift instead of applying Terraform
        yourself. Both options run the same install-stacks module and mount your
        generated tfvars — pick whichever fits how you manage Spacelift.
      </Text>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Option A — administrative stack
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Apply this from an existing admin stack as{' '}
              <code>spacelift.tf</code>
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
      </div>

      <Divider dividerWord="or" />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Option B — blueprint
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Import and publish this blueprint, then create a stack from it
            </Text>
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

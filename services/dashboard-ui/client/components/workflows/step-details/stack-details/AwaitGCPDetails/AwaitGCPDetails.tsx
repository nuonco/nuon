import { useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from '../types'

interface TfvarsEnvelope {
  inputs: string
  secrets: string
}

function parseTfvars(contents: unknown): TfvarsEnvelope {
  if (!contents) return { inputs: '', secrets: '' }

  let raw = contents
  if (typeof raw === 'string') {
    try {
      raw = JSON.parse(raw)
    } catch {
      try {
        raw = JSON.parse(atob(raw as string))
      } catch {
        return { inputs: '', secrets: '' }
      }
    }
  }

  if (typeof raw === 'object' && raw !== null) {
    const rec = raw as Record<string, unknown>
    return {
      inputs: String(rec.inputs_tfvars ?? ''),
      secrets: String(rec.secrets_tfvars ?? ''),
    }
  }

  return { inputs: '', secrets: '' }
}

interface IAwaitGCPDetails extends IStackDetails {
  installId?: string
}

export const AwaitGCPDetails = ({ stack, installId }: IAwaitGCPDetails) => {
  const version = stack?.versions?.at(0)
  const tfvars = useMemo(
    () => parseTfvars(version?.contents),
    [version?.contents]
  )

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
    <>
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
              <ClickToCopyButton textToCopy={tfvars.inputs} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(tfvars.inputs, 'inputs.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{tfvars.inputs}</Code>
        </Card>
        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Save this as <code>secrets.auto.tfvars</code>
            </Text>
            <span className="flex gap-2 items-center">
              <ClickToCopyButton textToCopy={tfvars.secrets} />
              <Button
                size="sm"
                variant="secondary"
                onClick={() =>
                  createFileDownload(tfvars.secrets, 'secrets.auto.tfvars')
                }
              >
                Download
              </Button>
            </span>
          </span>
          <Code variant="preformated">{tfvars.secrets}</Code>
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
    </>
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

'use client'

import { useMemo } from 'react'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import type { IStackDetails } from './types'

function parseTfvars(contents: unknown): string {
  if (!contents) return ''

  let raw = contents
  if (typeof raw === 'string') {
    try {
      raw = JSON.parse(raw)
    } catch {
      try {
        raw = JSON.parse(atob(raw as string))
      } catch {
        return ''
      }
    }
  }

  if (typeof raw === 'object' && raw !== null && 'tfvars' in raw) {
    return String((raw as Record<string, unknown>).tfvars ?? '')
  }

  return ''
}

export const AwaitGCPDetails = ({ stack }: IStackDetails) => {
  const { install } = useInstall()

  const version = stack?.versions?.at(0)
  const tfvarsContent = useMemo(
    () => parseTfvars(version?.contents),
    [version?.contents]
  )

  const cloneCmd = `git clone https://github.com/nuonco/install-stacks.git
cd install-stacks/gcp`

  const backendSnippet = `terraform {
  backend "gcs" {
    bucket = "<your-state-bucket>"
    prefix = "nuon/${install?.id}"
  }
}`

  const applyCmd = `terraform init && terraform apply -var-file=install.tfvars`

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
              Save this as <code>install.tfvars</code>
            </Text>
            <ClickToCopyButton textToCopy={tfvarsContent} />
          </span>
          <Code variant="preformated">{tfvarsContent}</Code>
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

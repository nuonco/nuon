'use client'

import { useState } from 'react'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { useInstall } from '@/hooks/use-install'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from './types'

export const AwaitGCPDetails = ({ stack }: IStackDetails) => {
  const { install } = useInstall()
  const templateUrl = stack?.versions?.at(0)?.template_url
  const [isDownloading, setIsDownloading] = useState(false)

  const handleDownload = async () => {
    if (!templateUrl) return
    setIsDownloading(true)
    try {
      const response = await fetch(templateUrl)
      const data = await response.arrayBuffer()
      createFileDownload(data, 'main.tf.json', 'application/json')
    } catch (error) {
      console.error('Error downloading template:', error)
    } finally {
      setIsDownloading(false)
    }
  }

  const terraformCmd = `mkdir -p nuon-stack && cd nuon-stack && curl -o main.tf.json "${templateUrl}" && terraform init && terraform apply`
  const gcpConsoleUrl = `https://console.cloud.google.com/infra-manager/deployments?project=${install?.gcp_account?.project_id}`

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Provision the install stack using Terraform
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>1. Authenticate to GCP</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`gcloud auth application-default login --project=${install?.gcp_account?.project_id}`}
            />
          </span>
          <Code>{`gcloud auth application-default login --project=${install?.gcp_account?.project_id}`}</Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>2. Download and apply the stack</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={terraformCmd}
            />
          </span>
          <Code>{terraformCmd}</Code>
        </Card>
      </div>

      <Divider dividerWord="or" />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Download and deploy manually
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <span className="flex flex-col gap-1">
              <Text weight="strong">Terraform configuration</Text>
              <Text variant="subtext">
                Download the pre-configured Terraform JSON and apply it in your GCP project
              </Text>
            </span>
            <Button
              size="sm"
              variant="secondary"
              onClick={handleDownload}
              disabled={isDownloading}
            >
              {isDownloading ? 'Downloading...' : 'Download main.tf.json'}
            </Button>
          </span>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text weight="strong">GCP Console</Text>
            <Link
              href={gcpConsoleUrl}
              target="_blank"
              rel="noopener noreferrer"
            >
              <Button size="sm" variant="secondary">
                Open Infrastructure Manager
              </Button>
            </Link>
          </span>
          <Text variant="subtext">
            Upload the downloaded template in GCP Infrastructure Manager to deploy
          </Text>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Template URL</Text>
            <ClickToCopyButton textToCopy={templateUrl} />
          </span>
          <Link
            href={templateUrl}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{templateUrl}</Code>
          </Link>
        </Card>
      </div>
    </>
  )
}

export const AwaitGCPDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="175px" />

      <Card>
        <Skeleton height="17px" width="100px" />
        <Skeleton height="52px" width="100%" />
      </Card>

      <Card>
        <Skeleton height="17px" width="120px" />
        <Skeleton height="52px" width="100%" />
      </Card>

      <Divider dividerWord="or" />

      <Skeleton height="24px" width="325px" />

      <Card>
        <Skeleton height="17px" width="219px" />
        <Skeleton height="72px" width="100%" />
      </Card>
    </>
  )
}

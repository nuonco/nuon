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
  const projectId = install?.gcp_account?.project_id
  const region = install?.gcp_account?.region || 'us-central1'
  const installId = install?.id
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

  const setupCmd = `gcloud services enable config.googleapis.com cloudbuild.googleapis.com --project=${projectId} && \\
gcloud iam service-accounts create nuon-deployer --project=${projectId} --display-name="Nuon Infrastructure Manager deployer" && \\
gcloud projects add-iam-policy-binding ${projectId} \\
  --member="serviceAccount:nuon-deployer@${projectId}.iam.gserviceaccount.com" \\
  --role="roles/config.agent" && \\
gcloud projects add-iam-policy-binding ${projectId} \\
  --member="serviceAccount:nuon-deployer@${projectId}.iam.gserviceaccount.com" \\
  --role="roles/editor"`

  const deployCmd = `gcloud infra-manager deployments apply \\
  projects/${projectId}/locations/${region}/deployments/nuon-${installId} \\
  --service-account=projects/${projectId}/serviceAccounts/nuon-deployer@${projectId}.iam.gserviceaccount.com \\
  --gcs-source=${templateUrl?.replace('https://storage.googleapis.com/', 'gs://')} \\
  --project=${projectId}`

  const terraformCmd = `mkdir -p nuon-stack && cd nuon-stack && curl -o main.tf.json "${templateUrl}" && terraform init && terraform apply`

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Deploy using GCP Infrastructure Manager
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <span className="flex flex-col gap-1">
              <Text weight="strong">Step 1: One-time setup</Text>
              <Text variant="subtext">
                Enable APIs and create a service account for Infrastructure Manager (run once per project)
              </Text>
            </span>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={setupCmd}
            />
          </span>
          <Code>{setupCmd}</Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <span className="flex flex-col gap-1">
              <Text weight="strong">Step 2: Deploy the stack</Text>
              <Text variant="subtext">
                Infrastructure Manager will provision VPC, runner, and all required resources
              </Text>
            </span>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={deployCmd}
            />
          </span>
          <Code>{deployCmd}</Code>
        </Card>
      </div>

      <Divider dividerWord="or" />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Deploy using Terraform CLI
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Download and apply with Terraform</Text>
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
          Download template
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <span className="flex flex-col gap-1">
              <Text weight="strong">Terraform configuration</Text>
              <Text variant="subtext">
                Download and deploy manually
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
        <Skeleton height="72px" width="100%" />
      </Card>

      <Card>
        <Skeleton height="17px" width="120px" />
        <Skeleton height="72px" width="100%" />
      </Card>

      <Divider dividerWord="or" />

      <Skeleton height="24px" width="200px" />

      <Card>
        <Skeleton height="17px" width="200px" />
        <Skeleton height="52px" width="100%" />
      </Card>

      <Divider dividerWord="or" />

      <Skeleton height="24px" width="175px" />

      <Card>
        <Skeleton height="17px" width="150px" />
      </Card>
    </>
  )
}

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
  const gcpAccount = (install as any)?.gcp_account as { project_id?: string; region?: string } | undefined
  const projectId = gcpAccount?.project_id
  const region = gcpAccount?.region || 'us-central1'
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

  const enableApisCmd = `gcloud services enable config.googleapis.com cloudbuild.googleapis.com --project=${projectId}`

  const createSaCmd = `gcloud iam service-accounts create nuon-deployer --project=${projectId} --display-name="Nuon Infrastructure Manager deployer"`

  const grantRolesCmd = `gcloud projects add-iam-policy-binding ${projectId} --member="serviceAccount:nuon-deployer@${projectId}.iam.gserviceaccount.com" --role="roles/config.agent"

gcloud projects add-iam-policy-binding ${projectId} --member="serviceAccount:nuon-deployer@${projectId}.iam.gserviceaccount.com" --role="roles/editor"

gcloud projects add-iam-policy-binding ${projectId} --member="serviceAccount:nuon-deployer@${projectId}.iam.gserviceaccount.com" --role="roles/resourcemanager.projectIamAdmin"`

  // Convert S3 URL to GCS URI if template is on GCS, otherwise use S3 URL as-is
  const gcsSource = templateUrl?.includes('storage.googleapis.com')
    ? templateUrl.replace('https://storage.googleapis.com/', 'gs://')
    : templateUrl

  const deployCmd = `gcloud infra-manager deployments apply projects/${projectId}/locations/${region}/deployments/nuon-${installId} --service-account=projects/${projectId}/serviceAccounts/nuon-deployer@${projectId}.iam.gserviceaccount.com --gcs-source=${gcsSource} --project=${projectId}`

  const terraformCmd = `mkdir -p nuon-stack && cd nuon-stack && curl -o main.tf.json "${templateUrl}" && terraform init && terraform apply`

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Deploy using GCP Infrastructure Manager
        </Text>

        <Card>
          <Text weight="strong">Step 1: Enable required APIs</Text>
          <span className="flex justify-between items-center">
            <Code className="flex-1">{enableApisCmd}</Code>
            <ClickToCopyButton className="w-fit ml-2" textToCopy={enableApisCmd} />
          </span>
        </Card>

        <Card>
          <Text weight="strong">Step 2: Create deployer service account</Text>
          <span className="flex justify-between items-center">
            <Code className="flex-1">{createSaCmd}</Code>
            <ClickToCopyButton className="w-fit ml-2" textToCopy={createSaCmd} />
          </span>
        </Card>

        <Card>
          <Text weight="strong">Step 3: Grant permissions</Text>
          <span className="flex justify-between items-center">
            <Code className="flex-1">{grantRolesCmd}</Code>
            <ClickToCopyButton className="w-fit ml-2" textToCopy={grantRolesCmd} />
          </span>
        </Card>

        <Card>
          <Text weight="strong">Step 4: Deploy the stack</Text>
          <Text variant="subtext">
            Infrastructure Manager will provision VPC, runner, and all required resources
          </Text>
          <span className="flex justify-between items-center">
            <Code className="flex-1">{deployCmd}</Code>
            <ClickToCopyButton className="w-fit ml-2" textToCopy={deployCmd} />
          </span>
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

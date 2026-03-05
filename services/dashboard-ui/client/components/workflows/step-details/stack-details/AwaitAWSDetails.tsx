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
import { useOrg } from '@/hooks/use-org'
import { createFileDownload } from '@/utils/file-download'
import type { IStackDetails } from './types'

export const AwaitAWSDetails = ({ stack }: IStackDetails) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [isDownloading, setIsDownloading] = useState(false)

  const handleDownloadTerraformConfig = async () => {
    if (!org?.id || !install?.id) {
      console.error('Missing org ID or install ID')
      return
    }

    setIsDownloading(true)
    try {
      const response = await fetch(
        `/api/orgs/${org.id}/installs/${install.id}/generate-terraform-installer-config`
      )

      if (!response.ok) {
        throw new Error('Failed to generate terraform installer config')
      }

      const configData = await response.arrayBuffer()
      createFileDownload(
        configData,
        'terraform.tfvars',
        'application/octet-stream'
      )
    } catch (error) {
      console.error('Error downloading terraform installer config:', error)
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Setup your install stack
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Install quick link</Text>
            <ClickToCopyButton
              textToCopy={stack?.versions?.at(0)?.quick_link_url}
            />
          </span>
          <Link
            href={stack?.versions?.at(0)?.quick_link_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{stack?.versions?.at(0)?.quick_link_url}</Code>
          </Link>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text weight="strong">Install template link</Text>
            <ClickToCopyButton
              textToCopy={stack?.versions?.at(0)?.template_url}
            />
          </span>
          <Link
            href={stack?.versions?.at(0)?.template_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{stack?.versions?.at(0)?.template_url}</Code>
          </Link>
        </Card>

        {org?.features?.['terraform-installer'] && (
          <Card>
            <span className="flex justify-between items-center">
              <Text weight="strong">Terraform installer config</Text>
              <Button
                size="sm"
                variant="secondary"
                onClick={handleDownloadTerraformConfig}
                disabled={isDownloading}
              >
                {isDownloading ? 'Downloading...' : 'Download terraform.tfvars'}
              </Button>
            </span>
            <Text variant="subtext">
              Download a pre-configured terraform.tfvars file with your
              install-specific values
            </Text>
          </Card>
        )}
      </div>

      <Divider dividerWord="or" />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Setup your install stack using CLI command
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>AWS CloudFormation create stack</Text>

            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={` aws cloudformation create-stack --stack-name [YOUR_STACK_NAME]
            --template-url ${stack?.versions?.at(0)?.template_url}`}
            />
          </span>
          <Code>
            aws cloudformation create-stack --stack-name [YOUR_STACK_NAME]
            --template-url {stack?.versions?.at(0)?.template_url}
          </Code>
        </Card>
      </div>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Update an existing install stack using CLI command
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>AWS CloudFormation update stack</Text>

            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={` aws cloudformation update-stack --stack-name [YOUR_STACK_NAME]
            --template-url ${stack?.versions?.at(0)?.template_url}`}
            />
          </span>
          <Code>
            aws cloudformation update-stack --stack-name [YOUR_STACK_NAME]
            --template-url {stack?.versions?.at(0)?.template_url}
          </Code>
        </Card>
      </div>
    </>
  )
}

export const AwaitAWSDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="175px" />

      <Card>
        <Skeleton height="17px" width="100px" />
        <Skeleton height="132px" width="100%" />
      </Card>

      <Card>
        <Skeleton height="17px" width="120px" />
        <Skeleton height="72px" width="100%" />
      </Card>

      <Card>
        <Skeleton height="17px" width="175px" />
        <Skeleton height="32px" width="100%" />
      </Card>

      <Divider dividerWord="or" />

      <Skeleton height="24px" width="325px" />

      <Card>
        <Skeleton height="17px" width="219px" />
        <Skeleton height="92px" width="100%" />
      </Card>

      <Skeleton height="24px" width="382px" />

      <Card>
        <Skeleton height="17px" width="223px" />
        <Skeleton height="92px" width="100%" />
      </Card>
    </>
  )
}

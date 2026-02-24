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

  const isReprovision = (stack?.versions?.length ?? 0) > 1
  const initFlag = isReprovision ? ' -reconfigure' : ''

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

  const terraformCmd = `mkdir -p nuon-${install?.id} && cd nuon-${install?.id} && curl -o main.tf.json "${templateUrl}" && terraform init${initFlag} && terraform apply`

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Setup your install stack
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Install template link</Text>
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

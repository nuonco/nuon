'use client'

import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Divider } from '@/components/common/Divider'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import type { IStackDetails } from './types'

export const AwaitGCPDetails = ({ stack }: IStackDetails) => {
  const { install } = useInstall()

  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Provision the install stack using Terraform
        </Text>

        <Card>
          <span className="flex justify-between items-center">
            <Text>
              Ensure you are authenticated to GCP
            </Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`gcloud auth application-default login --project=${install?.gcp_account?.project_id}`}
            />
          </span>
          <Code>{`gcloud auth application-default login --project=${install?.gcp_account?.project_id}`}</Code>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text>Download and apply the stack template</Text>
            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={`curl -o stack.tf.json "${stack?.versions?.at(0)?.template_url}" && terraform init && terraform apply`}
            />
          </span>
          <Code>{`curl -o stack.tf.json "${stack?.versions?.at(0)?.template_url}" && terraform init && terraform apply`}</Code>
        </Card>
      </div>

      <Divider dividerWord="or" />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Download the install stack template
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Install template link</Text>
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

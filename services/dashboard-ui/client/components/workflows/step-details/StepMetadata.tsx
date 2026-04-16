'use client'

import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Link } from '@/components/common/Link'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import type { IStepDetails } from './types'

const StepHistoryStatus = ({
  status,
}: {
  status: IStepDetails['step']['status']['history'][number]
}) => {
  return (
    <span className="flex items-center gap-4 py-2">
      <Status status={status.status} variant="badge" />
      <Time seconds={status.created_at_ts} variant="subtext" theme="neutral" />
    </span>
  )
}

export const StepMetadata = ({ step }: IStepDetails) => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <div className="flex flex-col gap-2">
      <Text variant="label" theme="neutral">
        Triggered by {step?.created_by?.email}
      </Text>

      <Expand
        className="border rounded-md"
        id="step-history"
        heading={
          <Text family="mono" variant="subtext">
            View status history
          </Text>
        }
      >
        <div className="border-t flex flex-col p-4 divide-y">
          {step.status?.history?.map((status, idx) => (
            <StepHistoryStatus
              key={`${status.created_at_ts}-${idx}`}
              status={status}
            />
          ))}
          <StepHistoryStatus status={step.status} />
        </div>
      </Expand>

      {step.status?.metadata?.retry_type && (
        <div className="flex flex-col gap-1 border rounded-md p-4">
          <Text variant="label" theme="neutral">
            Retry info
          </Text>
          <Text variant="subtext">Type: {step.status.metadata.retry_type as string}</Text>
          {step.status.metadata.retry_idx !== undefined && (
            <Text variant="subtext">
              Attempt: {step.status.metadata.retry_idx as number}
              {step.status.metadata.max_retries !== undefined
                ? ` / ${step.status.metadata.max_retries}`
                : ''}
            </Text>
          )}
        </div>
      )}

      <Expand
        className="border rounded-md"
        id="step-json"
        heading={
          <Text family="mono" variant="subtext">
            View step JSON
          </Text>
        }
      >
        <div className="border-t">
          <CodeBlock language="json">{JSON.stringify(step, null, 2)}</CodeBlock>
        </div>
      </Expand>

      <Link className="text-xs" href={`/${org?.id}/installs/${install?.id}/workflows`}>
        View workflows
      </Link>
    </div>
  )
}

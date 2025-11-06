'use client'

import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
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
    </div>
  )
}

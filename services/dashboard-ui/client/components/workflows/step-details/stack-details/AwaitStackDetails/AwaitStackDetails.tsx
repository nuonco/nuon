import { useMemo } from 'react'
import { Card } from '@/components/common/Card'
import {
  KeyValueList,
  KeyValueListSkeleton,
} from '@/components/common/KeyValueList'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { AwaitAWSDetails } from '../AwaitAWSDetails'
import { AwaitAzureDetails } from '../AwaitAzureDetails'
import { AwaitGCPDetails } from '../AwaitGCPDetails'
import type { IStackDetails } from '../types'

interface IAwaitStackDetails extends IStackDetails {
  runnerType?: string
  spaceliftEnabled?: boolean
}

export const AwaitStackDetails = ({
  stack,
  step,
  runnerType,
  spaceliftEnabled,
  loading,
  ...props
}: IAwaitStackDetails) => {
  // Strictly this step's own stack version. Every version phones home
  // separately, so borrowing another one — the newest, or the stack-level
  // outputs record, which is overwritten by whichever version ran last — shows
  // a different stack's values under this step.
  const version = useMemo(
    () =>
      step?.step_target_id
        ? stack?.versions?.find((v) => v?.id === step.step_target_id)
        : stack?.versions?.at(0),
    [stack?.versions, step?.step_target_id]
  )

  // The stack endpoint only returns the most recent versions, so a step old
  // enough to have fallen off the end has no version to read.
  const versionMissing = !!step?.step_target_id && !version && !loading

  const latestRun = version?.runs?.at(0)
  const outputValues = useMemo(
    () => objectToKeyValueArray(latestRun?.data_contents),
    [latestRun]
  )

  const cloudDetails = runnerType?.startsWith('aws') ? (
    <AwaitAWSDetails stack={stack} step={step} loading={loading} {...props} />
  ) : runnerType === 'gcp' ? (
    <AwaitGCPDetails
      stack={stack}
      step={step}
      loading={loading}
      spaceliftEnabled={spaceliftEnabled}
      {...props}
    />
  ) : (
    <AwaitAzureDetails stack={stack} step={step} loading={loading} {...props} />
  )

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <Text>
          {loading || !version
            ? 'Install stack'
            : version.composite_status?.status === 'active'
              ? 'Install stack up and running'
              : 'Install stack is waiting to run'}
        </Text>

        <div className="grid grid-cols-4">
          {loading ? (
            <LabeledStatus label="Current status" loading />
          ) : (
            <LabeledStatus
              label="Current status"
              statusProps={{ status: version?.composite_status?.status }}
              tooltipProps={{
                tipContent: version?.composite_status?.status_human_description,
              }}
            />
          )}

          <LabeledValue label="Last checked" loading={loading}>
            {latestRun?.updated_at ? (
              <Time
                variant="subtext"
                time={latestRun.updated_at}
                format="relative"
              />
            ) : (
              <Text variant="subtext" theme="neutral">
                —
              </Text>
            )}
          </LabeledValue>
        </div>
      </Card>

      {cloudDetails}

      <Card>
        <Text>Stack outputs</Text>
        {loading ? (
          <KeyValueListSkeleton />
        ) : (
          <KeyValueList
            values={outputValues}
            emptyStateProps={{
              variant: 'table',
              size: 'sm',
              ...(versionMissing
                ? {
                    emptyTitle: 'Outputs not available',
                    emptyMessage:
                      'This stack version has dropped out of the install stack history.',
                  }
                : {
                    emptyTitle: 'No outputs yet',
                    emptyMessage:
                      'Outputs show up here once this stack has been applied in your cloud account.',
                  }),
            }}
          />
        )}
      </Card>
    </div>
  )
}

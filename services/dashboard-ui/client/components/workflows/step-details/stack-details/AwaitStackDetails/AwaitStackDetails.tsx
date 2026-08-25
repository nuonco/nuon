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
  runnerType,
  spaceliftEnabled,
  loading,
  ...props
}: IAwaitStackDetails) => {
  const outputValues = useMemo(
    () => objectToKeyValueArray(stack?.install_stack_outputs?.data_contents),
    [stack?.install_stack_outputs]
  )

  const cloudDetails = runnerType?.startsWith('aws') ? (
    <AwaitAWSDetails stack={stack} loading={loading} {...props} />
  ) : runnerType === 'gcp' ? (
    <AwaitGCPDetails
      stack={stack}
      loading={loading}
      spaceliftEnabled={spaceliftEnabled}
      {...props}
    />
  ) : (
    <AwaitAzureDetails stack={stack} loading={loading} {...props} />
  )

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <Text>
          {loading
            ? 'Install stack'
            : stack?.versions?.at(0)?.composite_status?.status === 'active'
              ? 'Install stack up and running'
              : 'Install stack is waiting to run'}
        </Text>

        <div className="grid grid-cols-4">
          {loading ? (
            <LabeledStatus label="Current status" loading />
          ) : (
            <LabeledStatus
              label="Current status"
              statusProps={{
                status: stack?.versions?.at(0)?.composite_status?.status,
              }}
              tooltipProps={{
                tipContent:
                  stack?.versions?.at(0)?.composite_status
                    ?.status_human_description,
              }}
            />
          )}

          <LabeledValue label="Last checked" loading={loading}>
            <Time
              variant="subtext"
              time={stack?.versions?.at(0)?.runs?.at(-1)?.updated_at}
              format="relative"
            />
          </LabeledValue>
        </div>
      </Card>

      {cloudDetails}

      <Card>
        <Text>Stack outputs</Text>
        {loading ? (
          <KeyValueListSkeleton />
        ) : (
          <KeyValueList values={outputValues} />
        )}
      </Card>
    </div>
  )
}

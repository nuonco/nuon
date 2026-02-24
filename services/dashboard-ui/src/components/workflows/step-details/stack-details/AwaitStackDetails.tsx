'use client'

import { useMemo } from 'react'
import { Card } from '@/components/common/Card'
import {
  KeyValueList,
  KeyValueListSkeleton,
} from '@/components/common/KeyValueList'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useInstall } from '@/hooks/use-install'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { AwaitAWSDetails, AwaitAWSDetailsSkeleton } from './AwaitAWSDetails'
import {
  AwaitAzureDetails,
  AwaitAzureDetailsSkeleton,
} from './AwaitAzureDetails'
import type { IStackDetails } from './types'

export const AwaitStackDetails = ({ stack, ...props }: IStackDetails) => {
  const outputValues = useMemo(
    () => objectToKeyValueArray(stack?.install_stack_outputs?.data),
    [stack?.install_stack_outputs]
  )
  const { install } = useInstall()

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <Text>
          Install stack{' '}
          {stack?.versions?.at(0)?.composite_status?.status === 'active'
            ? 'up and running'
            : 'is waiting to run'}
        </Text>

        <div className="grid grid-cols-4">
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

          <LabeledValue label="Last checked">
            <Time
              variant="subtext"
              time={stack?.versions?.at(0).runs?.at(-1)?.updated_at}
              format="relative"
            />
          </LabeledValue>
        </div>
      </Card>

      {install?.app_runner_config?.app_runner_type?.startsWith('aws') ? (
        <AwaitAWSDetails stack={stack} {...props} />
      ) : (
        <AwaitAzureDetails stack={stack} {...props} />
      )}

      <Card>
        <Text>Stack outputs</Text>
        <KeyValueList values={outputValues} />
      </Card>
    </div>
  )
}

export const AwaitStackDetailsSkeleton = () => {
  const { install } = useInstall()

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <Skeleton height="17px" width="175px" />

        <div className="grid grid-cols-4">
          <LabeledValue label={<Skeleton height="17px" width="34px" />}>
            <Skeleton height="23px" width="75px" />
          </LabeledValue>

          <LabeledValue label={<Skeleton height="17px" width="53px" />}>
            <Skeleton height="23px" width="148px" />
          </LabeledValue>
        </div>
      </Card>

      {install?.app_runner_config?.app_runner_type?.startsWith('aws') ? (
        <AwaitAWSDetailsSkeleton />
      ) : (
        <AwaitAzureDetailsSkeleton />
      )}

      <Card>
        <Skeleton height="24px" width="142px" />
        <KeyValueListSkeleton />
      </Card>
    </div>
  )
}

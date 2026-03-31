'use client'

import { Duration } from '@/components/common/Duration'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { hydrateActionRunSteps, sortByIdx } from '@/utils/action-utils'
import { toSentenceCase } from '@/utils/string-utils'
import type { IStandardActionSteps } from './types'

export const StandardActionSteps = ({ actionRun }: IStandardActionSteps) => {
  const hydratedSteps = sortByIdx(
    hydrateActionRunSteps({
      steps: actionRun.steps,
      stepConfigs: actionRun?.config?.steps,
    })
  )

  return (
    <div className="flex flex-col gap-2">
      <Text weight="strong">Action steps</Text>
      {hydratedSteps?.map((actionStep) => (
        <span
          key={actionStep.id}
          className="py-2 px-4 border rounded-md flex items-center justify-between"
        >
          <span className="flex items-center gap-2">
            <Status status={actionStep.status} isWithoutText />
            <Text>{toSentenceCase(actionStep?.name)}</Text>
          </span>

          <Text
            className="flex items-center gap-1"
            variant="subtext"
            theme="neutral"
          >
            {toSentenceCase(actionStep.status)}{' '}
            {actionStep?.execution_duration > 1000000 ? (
              <>
                in{' '}
                <Duration
                  variant="subtext"
                  nanoseconds={actionStep?.execution_duration}
                  theme="neutral"
                />
              </>
            ) : null}
          </Text>
        </span>
      ))}
    </div>
  )
}

export const StandardActionStepsSkeleton = () => {
  return (
    <div className="flex flex-col gap-2">
      <Skeleton height="17px" width="80px" />
      {Array.from({ length: 3 }).map((_, idx) => (
        <Skeleton key={idx} height="42px" width="100%" />
      ))}
    </div>
  )
}

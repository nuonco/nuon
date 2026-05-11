import { Text } from '@/components/common/Text'
import type { TInstallActionRun, TAccount, TWorkflowStep } from '@/types'
import { ActionRunHeader, ActionRunHeaderSkeleton } from '../ActionRunHeader'
import {
  ActionRunMetadata,
  ActionRunMetadataSkeleton,
} from '../ActionRunMetadata'
import { AdhocActionDetails } from '../AdhocActionDetails'
import { StandardActionSteps, StandardActionStepsSkeleton } from '../StandardActionSteps'
import { ActionRunLogs, ActionRunLogsSkeleton } from '../ActionRunLogs'

export interface IActionRunStepDetails {
  step?: TWorkflowStep
  actionRun?: TInstallActionRun
  createdBy?: TAccount
  error: any
  isLoading: boolean
}

export const ActionRunStepDetails = ({
  step,
  actionRun,
  createdBy,
  error,
  isLoading,
}: IActionRunStepDetails) => {
  const isAdhoc = actionRun?.trigger_type === 'adhoc'

  if (isLoading && !actionRun) {
    return (
      <div className="flex flex-col gap-4">
        <ActionRunStepDetailsSkeleton />
      </div>
    )
  }

  if (error || !actionRun) {
    return (
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong" theme="error">
          Unable to load action run details
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <ActionRunHeader actionRun={actionRun} isAdhoc={isAdhoc} step={step} />

      <ActionRunMetadata
        actionRun={actionRun}
        createdBy={createdBy}
        step={step}
      />

      {isAdhoc ? (
        <AdhocActionDetails actionRun={actionRun} />
      ) : (
        <StandardActionSteps actionRun={actionRun} />
      )}

      <ActionRunLogs actionRun={actionRun} isAdhoc={isAdhoc} step={step} />
    </div>
  )
}

export const ActionRunStepDetailsSkeleton = () => {
  return (
    <>
      <ActionRunHeaderSkeleton />
      <ActionRunMetadataSkeleton />
      <StandardActionStepsSkeleton />
      <ActionRunLogsSkeleton />
    </>
  )
}

'use client'

import { useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { getInstallActionRun } from '@/lib'
import type { TInstallActionRun } from '@/types'
import { ActionRunHeader, ActionRunHeaderSkeleton } from './ActionRunHeader'
import {
  ActionRunMetadata,
  ActionRunMetadataSkeleton,
} from './ActionRunMetadata'
import { AdhocActionDetails } from './AdhocActionDetails'
import { StandardActionSteps, StandardActionStepsSkeleton } from './StandardActionSteps'
import { ActionRunLogs, ActionRunLogsSkeleton } from './ActionRunLogs'
import type { IActionRunDetails } from './types'

export const ActionRunStepDetails = ({ step }: IActionRunDetails) => {
  const { org } = useOrg()

  const {
    data: actionRun,
    error,
    isLoading,
  } = useQuery<TInstallActionRun>({
    queryKey: ['action-run', org?.id, step?.owner_id, step?.step_target_id],
    queryFn: () => getInstallActionRun({ orgId: org.id, installId: step.owner_id, runId: step.step_target_id }),
    enabled: !!org?.id && !!step?.owner_id && !!step?.step_target_id,
  })

  const isAdhoc = actionRun?.trigger_type === 'adhoc'
  const createdBy =
    actionRun?.created_by_id === step?.created_by_id
      ? step?.created_by
      : undefined

  if (isLoading && !actionRun) {
    return (
      <div className="flex flex-col gap-4">
        <ActionRunStepDetailsSkeleton />
      </div>
    )
  }

  if (error) {
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

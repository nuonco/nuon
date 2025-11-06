'use client'

import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TWorkflowStep, TSandboxRun } from '@/types'

export const SandboxRunApply = ({ step }: { step: TWorkflowStep }) => {
  const { org } = useOrg()
  const { data: sandboxRun, isLoading } = useQuery<TSandboxRun>({
    path: `/api/orgs/${org.id}/installs/${step?.owner_id}/sandbox/runs/${step?.step_target_id}`,
  })

  return (
    <>
      {isLoading || !sandboxRun ? (
        <SandboxRunApplySkeleton />
      ) : (
        <div className="flex items-start gap-6">
          <LabeledStatus
            label="Status"
            statusProps={{
              status: sandboxRun?.status_v2?.status,
            }}
            tooltipProps={{
              position: 'top',
              tipContent: sandboxRun?.status_v2?.status_human_description,
            }}
          />
        </div>
      )}
    </>
  )
}

export const SandboxRunApplySkeleton = () => {
  return (
    <div className="flex items-start gap-6">
      <LabeledValue label={<Skeleton height="17px" width="34px" />}>
        <Skeleton height="23px" width="75px" />
      </LabeledValue>
    </div>
  )
}

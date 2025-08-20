'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Text } from '@/components/Typography'
import type { TSandboxRun } from '@/types'
import { ApprovalStep } from './ApproveStep'
import type { IPollStepDetails } from './InstallWorkflowSteps'

export const SandboxStepDetails: FC<IPollStepDetails> = ({
  step,
  shouldPoll = false,
  pollDuration = 5000,
  workflowApproveOption,
}) => {
  const params = useParams<Record<'org-id', string>>()
  const orgId = params?.['org-id']
  const [sandboxRun, setData] = useState<TSandboxRun>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const fetchData = () => {
    fetch(
      `/api/${orgId}/installs/${step?.owner_id}/sandbox-runs/${step?.step_target_id}`
    ).then((r) =>
      r.json().then((res) => {
        setIsLoading(false)
        if (res?.error) {
          setError(res?.error?.error)
        } else {
          setError(undefined)
          setData(res.data)
        }
      })
    )
  }

  useEffect(() => {
    fetchData()
  }, [])

  useEffect(() => {
    if (shouldPoll) {
      const pollData = setInterval(fetchData, pollDuration)

      return () => clearInterval(pollData)
    }
  }, [shouldPoll])

  return (
    <>
    { step?.execution_type === 'approval' && step?.status?.status === 'auto-skipped' ? (
        <div className="flex flex-col gap-2">
        <Text variant="reg-14">Plan had no changes, skipping deployent.</Text>
        <br></br>
        </div>
    ) : null }
    {isLoading ? (
        <div className="border rounded-md p-6">
          <Loading loadingText="Loading sandobx details..." variant="stack" />
        </div>
      ) : (
        <>
          {error ? <Notice>{error}</Notice> : null}
          {step?.approval && step?.execution_type !== 'system' ? (
            <ApprovalStep
              approval={step?.approval}
              step={step}
              workflowId={step?.install_workflow_id}
              workflowApproveOption={workflowApproveOption}
            />
          ) : null}
          {sandboxRun ? (
            <div className="flex flex-col border rounded-md shadow">
              <div className="flex items-center justify-between p-3 border-b">
                <Text variant="med-14">
                  Install sandbox {sandboxRun?.run_type}
                </Text>
                <Link
                  className="text-sm gap-0"
                  href={`/${orgId}/installs/${step?.owner_id}/sandbox/${sandboxRun?.id}`}
                >
                  View details
                  <CaretRight />
                </Link>
              </div>
              <div className="p-6">
                <span className="flex gap-4 items-center">
                  <StatusBadge
                    description={
                      sandboxRun?.status_v2?.status_human_description ||
                      sandboxRun?.status_description
                    }
                    status={sandboxRun?.status_v2?.status || sandboxRun?.status}
                    label="Sandbox run status"
                  />
                </span>
              </div>
            </div>
          ) : null}
        </>
      )}
    </>
  )
}

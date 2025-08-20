'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Text } from '@/components/Typography'
import type { TInstallDeploy } from '@/types'
import { ApprovalStep } from './ApproveStep'
import type { IPollStepDetails } from './InstallWorkflowSteps'

export const DeployStepDetails: FC<IPollStepDetails> = ({
  step,
  shouldPoll = false,
  pollDuration = 5000,
  workflowApproveOption = 'prompt',
}) => {
  const params = useParams<Record<'org-id', string>>()
  const orgId = params?.['org-id']
  const [deploy, setData] = useState<TInstallDeploy>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const fetchData = () => {
    fetch(
      `/api/${orgId}/installs/${step?.owner_id}/deploys/${step?.step_target_id}`
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
            <Loading loadingText="Loading deploy details..." variant="stack" />
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
          {deploy ? (
            <div className="flex flex-col gap-8">
              <>
                {deploy ? (
                  <div className="flex flex-col border rounded-md shadow">
                    <div className="flex items-center justify-between p-3 border-b">
                      <Text variant="med-14">{deploy?.component_name}</Text>
                      <div className="flex items-center gap-4">
                        {deploy?.status !== 'queued' ? (
                          <Link
                            className="text-sm gap-0"
                            href={`/${orgId}/installs/${step?.owner_id}/components/${deploy?.component_id}`}
                          >
                            View component
                            <CaretRight />
                          </Link>
                        ) : null}
                        {deploy?.status !== 'queued' ? (
                          <Link
                            className="text-sm gap-0"
                            href={`/${orgId}/installs/${step?.owner_id}/components/${deploy?.component_id}/deploys/${deploy?.id}`}
                          >
                            View deployment
                            <CaretRight />
                          </Link>
                        ) : null}
                      </div>
                    </div>
                    <div className="p-6 flex flex-col gap-8">
                      <span className="flex gap-4 items-center">
                        <StatusBadge
                          description={
                            deploy?.status_v2?.status_human_description ||
                            deploy?.status
                          }
                          status={
                            deploy?.status_v2?.status ||
                            deploy?.status_description
                          }
                          label="Deployment status"
                        />
                      </span>
                    </div>
                  </div>
                ) : null}
              </>
            </div>
          ) : null}
        </>
      )}
    </>
  )
}

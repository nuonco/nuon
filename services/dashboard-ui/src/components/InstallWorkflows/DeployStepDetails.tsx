'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import type { TInstallDeploy } from '@/types'

import { IPollStepDetails } from './StepDetails'

export const DeployStepDetails: FC<IPollStepDetails> = ({
  step,
  shouldPoll = false,
  pollDuration = 5000,
}) => {
  const params = useParams<Record<'org-id', string>>()
  const orgId = params?.['org-id']
  const [deploy, setData] = useState<TInstallDeploy>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const fetchData = () => {
    fetch(
      `/api/${orgId}/installs/${step?.install_id}/deploys/${step?.step_target_id}`
    ).then((r) =>
      r.json().then((res) => {
        setIsLoading(false)
        if (res?.error) {
          setError(res?.error?.error)
        } else {
          setData(res.data)
        }
      })
    )
  }

  useEffect(() => {
    fetchData()
  }, [])

  useEffect(() => {
    setIsLoading(true)
    fetchData()
  }, [step])

  useEffect(() => {
    if (shouldPoll) {
      const pollData = setInterval(fetchData, pollDuration)

      return () => clearInterval(pollData)
    }
  }, [shouldPoll])

  return (
    <>
      {isLoading ? (
        <Loading loadingText="Loading deploy details..." variant="page" />
      ) : (
        <>
          {error ? <Notice>{error}</Notice> : null}
          {deploy ? (
            <div className="flex flex-col gap-8">
              <>
                {deploy ? (
                  <>
                    <span className="flex gap-4 items-center">
                      <StatusBadge
                        description={deploy?.status_description}
                        status={deploy?.status}
                      />
                      <Link
                        className="text-sm gap-0"
                        href={`/${orgId}/installs/${step?.install_id}/components/${deploy?.component_id}/deploys/${deploy?.id}`}
                      >
                        View details
                        <CaretRight />
                      </Link>
                    </span>
                  </>
                ) : null}
              </>
            </div>
          ) : null}
        </>
      )}
    </>
  )
}

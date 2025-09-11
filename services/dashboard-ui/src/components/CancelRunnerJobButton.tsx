'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { Check } from '@phosphor-icons/react'
import { Button, type IButton } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import { cancelRunnerJob } from '@/components/runner-actions'
import { useOrg } from '@/hooks/use-org'
import { trackEvent } from '@/utils'

export type TCancelJobType =
  | 'build'
  | 'deploy'
  | 'sandbox-run'
  | 'workflow-run'
  | 'sandbox'
  | 'actions'
  | 'sync'
  | 'operations'
  | 'runner'
  | 'health-checks'

type TCancelJobData = {
  buttonText: string
  confirmHeading: string
  confirmMessage: string
}

const cancelJobOptions: Record<TCancelJobType, TCancelJobData> = {
  build: {
    buttonText: 'Cancel build',
    confirmHeading: 'Cancel component build?',
    confirmMessage: 'Are you sure you want to cancel this component build?',
  },
  deploy: {
    buttonText: 'Cancel deploy',
    confirmHeading: 'Cancel component deployment?',
    confirmMessage:
      'Are you sure you want to cancel this component deployment?',
  },
  sync: {
    buttonText: 'Cancel sync',
    confirmHeading: 'Cancel component sync?',
    confirmMessage: 'Are you sure you want to cancel this component sync?',
  },
  'sandbox-run': {
    buttonText: 'Cancel sandbox job',
    confirmHeading: 'Cancel sandbox job?',
    confirmMessage: 'Are you sure you want to cancel this sandbox job?',
  },

  sandbox: {
    buttonText: 'Cancel sandbox job',
    confirmHeading: 'Cancel sandbox job?',
    confirmMessage: 'Are you sure you want to cancel this sandbox job?',
  },
  'workflow-run': {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel action workflow?',
    confirmMessage: 'Are you sure you want to cancel this action workflow?',
  },
  actions: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel action workflow?',
    confirmMessage: 'Are you sure you want to cancel this action workflow?',
  },
  operations: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel shut down?',
    confirmMessage: 'Are you sure you want to cancel this shut down job?',
  },
  runner: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel runner job?',
    confirmMessage: 'Are you sure you want to cancel this runner job?',
  },
  'health-checks': {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel health check?',
    confirmMessage: 'Are you sure you want to cancel this health check job?',
  },
}

interface ICancelRunnerJobButton extends IButton {
  jobType: TCancelJobType
  runnerJobId: string
  orgId: string
}

export const CancelRunnerJobButton: FC<ICancelRunnerJobButton> = ({
  jobType,
  runnerJobId,
  orgId,
  ...props
}) => {
  const { user } = useUser()
  const { org } = useOrg()
  const cancelJobData = cancelJobOptions[jobType]
  const pathName = usePathname()
  const [cancelError, setCancelError] = useState<string>()
  const [hasBeenCanceled, setHasBeenCanceled] = useState(false)
  const [isConfirmOpen, setIsConfirmOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  return (
    <>
      {isConfirmOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              isOpen={isConfirmOpen}
              heading={cancelJobData.confirmHeading}
              onClose={() => {
                setIsConfirmOpen(false)
              }}
            >
              <div className="mb-6">
                {cancelError ? (
                  <span className="flex w-full p-2 border rounded-md border-red-400 bg-red-300/20 text-red-800 dark:border-red-600 dark:bg-red-600/5 dark:text-red-600 text-base font-medium mb-6">
                    {cancelError}
                  </span>
                ) : null}
                <Text variant="reg-14" className="leading-relaxed">
                  {cancelJobData.confirmMessage}
                </Text>
              </div>
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsConfirmOpen(false)
                  }}
                  className="text-base"
                >
                  Cancel
                </Button>
                <Button
                  disabled={Boolean(cancelError)}
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    cancelRunnerJob({ orgId, runnerJobId, path: pathName })
                      .then(() => {
                        trackEvent({
                          event: 'runner_job_cancel',
                          status: 'ok',
                          user,
                          props: {
                            jobType,
                            orgId: org.id,
                            runnerJobId,
                          },
                        })
                        setIsLoading(false)
                        setIsKickedOff(true)
                        setIsConfirmOpen(false)
                        setHasBeenCanceled(true)
                      })
                      .catch((error) => {
                        trackEvent({
                          event: 'runner_job_cancel',
                          status: 'error',
                          user,
                          props: {
                            jobType,
                            orgId: org.id,
                            runnerJobId,
                          },
                        })
                        console.error(error?.message)
                        setIsLoading(false)
                        setCancelError(
                          'Error occured, please refresh page and try again.'
                        )
                      })
                  }}
                  variant="danger"
                >
                  {isKickedOff ? (
                    <Check size="16" />
                  ) : isLoading ? (
                    <SpinnerSVG />
                  ) : null}{' '}
                  {cancelJobData?.buttonText}
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        disabled={hasBeenCanceled}
        className="text-sm flex items-center gap-1 text-red-800 dark:text-red-500"
        onClick={() => {
          setIsConfirmOpen(true)
        }}
        {...props}
      >
        {cancelJobData?.buttonText}
      </Button>
    </>
  )
}

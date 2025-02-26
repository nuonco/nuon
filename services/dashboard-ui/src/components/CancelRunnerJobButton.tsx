'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { Check, XCircle } from '@phosphor-icons/react'
import { Button, type IButton } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import { cancelRunnerJob } from '@/components/runner-actions'

type TCancelJobType = 'build' | 'deploy' | 'sandbox-run' | 'workflow-run'

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
      'Are you sure you want to cancel this component depolyment?',
  },
  'sandbox-run': {
    buttonText: 'Cancel sandbox job',
    confirmHeading: 'Cancel sandbox job?',
    confirmMessage: 'Are you sure you want to cancel this sandbox job?',
  },
  'workflow-run': {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel action workflow?',
    confirmMessage: 'Are you sure you want to cancel this action workflow?',
  },
}

interface ICancelRunnerJobButton extends IButton {
  jobType: TCancelJobType
  runnerJobId: string
  orgId: string
  //onComplete?: () => void
}

export const CancelRunnerJobButton: FC<ICancelRunnerJobButton> = ({
  jobType,
  runnerJobId,
  orgId,
  ...props
}) => {
  const cancelJobData = cancelJobOptions[jobType]
  const pathName = usePathname()
  const [cancelError, setCancelError] = useState()
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
                  setIsLoading(false)
                  setIsKickedOff(true)
                  setIsConfirmOpen(false)
                  setHasBeenCanceled(true)
                  //if (props.onComplete) props.onComplete()
                })
                .catch((error) => {
                  setIsLoading(false)
                  setCancelError(
                    error?.message ||
                      'Error occured, please refresh page and try again.'
                  )
                })
            }}
            variant="primary"
          >
            {isKickedOff ? (
              <Check size="16" />
            ) : isLoading ? (
              <SpinnerSVG />
            ) : (
              <XCircle size="16" />
            )}{' '}
            Cancel
          </Button>
        </div>
      </Modal>
      <Button
        disabled={hasBeenCanceled}
        className="text-sm flex items-center gap-1 text-red-800 dark:text-red-500"
        onClick={() => {
          setIsConfirmOpen(true)
        }}
        {...props}
      >
        Cancel
      </Button>
    </>
  )
}

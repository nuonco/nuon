'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { Check, XCircle } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
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

interface ICancelRunnerJobButton {
  jobType: TCancelJobType
  runnerJobId: string
  orgId: string
  onComplete?: () => void
}

export const CancelRunnerJobButton: FC<ICancelRunnerJobButton> = ({
  jobType,
  runnerJobId,
  orgId,
  ...props
}) => {
  const cancelJobData = cancelJobOptions[jobType]
  const pathName = usePathname()
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
            className="text-sm flex items-center gap-1"
            onClick={() => {
              setIsLoading(true)
              cancelRunnerJob({ orgId, runnerJobId, path: pathName }).then(
                () => {
                  setIsLoading(false)
                  setIsKickedOff(true)
                  setIsConfirmOpen(false)
                  if (props.onComplete) props.onComplete()
                }
              )
            }}
            variant="primary"
          >
            {isKickedOff ? (
              <Check size="18" />
            ) : isLoading ? (
              <SpinnerSVG />
            ) : (
              <XCircle size="18" />
            )}{' '}
            {cancelJobData.buttonText}
          </Button>
        </div>
      </Modal>
      <Button
        className="text-sm flex items-center gap-1"
        onClick={() => {
          setIsConfirmOpen(true)
        }}
      >
        {cancelJobData.buttonText}
      </Button>
    </>
  )
}

'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { CodeBlock } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { CodeViewer, JsonView } from '@/components/Code'
import { useOrg } from '@/hooks/use-org'

interface IRunnerJobPlanModal {
  buttonText?: string
  headingText?: string
  runnerJobId: string
}

export const RunnerJobPlanModal: FC<IRunnerJobPlanModal> = ({
  buttonText = 'View job plan',
  headingText = 'Runner job plan',
  runnerJobId,
}) => {
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [plan, setPlan] = useState()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState()

  useEffect(() => {
    if (isOpen) {
      fetch(`/api/${org?.id}/runner-jobs/${runnerJobId}/plan`).then((r) =>
        r.json().then((res) => {
          setIsLoading(false)
          if (res?.error) {
            setError(res?.error?.error || 'Unable to fetch job plan')
          } else {
            setError(undefined)
            setPlan(res.data)
          }
        })
      )
    }
  }, [isOpen])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-4xl"
              heading={headingText}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4">
                {error ? <Notice>{error}</Notice> : null}
                {isLoading ? (
                  <Loading loadingText="Loading job plan..." variant="stack" />
                ) : (
                  <JsonView data={plan} />
                )}
              </div>
              <div className="mt-4 flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-sm"
                >
                  Close
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3 w-fit"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <CodeBlock size="16" />
        {buttonText}
      </Button>
    </>
  )
}

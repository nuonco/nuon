'use client'

import { useState } from 'react'
import { createPortal } from 'react-dom'
import { CodeBlockIcon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { JsonView } from '@/components/Code'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TRunnerJobPlan } from '@/types'

interface IRunnerJobPlanModal {
  buttonText?: string
  headingText?: string
  runnerJobId: string
}

export const RunnerJobPlanModal = ({
  buttonText = 'View job plan',
  headingText = 'Runner job plan',
  runnerJobId,
}: IRunnerJobPlanModal) => {
  const [isOpen, setIsOpen] = useState(false)

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
              <RunnerJobPlan runnerJobId={runnerJobId} />
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
        <CodeBlockIcon size="16" />
        {buttonText}
      </Button>
    </>
  )
}

// TODO(nnnat): temp until we update to the stratus modal style
const RunnerJobPlan = ({ runnerJobId }: { runnerJobId: string }) => {
  const { org } = useOrg()
  const {
    data: plan,
    error,
    isLoading,
  } = useQuery<TRunnerJobPlan>({
    // TODO(nnnnat): remove once the endpoint is fixed with the correct content-type
    path: `/api/${org.id}/runner-jobs/${runnerJobId}/plan`,
  })
  return (
    <div className="flex flex-col gap-4">
      {error ? <Notice>{error?.error}</Notice> : null}
      {isLoading ? (
        <Loading loadingText="Loading job plan..." variant="stack" />
      ) : (
        <JsonView data={plan} />
      )}
    </div>
  )
}

'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { CodeBlock } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Modal } from '@/components/Modal'
import { CodeViewer } from '@/components/Code'

interface IRunnerJobPlanModal {
  buttonText?: string
  headingText?: string
  plan: string
}

export const RunnerJobPlanModal: FC<IRunnerJobPlanModal> = ({
  buttonText = 'View job plan',
  headingText = 'Runner job plan',
  plan,
}) => {
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
              <div className="flex flex-col gap-4">
                <CodeViewer
                  initCodeSource={JSON.stringify(plan, null, 2)}
                  language="json"
                />
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

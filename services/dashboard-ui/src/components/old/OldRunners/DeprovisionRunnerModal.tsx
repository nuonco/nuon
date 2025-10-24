'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { BoxArrowDownIcon } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Modal } from '@/components/old/Modal'
import { Text } from '@/components/old/Typography'

interface IDeprovisionRunnerModal {
  buttonText?: string
  headingText?: string
}

export const DeprovisionRunnerModal: FC<IDeprovisionRunnerModal> = ({
  buttonText = 'Deprovision runner',
  headingText = 'Deprovision runner information',
}) => {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-xl"
              heading={headingText}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4">
                <Text>
                  You can use the shut down button to restart your runner during
                  normal operation. If you need to forcefully terminate your
                  runner, you can terminate the instance directly from the
                  AutoScaling group in your AWS account.
                </Text>
                <Text>
                  Deleting the instance has the chance to lose any state of
                  in-flight jobs, but in other cases is a safe operation.
                </Text>
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
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full text-red-800 dark:text-red-500"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <BoxArrowDownIcon size="16" />
        {buttonText}
      </Button>
    </>
  )
}

'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { CodeBlock } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Modal } from '@/components/old/Modal'
import { CodeViewer } from '@/components/old/Code'

interface IValuesFileModal {
  buttonText?: string
  headingText?: string
  valuesFiles: Array<string>
}

export const ValuesFileModal: FC<IValuesFileModal> = ({
  buttonText = 'View variable files',
  headingText = 'Terraform variable files',
  valuesFiles,
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
                {valuesFiles.map((vf, i) => (
                  <CodeViewer key={i} initCodeSource={vf} language="hcl" />
                ))}
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
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-fit"
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

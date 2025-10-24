'use client'

import { useState } from 'react'
import { createPortal } from 'react-dom'
import { StackMinusIcon } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { useInstall } from '@/hooks/use-install'

export const DeprovisionStackModal = () => {
  const { install } = useInstall()
  const [isOpen, setIsOpen] = useState(false)

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-xl"
              isOpen={isOpen}
              heading={
                <span className="flex items-center gap-3">
                  Deprovision stack for {install.name}?
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-6">
                <Notice variant="warn">
                  Once you have deprovisioned the install from the UI, please go
                  to the cloud platform console and destroy this stack for your
                  install.
                </Notice>
              </div>
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-base"
                >
                  Cancel
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}

      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <StackMinusIcon size="16" />
        Deprovision stack
      </Button>
    </>
  )
}

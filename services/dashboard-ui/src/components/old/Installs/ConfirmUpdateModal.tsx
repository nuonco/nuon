'use client'

import React, { type FC, use, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { WarningIcon } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Modal } from '@/components/old/Modal'
import { Text } from '@/components/old/Typography'

import type { TInstall } from '@/types'

interface IConfirmUpdateModal {
  install: TInstall
  isOpen: boolean
  onClose: (isConfirmed: boolean) => void
}

export const ConfirmUpdateModal: FC<IConfirmUpdateModal> = ({
  install,
  isOpen: isUpOpen,
  onClose,
}) => {
  const isInstallManagedByConfig =
    install?.metadata &&
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  const [isOpen, setIsOpen] = useState(isUpOpen)

  useEffect(() => {
    setIsOpen(isUpOpen)
  }, [isUpOpen])

  useEffect(() => {
    if (!isInstallManagedByConfig && isOpen) {
      setIsOpen(false)
      onClose(true)
    }
  }, [isOpen])

  if (!isInstallManagedByConfig) {
    return <></>
  }

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-xl"
              heading={
                <span className="flex gap-2 text-orange-800">
                  <WarningIcon size={24} />
                  <Text variant="med-18">
                    Override changes to this install?
                  </Text>
                </span>
              }
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
                onClose(false)
              }}
            >
              <div className="flex flex-col gap-2 mb-6">
                <span className="flex flex-col gap-2 mb-12">
                  <Text variant="med-14" className="mb-2">
                    You are about update an Install managed by a Config file.
                  </Text>
                  <Text variant="reg-14">
                    If you proceed, the config file and install state will no
                    longer be in sync. Are you sure you want to continue?
                  </Text>
                </span>
                <Text variant="med-14">
                  Tip: To revert this override later, please sync the install
                  config file again with the CLI.
                </Text>
              </div>
              <div className="flex gap-3 justify-between">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                    onClose(false)
                  }}
                  className="text-sm"
                >
                  Cancel
                </Button>
                <Button
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsOpen(false)
                    onClose(true)
                  }}
                  variant="primary"
                >
                  Confirm override
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
    </>
  )
}

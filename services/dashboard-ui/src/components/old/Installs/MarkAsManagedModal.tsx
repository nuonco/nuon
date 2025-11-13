'use client'

import { useState } from 'react'
import { createPortal } from 'react-dom'
import {
  ToggleLeftIcon,
  ToggleRightIcon,
  FileCloudIcon,
} from '@phosphor-icons/react'
import { updateInstall } from '@/actions/installs/update-install'
import { Button } from '@/components/old/Button'
import { SpinnerSVG } from '@/components/old/Loading'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { Text } from '@/components/old/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const MarkAsManagedModal = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState("")

  const hasManagedBy = Boolean(install?.metadata?.managed_by)
  const isManagedByConfig =
    hasManagedBy && install?.metadata?.managed_by === 'nuon/cli/install-config'
  const buttonText = isManagedByConfig ? (
    <>Disable Install Config Sync</>
  ) : (
    <>Enable Install Config Sync</>
  )
  const buttonIcon = isManagedByConfig ? (
    <ToggleRightIcon size="18" />
  ) : (
    <ToggleLeftIcon size="18" />
  )

  const handleUpdateInstallError = (err) => {
    setIsLoading(false)
    setError(err?.message || 'Unable to mark install managed by config')
  }
  const handleManagedByChange = ({ data, error }) => {
    setIsLoading(false)
    if (error) {
      setError(error?.error)
      console.error(error)
    } else {
      setError(undefined)
      setIsOpen(false)
    }
  }
  const toggleManagedBy = () => {
    setIsLoading(true)

    updateInstall({
      body: {
        metadata: {
          managed_by: isManagedByConfig
            ? 'nuon/dashboard'
            : 'nuon/cli/install-config',
        },
      },
      installId: install.id,
      orgId: org.id,
    })
      .then(handleManagedByChange)
      .catch(handleUpdateInstallError)
  }

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading={
                isManagedByConfig
                  ? 'Disable Install Config Sync?'
                  : 'Enable Install Config Sync?'
              }
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  This Install can be managed via an Install Config file only
                  after marking it as managed by Install Config.
                </Text>
                <Text variant="reg-14" className="leading-relaxed">
                  {isManagedByConfig
                    ? ' Disabling this will stop any future syncs from the Install Config file.'
                    : ' Enable this to allow syncing from an Install Config file.'}
                </Text>
              </div>
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-sm"
                >
                  Cancel
                </Button>
                <Button
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    toggleManagedBy()
                  }}
                  variant="primary"
                >
                  {isLoading ? (
                    <SpinnerSVG />
                  ) : (
                    buttonIcon
                  )}{' '}
                  {buttonText}
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
        <FileCloudIcon size="18" /> {buttonText}
      </Button>
    </>
  )
}

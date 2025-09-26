'use client'

import { usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import {
  CheckCircleIcon,
  ToggleLeftIcon,
  ToggleRightIcon,
} from '@phosphor-icons/react'
import { createInstallConfig } from '@/actions/installs/create-install-config'
import { updateInstallConfig } from '@/actions/installs/update-install-config'
import { updateInstall } from '@/actions/installs/update-install'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { ConfirmUpdateModal } from './ConfirmUpdateModal'

export const AutoApproveModal = () => {
  const path = usePathname()
  const { org } = useOrg()
  const { install } = useInstall()
  const [isOpen, setIsOpen] = useState(false)
  const [isConfirmOpen, setIsConfirmOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState()

  const hasInstallConfig = Boolean(install?.install_config)
  const isApproveAll =
    hasInstallConfig &&
    install?.install_config?.approval_option === 'approve-all'
  const buttonText = isApproveAll ? (
    <>Disable auto approval</>
  ) : (
    <>Enable auto approval</>
  )
  const buttonIcon = isApproveAll ? (
    <ToggleRightIcon size="18" />
  ) : (
    <ToggleLeftIcon size="18" />
  )

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  const handleApprovalOptionError = (err) => {
    setIsLoading(false)
    setError(err?.message || 'Unable to set approval option')
  }
  const handleApprovalOptionChange = ({ data, error }) => {
    setIsLoading(false)
    if (error) {
      setError(error?.error)
      console.error(error)
    } else {
      setError(undefined)
      setIsOpen(false)
    }
  }
  const toggleApprovalOption = () => {
    setIsLoading(true)

    updateInstallConfig({
      body: { approval_option: isApproveAll ? 'prompt' : 'approve-all' },
      installConfigId: install?.install_config?.id,
      installId: install?.id,
      orgId: org?.id,
      path,
    })
      .then(handleApprovalOptionChange)
      .catch(handleApprovalOptionError)

    if (install?.metadata?.managed_by === 'nuon/cli/install-config') {
      updateInstall({
        body: { metadata: { managed_by: 'nuon/dashboard' } },
        installId: install.id,
        orgId: org.id,
      })
    }
  }

  const createApprovalOption = () => {
    createInstallConfig({
      body: { approval_option: 'approve-all' },
      installId: install?.id,
      orgId: org?.id,
      path,
    })
      .then(handleApprovalOptionChange)
      .catch(handleApprovalOptionError)
  }

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading="Auto approve changes?"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to auto approve changes to this install?
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
                    if (hasInstallConfig) {
                      toggleApprovalOption()
                    } else {
                      createApprovalOption()
                    }
                  }}
                  variant="primary"
                >
                  {isKickedOff ? (
                    <CheckCircleIcon size="18" />
                  ) : isLoading ? (
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
      <ConfirmUpdateModal
        install={install}
        isOpen={isConfirmOpen}
        onClose={(isConfirmed) => {
          setIsOpen(isConfirmed)
          setIsConfirmOpen(false)
        }}
      />
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        variant="ghost"
        onClick={() => {
          setIsConfirmOpen(true)
        }}
      >
        {buttonIcon} {buttonText}
      </Button>
    </>
  )
}

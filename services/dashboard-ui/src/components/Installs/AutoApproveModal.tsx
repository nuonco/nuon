'use client'

import classNames from 'classnames'
import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckCircle, ToggleLeft, ToggleRight } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import {
  createInstallConfig,
  updateInstallConfig,
  updateInstallManagedBy,
} from '@/components/install-actions'
import { useOrg } from '@/hooks/use-org'
import type { TInstall } from '@/types'
import { trackEvent } from '@/utils'
import { ConfirmUpdateModal } from './ConfirmUpdateModal'

interface IAutoApproveModal {
  install: TInstall
}

export const AutoApproveModal: FC<IAutoApproveModal> = ({ install }) => {
  const { org } = useOrg()
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
    <ToggleRight size="18" />
  ) : (
    <ToggleLeft size="18" />
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
    if (isApproveAll) {
      updateInstallConfig({
        approvalOption: 'prompt',
        configId: install?.install_config?.id,
        installId: install?.id,
        orgId: org?.id,
      })
        .then(handleApprovalOptionChange)
        .catch(handleApprovalOptionError)
    } else {
      updateInstallConfig({
        approvalOption: 'approve-all',
        configId: install?.install_config?.id,
        installId: install?.id,
        orgId: org?.id,
      })
        .then(handleApprovalOptionChange)
        .catch(handleApprovalOptionError)
    }
    updateInstallManagedBy({
      installId: install?.id,
      orgId: org?.id,
      managedBy: install?.metadata?.managed_by,
    })
  }

  const createApprovalOption = () => {
    createInstallConfig({
      approvalOption: 'approve-all',
      installId: install?.id,
      orgId: org?.id,
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
                    <CheckCircle size="18" />
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

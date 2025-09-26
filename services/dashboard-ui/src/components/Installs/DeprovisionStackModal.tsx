'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Check, StackMinus } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { trackEvent } from '@/utils'
import type { TInstall } from '@/types'

interface IDeprovisionStackModal {
  install: TInstall
  orgId: string
}

export const DeprovisionStackModal: FC<IDeprovisionStackModal> = ({
  install,
  orgId,
}) => {
  const router = useRouter()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState()

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

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
                {error ? <Notice>{error}</Notice> : null}
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
                {/* <Button
                    className="text-sm flex items-center gap-1"
                    onClick={() => {
                    setIsLoading(true)
                    forgetInstall({ installId: install.id, orgId })
                    .then(() => {
                    trackEvent({
                    event: 'install_forget',
                    user,
                    status: 'ok',
                    props: {
                    orgId,
                    installId: install?.id,
                    },
                    })
                    router.push(`/${orgId}/installs`)
                    setIsLoading(false)
                    setIsKickedOff(true)
                    })
                    .catch((err) => {
                    trackEvent({
                    event: 'install_forget',
                    user,
                    status: 'error',
                    props: {
                    orgId,
                    installId: install?.id,
                    err,
                    },
                    })
                    setError(
                    err?.message ||
                    'Error occured, please refresh page and try again.'
                    )
                    setIsLoading(false)
                    })
                    }}
                    variant="danger"
                    >
                    {isKickedOff ? (
                    <Check size="18" />
                    ) : isLoading ? (
                    <SpinnerSVG />
                    ) : (
                    <StackMinus size="18" />
                    )}{' '}
                    Deprovision stack
                    </Button> */}
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
        <StackMinus size="16" />
        Deprovision stack
      </Button>
    </>
  )
}

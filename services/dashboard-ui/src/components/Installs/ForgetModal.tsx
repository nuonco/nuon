'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Check, Trash } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { forgetInstall } from '@/components/install-actions'
import type { TInstall } from '@/types'

interface IForgetModal {
  install: TInstall
  orgId: string
}

export const ForgetModal: FC<IForgetModal> = ({ install, orgId }) => {
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
              className="max-w-lg"
              isOpen={isOpen}
              heading={
                <span className="flex items-center gap-3">
                  Forget {install.name}?
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-4 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Notice>
                  This should only be used in cases where an install was broken
                  in an unordinary way and needs to be manually removed.
                </Notice>
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to forget {install?.name}? <br /> This
                  action will remove the install and can not be undone.
                </Text>
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
                <Button
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    forgetInstall({ installId: install.id, orgId })
                      .then(() => {
                        router.push(`/${orgId}/installs`)
                        setIsLoading(false)
                        setIsKickedOff(true)
                      })
                      .catch((err) => {
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
                    <Trash size="18" />
                  )}{' '}
                  Forget install
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}

      <Button
        className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3 !rounded-none w-full text-red-800 dark:text-red-500"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <Trash size="18" />
        Forget install
      </Button>
    </>
  )
}

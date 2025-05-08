'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Check, Trash } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Input } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { forgetInstall } from '@/components/install-actions'
import { trackEvent } from '@/utils'
import type { TInstall } from '@/types'

interface IForgetModal {
  install: TInstall
  orgId: string
}

export const ForgetModal: FC<IForgetModal> = ({ install, orgId }) => {
  const { user } = useUser()
  const router = useRouter()
  const [confirm, setConfirm] = useState<string>()
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
                  Forget {install.name}
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-8 mb-12">
                {error ? <Notice>{error}</Notice> : null}
                <Notice>
                  This should only be used in cases where an install was broken
                  in an unordinary way and needs to be manually removed.
                </Notice>
                <span className="flex flex-col gap-1">
                  <Text variant="med-18" className="leading-relaxed">
                    Are you sure you want to forget {install?.name}?
                  </Text>
                  <Text
                    className="text-cool-grey-600 dark:text-white/70"
                    variant="reg-12"
                  >
                    This action will remove the install and can not be undone.
                  </Text>
                </span>

                <div className="flex flex-col gap-2">
                  <Text variant="reg-14">
                    You should only do this after you have:
                  </Text>

                  <ul className="flex flex-col gap-1 list-disc pl-4">
                    <li className="text-sm">
                      Successfully deprovisioned the install
                    </li>
                    <li className="text-sm">
                      Deprovisioned the CloudFormation stack for this install
                    </li>
                  </ul>
                </div>

                <div className="w-full">
                  <label className="flex flex-col gap-1 w-full">
                    <Text variant="med-14">
                      To verify, type{' '}
                      <span className="text-red-800 dark:text-red-500">
                        {install.name}
                      </span>{' '}
                      below.
                    </Text>
                    <Input
                      placeholder="install name"
                      className="w-full"
                      type="text"
                      value={confirm}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        setConfirm(e?.currentTarget?.value)
                      }}
                    />
                  </label>
                </div>
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
                  disabled={confirm !== install.name}
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
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full text-red-800 dark:text-red-500"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <Trash size="16" />
        Forget install
      </Button>
    </>
  )
}

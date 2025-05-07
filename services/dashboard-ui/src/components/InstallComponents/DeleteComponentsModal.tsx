'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Check, TrashSimple } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { CheckboxInput, Input } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { deleteComponents } from '@/components/install-actions'
import { trackEvent } from '@/utils'

interface IDeleteComponentsModal {
  installId: string
  orgId: string
}

export const DeleteComponentsModal: FC<IDeleteComponentsModal> = ({
  installId,
  orgId,
}) => {
  const router = useRouter()
  const { user } = useUser()
  const [confirm, setConfirm] = useState<string>()
  const [force, setForceDelete] = useState(false)
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
              className="!max-w-2xl"
              isOpen={isOpen}
              heading={`Teardown all components`}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-6 mb-12">
                {error ? <Notice>{error}</Notice> : null}
                <span className="flex flex-col gap-1">
                  <Text variant="med-18">
                    Are you sure you want to teardown all components?
                  </Text>
                  <Text variant="reg-12">
                    Tearing down components will affect the working nature of this
                    install.
                  </Text>
                </span>
                <Notice>
                  Warning, this action is not reversible. Please be certain.
                </Notice>

                <div className="w-full">
                  <label className="flex flex-col gap-1 w-full">
                    <Text variant="med-14">
                      To verify, type{' '}
                      <span className="text-red-800 dark:text-red-500">
                        teardown
                      </span>{' '}
                      below.
                    </Text>
                    <Input
                      placeholder="teardown"
                      className="w-full"
                      type="text"
                      value={confirm}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        setConfirm(e?.currentTarget?.value)
                      }}
                    />
                  </label>
                </div>
                <div className="flex items-start">
                  <CheckboxInput
                    name="ack"
                    defaultChecked={force}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setForceDelete(Boolean(e?.currentTarget?.checked))
                    }}
                    className="mt-1.5"
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px] !items-start"
                    labelText={
                      <span className="flex flex-col gap2">
                        <Text variant="med-14">Force teardown</Text>
                        <Text className="!font-normal" variant="reg-12">
                          Force tearing down may result in orphaned artifacts that
                          will need manual removal.
                        </Text>
                      </span>
                    }
                  />
                </div>
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
                  disabled={confirm !== 'teardown'}
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    deleteComponents({ installId, orgId, force })
                      .then((workflowId) => {
                        trackEvent({
                          event: 'components_teardown',
                          user,
                          status: 'ok',
                          props: { orgId, installId },
                        })
                        setForceDelete(false)
                        setIsLoading(false)
                        setIsKickedOff(true)

                        if (workflowId) {
                          router.push(
                            `/${orgId}/installs/${installId}/history/${workflowId}`
                          )
                        } else {
                          router.push(`/${orgId}/installs/${installId}/history`)
                        }

                        setIsOpen(false)
                      })
                      .catch((err) => {
                        trackEvent({
                          event: 'components_teardown',
                          user,
                          status: 'error',
                          props: { orgId, installId, err },
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
                    <TrashSimple size="18" />
                  )}{' '}
                  Teardown all components
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-fit text-red-800 dark:text-red-500"        
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <TrashSimple size="16" /> Teardown all components
      </Button>
    </>
  )
}

'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { Check, ArrowsClockwise } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { CheckboxInput } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { syncSecrets } from '@/components/install-actions'
import { trackEvent } from '@/utils'

interface ISyncSecretsModal {
  installId: string
  orgId: string
}

export const SyncSecretsModal: FC<ISyncSecretsModal> = ({
  installId,
  orgId,
}) => {
  const router = useRouter()
  const { user } = useUser()
  const [isOpen, setIsOpen] = useState(false)
  const [planOnly, setPlanOnly] = useState(false)
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
              heading="Sync secrets?"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to sync secrets for this install?
                </Text>
                {/* <CheckboxInput
                    name="ack"
                    defaultChecked={planOnly}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setPlanOnly(Boolean(e?.currentTarget?.checked))
                    }}
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px]"
                    labelText={'Only create a sync secrets plan?'}
                    /> */}
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
                    setIsLoading(true)
                    syncSecrets({
                      installId,
                      orgId,
                      planOnly,
                    })
                      .then((workflowId) => {
                        trackEvent({
                          event: 'install_sync_secrets',
                          user,
                          status: 'ok',
                          props: { orgId, installId },
                        })
                        setIsLoading(false)
                        setIsKickedOff(true)

                        if (workflowId) {
                          router.push(
                            `/${orgId}/installs/${installId}/workflows/${workflowId}`
                          )
                        } else {
                          router.push(
                            `/${orgId}/installs/${installId}/workflows`
                          )
                        }

                        setIsOpen(false)
                      })
                      .catch((err) => {
                        trackEvent({
                          event: 'install_sync_secrets',
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
                  variant="primary"
                >
                  {isKickedOff ? (
                    <Check size="18" />
                  ) : isLoading ? (
                    <SpinnerSVG />
                  ) : (
                    <ArrowsClockwise size="18" />
                  )}{' '}
                  Sync secrets
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
        <ArrowsClockwise size="16" />
        Sync secrets
      </Button>
    </>
  )
}

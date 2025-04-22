'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useRouter } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Check, StackSimple } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { CheckboxInput } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { reprovisionSandbox } from '@/components/install-actions'
import { trackEvent } from '@/utils'

interface IReprovisionSandboxModal {
  installId: string
  orgId: string
}

export const ReprovisionSandboxModal: FC<IReprovisionSandboxModal> = ({
  installId,
  orgId,
}) => {
  const router = useRouter()
  const { user } = useUser()
  const [force, setForce] = useState(false)
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
              heading="Reprovision sandbox?"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to reprovision this sandbox?
                </Text>
                <div className="flex items-start">
                  <CheckboxInput
                    name="ack"
                    defaultChecked={force}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setForce(Boolean(e?.currentTarget?.checked))
                    }}
                    className="mt-1.5"
                    labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px] !items-start"
                    labelText={
                      <span className="flex flex-col gap2">
                        <Text variant="med-14">Continue on error</Text>
                        <Text className="!font-normal" variant="reg-12">
                          Continue running workflow steps if one fails.
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
                  className="text-sm flex items-center gap-1"
                  onClick={() => {
                    setIsLoading(true)
                    reprovisionSandbox({
                      installId,
                      orgId,
                      continueOnError: force,
                    })
                      .then((workflowId) => {
                        trackEvent({
                          event: 'install_sandbox_reprovision',
                          user,
                          status: 'ok',
                          props: { orgId, installId },
                        })
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
                          event: 'install_sandbox_reprovision',
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
                    <StackSimple size="18" />
                  )}{' '}
                  Reprovision sandbox
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
        <StackSimple size="16" />
        Reprovision sandbox
      </Button>
    </>
  )
}

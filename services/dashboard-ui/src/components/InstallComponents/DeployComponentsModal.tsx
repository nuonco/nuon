'use client'

import { useRouter } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CloudArrowUp, CloudCheck } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { CheckboxInput } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { deployComponents } from '@/components/install-actions'
import { trackEvent } from '@/utils'

interface IDeployComponentsModal {
  installId: string
  orgId: string
}

export const DeployComponentsModal: FC<IDeployComponentsModal> = ({
  installId,
  orgId,
}) => {
  const router = useRouter()
  const { user } = useUser()
  const [planOnly, setPlanOnly] = useState(false)
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
              heading="Deploy all components"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to deploy components? This will deploy
                  all components to this install.
                </Text>
                <CheckboxInput
                  name="ack"
                  defaultChecked={planOnly}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setPlanOnly(Boolean(e?.currentTarget?.checked))
                  }}
                  labelClassName="hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 gap-4 max-w-[300px]"
                  labelText={'Plan Only?'}
                />
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
                    deployComponents({
                      installId,
                      orgId,
                      planOnly,
                    })
                      .then((workflowId) => {
                        trackEvent({
                          event: 'components_deploy',
                          status: 'ok',
                          user,
                          props: {
                            installId,
                            orgId,
                          },
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
                          event: 'components_deploy',
                          status: 'error',
                          user,
                          props: {
                            installId,
                            orgId,
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
                  variant="primary"
                >
                  {isKickedOff ? (
                    <CloudCheck size="18" />
                  ) : isLoading ? (
                    <SpinnerSVG />
                  ) : (
                    <CloudArrowUp size="18" />
                  )}{' '}
                  Deploy all components
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-fit"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <CloudArrowUp size="16" />
        Deploy all components
      </Button>
    </>
  )
}

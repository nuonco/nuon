'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { CloudArrowUp, CloudCheck } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { deployComponents } from '@/components/install-actions'

interface IDeployComponentsModal {
  installId: string
  orgId: string
}

export const DeployComponentsModal: FC<IDeployComponentsModal> = ({
  installId,
  orgId,
}) => {
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
              heading="Deploy all components?"
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
                    deployComponents({ installId, orgId })
                      .then(() => {
                        setIsLoading(false)
                        setIsKickedOff(true)
                        setIsOpen(false)
                      })
                      .catch((err) => {
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
                  Deploy components
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
        <CloudArrowUp size="16" />
        Deploy components
      </Button>
    </>
  )
}

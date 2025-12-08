'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0'
import { CheckIcon, HammerIcon } from '@phosphor-icons/react'
import { buildComponents } from '@/actions/apps/build-components'
import { Button } from '@/components/old/Button'
import { SpinnerSVG } from '@/components/old/Loading'
import { Modal } from '@/components/old/Modal'
import { Notice } from '@/components/old/Notice'
import { Text, ID } from '@/components/old/Typography'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import type { TAPIError } from '@/types'
import { trackEvent } from '@/lib/segment-analytics'

export const BuildAllComponentsButton: FC = () => {
  const path = usePathname()
  const { user } = useUser()
  const { org } = useOrg()
  const { app } = useApp()
  const [isOpen, setIsOpen] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | TAPIError>()

  const handleClose = () => {
    setIsOpen(false)
  }

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)
    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 30000)
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
              heading={`Build all components?`}
              onClose={handleClose}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? (
                  <Notice>
                    {typeof error === 'string' ? error : error.error ||
                      'Unable to kick off component builds, please refresh page and try again.'}
                  </Notice>
                ) : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to build all components?
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
                    setIsKickedOff(true)
                    buildComponents({
                      appId: app.id,
                      orgId: org.id,
                      path,
                    }).then((res) => {
                      setIsLoading(false)
                      if (res.error) {
                        trackEvent({
                          event: 'components_build',
                          user,
                          status: 'error',
                          props: {
                            appId: app.id,
                            orgId: org.id,
                          },
                        })

                        setError(
                          res.error ||
                            'Unable to kick off component builds, please refresh page and try again.'
                        )
                      } else {
                        trackEvent({
                          event: 'components_build',
                          user,
                          status: 'ok',
                          props: {
                            appId: app.id,
                            orgId: org.id,
                          },
                        })
                        handleClose()
                      }
                    })
                  }}
                  variant="primary"
                >
                  {isLoading ? (
                    <SpinnerSVG />
                  ) : isKickedOff ? (
                    <CheckIcon size="18" />
                  ) : (
                    <HammerIcon size="18" />
                  )}{' '}
                  Build all components
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm flex items-center gap-1"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        Build all components
      </Button>
    </>
  )
}

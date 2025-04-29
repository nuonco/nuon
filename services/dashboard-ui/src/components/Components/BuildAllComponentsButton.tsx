'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Check, Hammer } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text, ID } from '@/components/Typography'
import { buildComponents } from '@/components/app-actions'
import type { TComponent } from '@/types'
import { trackEvent, type TQueryError } from '@/utils'

export const BuildAllComponentsButton: FC<{
  components: Array<TComponent>
}> = ({ components }) => {
  const { user } = useUser()
  const params = useParams<{ 'org-id': string; 'app-id': string }>()
  const appId = params?.['app-id']
  const orgId = params?.['org-id']
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState<string | Array<TQueryError>>()

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
              heading={`Build all components?`}
              onClose={() => {
                setError(undefined)
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? (
                  Array.isArray(error) ? (
                    <div className="flex flex-col gap-1">
                      {error?.map((err) => (
                        <Notice key={err.meta?.id}>
                          {err?.meta?.name && err?.meta?.id ? (
                            <span className="flex gap-2">
                              <Text variant="med-14">{err?.meta?.name}:</Text>
                              <ID
                                className="!text-current"
                                id={err?.meta?.id}
                              />
                            </span>
                          ) : null}
                          <Text>{err?.error}</Text>
                        </Notice>
                      ))}
                    </div>
                  ) : (
                    <Notice>{error}</Notice>
                  )
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
                    buildComponents({
                      appId,
                      components: components,
                      orgId,
                    }).then((res) => {
                      if (res?.some((r) => r?.error)) {
                        trackEvent({
                          event: 'components_build',
                          user,
                          status: 'error',
                          props: {
                            appId,
                            orgId,
                          },
                        })

                        setError(
                          res.filter((r) => r?.error).map((r) => r?.error) ||
                            'Unable to kick off component builds, please refresh page and try again.'
                        )
                        setIsLoading(false)
                      } else {
                        trackEvent({
                          event: 'components_build',
                          user,
                          status: 'ok',
                          props: {
                            appId,
                            orgId,
                          },
                        })
                        setIsLoading(false)
                        setIsKickedOff(true)
                        setIsOpen(false)
                      }
                    })
                  }}
                  variant="primary"
                >
                  {isKickedOff ? (
                    <Check size="18" />
                  ) : isLoading ? (
                    <SpinnerSVG />
                  ) : (
                    <Hammer size="18" />
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

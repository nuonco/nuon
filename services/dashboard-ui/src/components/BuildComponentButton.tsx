'use client'

import { useRouter, useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Check, Hammer } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { createComponentBuild } from '@/components/app-actions'
import { trackEvent } from '@/utils'

export const BuildComponentButton: FC<{
  componentName: string
}> = ({ componentName }) => {
  const { user } = useUser()
  const router = useRouter()
  const params = useParams<{
    'org-id': string
    'app-id': string
    'component-id': string
  }>()
  const appId = params?.['app-id']
  const orgId = params?.['org-id']
  const componentId = params?.['component-id']
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState<string>()

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
              heading={`Build ${componentName} component?`}
              onClose={() => {
                setError(undefined)
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                {error ? <Notice>{error}</Notice> : null}
                <Text variant="reg-14" className="leading-relaxed">
                  Are you sure you want to build {componentName}?
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
                    createComponentBuild({ componentId, orgId }).then((r) => {
                      if (r?.error) {
                        trackEvent({
                          event: 'component_build',
                          user,
                          status: 'error',
                          props: {
                            orgId,
                            appId,
                            componentId,
                          },
                        })
                        setError(
                          r?.error?.error ||
                            'Error occured, please refresh page and try again.'
                        )
                        setIsLoading(false)
                      } else {
                        trackEvent({
                          event: 'component_build',
                          user,
                          status: 'ok',
                          props: {
                            orgId,
                            appId,
                            componentId,
                          },
                        })
                        setIsLoading(false)
                        setIsKickedOff(true)

                        router.push(
                          `/${orgId}/apps/${appId}/components/${componentId}/builds/${r?.data?.id}`
                        )
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
                  Build component
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
        Build component
      </Button>
    </>
  )
}

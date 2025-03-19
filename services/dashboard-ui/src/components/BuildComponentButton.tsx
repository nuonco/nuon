'use client'

import React, { type FC, useEffect, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0/client'
import { Check, Hammer, WarningOctagon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import { createComponentBuild } from '@/components/app-actions'
import { trackEvent } from '@/utils'

export const BuildComponentButton: FC<{
  appId: string
  componentId: string
  componentName: string
  orgId: string
  onComplete?: () => void
}> = ({ appId, componentId, componentName, orgId, ...props }) => {
  const { user } = useUser()
  const [isConfirmOpen, setIsConfirmOpen] = useState(false)
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
      <Modal
        className="max-w-lg"
        isOpen={isConfirmOpen}
        heading={`Build ${componentName} component?`}
        onClose={() => {
          setIsConfirmOpen(false)
        }}
      >
        <div className="mb-6">
          {error ? (
            <span className="flex items-center gap-3  w-full p-2 border rounded-md border-red-400 bg-red-300/20 text-red-800 dark:border-red-600 dark:bg-red-600/5 dark:text-red-600 text-base font-medium">
              <WarningOctagon size="20" /> {error}
            </span>
          ) : null}
          <Text variant="reg-14" className="leading-relaxed">
            Are you sure you want to build {componentName}?
          </Text>
        </div>
        <div className="flex gap-3 justify-end">
          <Button
            onClick={() => {
              setIsConfirmOpen(false)
            }}
            className="text-base"
          >
            Cancel
          </Button>
          <Button
            className="text-sm flex items-center gap-1"
            onClick={() => {
              setIsLoading(true)
              createComponentBuild({ appId, componentId, orgId })
                .then(() => {
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
                  setIsConfirmOpen(false)
                  if (props.onComplete) props.onComplete()
                })
                .catch((err) => {
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
              <Hammer size="18" />
            )}{' '}
            Build component
          </Button>
        </div>
      </Modal>
      <Button
        className="text-sm flex items-center gap-1"
        onClick={() => {
          setIsConfirmOpen(true)
        }}
      >
        Build component
      </Button>
    </>
  )
}

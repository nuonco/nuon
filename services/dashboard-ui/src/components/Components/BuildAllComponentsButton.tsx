'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Check, Hammer } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { buildComponents } from '@/components/app-actions'
import { TComponent } from '@/types'

export const BuildAllComponentsButton: FC<{
  appId: string
  components: Array<TComponent>
  orgId: string
}> = ({ appId, components, orgId }) => {
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
      <Modal
        className="max-w-lg"
        isOpen={isOpen}
        heading={`Build all components?`}
        onClose={() => {
          setIsOpen(false)
        }}
      >
        <div className="mb-6">
          {error ? <Notice>{error}</Notice> : null}
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
                componentIds: components.map((c) => c.id),
                orgId,
              })
                .then(() => {
                  setIsLoading(false)
                  setIsKickedOff(true)
                  setIsOpen(false)
                })
                .catch((err) => {
                  console.error(err?.message)
                  setError(
                    'Unable to kick off component builds, please refresh page and try again.'
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
            Build all components
          </Button>
        </div>
      </Modal>
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

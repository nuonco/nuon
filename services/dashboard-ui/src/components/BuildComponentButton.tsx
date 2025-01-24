'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Check, Hammer } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'
import { createComponentBuild } from '@/components/app-actions'

export const BuildComponentButton: FC<{
  appId: string
  componentId: string
  componentName: string
  orgId: string
  onComplete?: () => void
}> = ({ appId, componentId, componentName, orgId, ...props }) => {
  const [isConfirmOpen, setIsConfirmOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)

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
              createComponentBuild({ appId, componentId, orgId }).then(() => {
                setIsLoading(false)
                setIsKickedOff(true)
                setIsConfirmOpen(false)
                if (props.onComplete) props.onComplete()
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

'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Check, PipeWrench } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { createComponentBuild } from '@/components/app-actions'

export const BuildComponentButton: FC<{
  appId: string
  componentId: string
  orgId: string
  onComplete?: () => void
}> = ({ appId, componentId, orgId, ...props }) => {
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
    <Button
      className="text-sm flex items-center gap-1"
      onClick={() => {
        setIsLoading(true)
        createComponentBuild({ appId, componentId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
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
        <PipeWrench size="18" />
      )}{' '}
      Build component
    </Button>
  )
}

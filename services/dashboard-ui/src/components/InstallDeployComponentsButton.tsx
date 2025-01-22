'use client'

import React, { type FC, useEffect, useState } from 'react'
import { CloudCheck, CloudArrowUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import {
  deployComponents,
  deployComponentBuild,
} from '@/components/install-actions'

export const InstallDeployComponentButton: FC<{
  installId: string
  orgId: string
  onComplete: () => void
}> = ({ installId, orgId, ...props }) => {
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
      className="text-base flex items-center gap-1"
      onClick={() => {
        setIsLoading(true)
        deployComponents({ installId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
          props.onComplete()
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
  )
}

export const InstallDeployLatestBuildButton: FC<{
  buildId: string
  installId: string
  orgId: string
}> = ({ buildId, installId, orgId }) => {
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
      className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3"
      onClick={() => {
        setIsLoading(true)
        deployComponentBuild({ buildId, installId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
        })
      }}
    >
      {isKickedOff ? (
        <CloudCheck size="18" />
      ) : isLoading ? (
        <SpinnerSVG />
      ) : (
        <CloudArrowUp size="18" />
      )}{' '}
      Deploy latest build
    </Button>
  )
}

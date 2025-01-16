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
}> = ({ installId, orgId }) => {
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
      className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-2"
      onClick={() => {
        setIsLoading(true)
        deployComponents({ installId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
        })
      }}
    >
      {isKickedOff ? (
        <CloudCheck size="20" />
      ) : isLoading ? (
        <SpinnerSVG />
      ) : (
        <CloudArrowUp size="20" />
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
      className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-2"
      onClick={() => {
        setIsLoading(true)
        deployComponentBuild({ buildId, installId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
        })
      }}
    >
      {isKickedOff ? (
        <CloudCheck size="20" />
      ) : isLoading ? (
        <SpinnerSVG />
      ) : (
        <CloudArrowUp size="20" />
      )}{' '}
      Deploy latest build
    </Button>
  )
}

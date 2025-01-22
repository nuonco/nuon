'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Check, ArrowURightUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { reprovisionInstall } from '@/components/install-actions'

export const InstallReprovisionButton: FC<{
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
      className="text-sm flex items-center gap-1"
      onClick={() => {
        setIsLoading(true)
        reprovisionInstall({ installId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
          props.onComplete()
        })
      }}
      variant="primary"
    >
      {isKickedOff ? (
        <Check size="18" />
      ) : isLoading ? (
        <SpinnerSVG />
      ) : (
        <ArrowURightUp size="18" />
      )}{' '}
      Reprovision
    </Button>
  )
}

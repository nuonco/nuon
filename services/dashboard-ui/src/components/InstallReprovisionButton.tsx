'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Check, ArrowURightUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { reprovisionInstall } from '@/components/install-actions'

export const InstallReprovisionButton: FC<{
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
      className="text-sm !font-medium !p-2 h-[32px] flex items-center gap-3 !rounded-none w-full"
      onClick={() => {
        setIsLoading(true)
        reprovisionInstall({ installId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
        })
      }}
      variant="ghost"
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

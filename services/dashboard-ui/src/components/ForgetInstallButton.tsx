'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Check, Trash } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { forgetInstall } from '@/components/install-actions'

export const ForgetInstallButton: FC<{
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
        forgetInstall({ installId, orgId }).then(() => {
          setIsLoading(false)
          setIsKickedOff(true)
          props.onComplete()
        }).catch(console.error)
      }}
      variant="danger"
    >
      {isKickedOff ? (
        <Check size="18" />
      ) : isLoading ? (
        <SpinnerSVG />
      ) : (
        <Trash size="18" />
      )}{' '}
      Forget install
    </Button>
  )
}

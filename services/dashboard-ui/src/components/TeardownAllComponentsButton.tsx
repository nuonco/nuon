'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Check, Axe } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { teardownAllComponents } from '@/components/install-actions'

export const TeardownAllComponentsButton: FC<{
  installId: string
  orgId: string
  onComplete?: () => void
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
        teardownAllComponents({ installId, orgId }).then(() => {
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
        <Axe size="18" />
      )}{' '}
      Teardown components
    </Button>
  )
}

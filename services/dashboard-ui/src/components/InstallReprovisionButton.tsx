'use client'

import React, { type FC, useEffect, useState } from 'react'
import { CloudCheck, CloudArrowUp } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { reprovisionInstall } from '@/components/install-actions'

export const InstallReprovisionButton: FC<{
  installId: string
  orgId: string
}> = ({ installId, orgId }) => {
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
        reprovisionInstall({ installId, orgId }).then(() => {
          setIsKickedOff(true)
        })
      }}
    >
      {isKickedOff ? <CloudCheck size="14" /> : <CloudArrowUp size="14" />}{' '}
      Reprovision
    </Button>
  )
}

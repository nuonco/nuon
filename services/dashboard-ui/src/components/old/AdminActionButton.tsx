'use client'

import React, { type FC, useState } from 'react'
import { CheckCircle, XCircle, Spinner } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { ToolTip } from '@/components/old/ToolTip'

interface IAdminButton {
  children: React.ReactNode
  action: () => Promise<Record<string, any>>
}

export const AdminBtn: FC<IAdminButton> = ({ children, action }) => {
  const [actionStatus, setActionStatus] = useState<
    'succeeded' | 'failed' | null
  >(null)
  const [isActing, setIsActing] = useState(false)

  return (
    <Button
      className="flex gap-2 items-center justify-center text-base w-full"
      onClick={() => {
        setIsActing(true)
        action()
          .then((res) => {
            if (res.status === 201) {
              setActionStatus('succeeded')
            }
            setIsActing(false)
          })
          .catch((err) => {
            setActionStatus('failed')
            console.error(err)
            setIsActing(false)
          })
      }}
      disabled={isActing}
    >
      {isActing ? (
        <>
          <Spinner className="animate-spin" /> Executing...
        </>
      ) : (
        <>
          <ActionIcon status={actionStatus} />
          {children}
        </>
      )}
    </Button>
  )
}

const ActionIcon: FC<{ status: 'succeeded' | 'failed' | null }> = ({
  status,
}) => {
  return status ? (
    <ToolTip tipContent={`Admin action ${status}`} isIconHidden>
      {status === 'failed' ? (
        <XCircle className="text-red-500" />
      ) : (
        <CheckCircle className="text-green-500" />
      )}
    </ToolTip>
  ) : null
}

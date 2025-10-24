'use client'

import React, { type FC, useEffect, useState } from 'react'
import { AdminTemporalLink } from '@/components/old/AdminTemporalLink'
import { ClickToCopy } from '@/components/old/ClickToCopy'
import { Text } from '@/components/old/Typography'
import { getInstallRunner } from '@/components/old/admin-actions'
import type { TRunner } from '@/types'

export const AdminInstallActions: FC<{
  children: any
  installId: string
}> = ({ children, installId }) => {
  const [runner, setInstallRunner] = useState<TRunner>()

  useEffect(() => {
    getInstallRunner(installId).then((r) => {
      setInstallRunner(r)
    })
  }, [])

  return (
    <div className="flex flex-col gap-4 pt-4">
      <Text variant="semi-18">Install admin controls</Text>
      <div className="flex gap-8">
        <Text variant="mono-14">
          Runner ID:{' '}
          {runner ? (
            <ClickToCopy>{runner?.id}</ClickToCopy>
          ) : (
            'Loading runner...'
          )}
        </Text>
        <AdminTemporalLink namespace="installs" id={installId} />
      </div>
      {children}
    </div>
  )
}

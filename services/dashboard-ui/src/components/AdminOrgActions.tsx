'use client'

import React, { type FC, useEffect, useState } from 'react'
import { ClickToCopy } from '@/components/ClickToCopy'
import { Text } from '@/components/Typography'
import { getOrgRunner } from '@/components/admin-actions'
import type { TRunner } from '@/types'


export const AdminOrgActions: FC<{
  children: any
  orgId: string
}> = ({ children, orgId }) => {
  const [runner, setInstallRunner] = useState<TRunner>()
  
  useEffect(() => {
    getOrgRunner(orgId).then((r) => {
      setInstallRunner(r)
    })
  }, [])
  
  return (
    <div className="flex flex-col gap-4 pt-4">
      <Text variant="semi-18">Org admin controls</Text>
      <Text variant="mono-14">
        Runner ID:{' '}
        {runner ? <ClickToCopy>{runner?.id}</ClickToCopy> : 'Loading runner...'}
      </Text>
      {children}
    </div>
  )
}

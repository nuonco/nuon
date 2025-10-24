'use client'

import React, { type FC, useEffect, useState } from 'react'
import { ClickToCopy } from '@/components/old/ClickToCopy'
import { AdminTemporalLink } from '@/components/old/AdminTemporalLink'
import { Text } from '@/components/old/Typography'
import { getOrgRunner } from '@/components/old/admin-actions'
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
      <div className="flex gap-8">
        <Text variant="mono-14">
          Runner ID:{' '}
          {runner ? (
            <ClickToCopy>{runner?.id}</ClickToCopy>
          ) : (
            'Loading runner...'
          )}
        </Text>
        <AdminTemporalLink namespace="orgs" id={orgId} />
      </div>
      {children}
    </div>
  )
}

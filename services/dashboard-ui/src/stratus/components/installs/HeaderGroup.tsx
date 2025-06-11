'use client'

import React, { type FC } from 'react'
import { Text, Time } from '@/stratus/components/common'
import { HeaderGroup } from '@/stratus/components/dashboard'
import { useInstall } from '@/stratus/context'

export const InstallHeaderGroup: FC = () => {
  const { install } = useInstall()

  return (
    <HeaderGroup>
      <Text variant="h3" weight="strong" level={1}>
        {install?.name}
      </Text>
      <Text family="mono" variant="subtext" theme="muted">
        {install?.id}
      </Text>
      <Text theme="highlighted" variant="subtext" weight="strong">
        Last updated <Time time={install?.updated_at} format="relative" />
      </Text>
    </HeaderGroup>
  )
}

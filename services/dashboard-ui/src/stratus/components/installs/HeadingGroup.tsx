'use client'

import React from 'react'
import { HeadingGroup, Text, Time } from '@/stratus/components/common'
import { useInstall } from '@/stratus/context'

export const InstallHeadingGroup = () => {
  const { install } = useInstall()

  return (
    <HeadingGroup>
      <Text variant="h3" weight="strong" level={1}>
        {install?.name}
      </Text>
      <Text family="mono" variant="subtext" theme="muted">
        {install?.id}
      </Text>
      <Text theme="highlighted" variant="subtext" weight="strong">
        Last updated <Time time={install?.updated_at} format="relative" />
      </Text>
    </HeadingGroup>
  )
}

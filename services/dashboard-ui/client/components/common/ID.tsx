import React from 'react'
import { useShowIds } from '@/hooks/use-show-ids'
import { ClickToCopy, IClickToCopy } from './ClickToCopy'
import { Text, IText } from './Text'

export interface IID extends IText {
  clickToCopyProps?: Omit<IClickToCopy, 'children'>
}

export function ID({ children, clickToCopyProps, loading, ...textProps }: IID) {
  const showIds = useShowIds()

  if (!showIds) return null

  return (
    <Text
      family="mono"
      variant="subtext"
      theme="neutral"
      loading={loading}
      {...textProps}
    >
      {loading ? null : (
        <ClickToCopy {...clickToCopyProps}>{children}</ClickToCopy>
      )}
    </Text>
  )
}

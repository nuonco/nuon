import React from 'react'
import { ClickToCopy, IClickToCopy } from './ClickToCopy'
import { Text, IText } from './Text'

export interface IID extends IText {
  clickToCopyProps?: Omit<IClickToCopy, 'children'>
}

export function ID({ children, clickToCopyProps, ...textProps }: IID) {
  return (
    <Text family="mono" variant="subtext" theme="neutral" {...textProps}>
      <ClickToCopy {...clickToCopyProps}>{children}</ClickToCopy>
    </Text>
  )
}

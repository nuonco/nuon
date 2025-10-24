import classNames from 'classnames'
import React, { type FC } from 'react'
import { Text } from '@/components/old/Typography'

export type TNuonVersions = {
  api: {
    git_ref: 'unknown'
    version: 'unknown'
  }
  ui: {
    version: 'unknown'
  }
}

interface INuonVersion
  extends React.HTMLAttributes<HTMLDivElement>,
    TNuonVersions {}

export const NuonVersions: FC<INuonVersion> = ({
  api,
  className,
  ui,
  ...props
}) => {
  return (
    <div
      {...props}
      className={classNames('flex gap-2 on-enter enter-left', {
        [`${className}`]: Boolean(className),
      })}
    >
      <Text className="!flex-nowrap" variant="reg-12">
        API: <b>{api.version}</b>
      </Text>
      <Text className="!flex-nowrap" variant="reg-12">
        UI: <b>{ui.version}</b>
      </Text>
    </div>
  )
}

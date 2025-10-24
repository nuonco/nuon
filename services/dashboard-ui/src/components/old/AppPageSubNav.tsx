import React, { type FC } from 'react'
import { SubNav } from '@/components/old/Nav'
import { WORKFLOWS } from '@/utils'

export interface IAppPageSubNav {
  appId: string
  orgId: string
}

export const AppPageSubNav: FC<IAppPageSubNav> = ({ appId, orgId }) => {
  return (
    <SubNav
      links={[
        { href: `/${orgId}/apps/${appId}`, text: 'Config' },
        { href: `/${orgId}/apps/${appId}/components`, text: 'Components' },
        { href: `/${orgId}/apps/${appId}/installs`, text: 'Installs' },
        WORKFLOWS
          ? { href: `/${orgId}/apps/${appId}/actions`, text: 'Actions' }
          : undefined,
      ]}
    />
  )
}

import React, { type FC } from 'react'
import { SubNav } from '@/components/Nav'
import { WORKFLOWS } from '@/utils'

export interface IInstallPageSubNav {
  installId: string
  orgId: string
}

export const InstallPageSubNav: FC<IInstallPageSubNav> = ({
  installId,
  orgId,
}) => {  
  return (
    <SubNav
      links={[
        { href: `/${orgId}/installs/${installId}`, text: 'Status' },
        {
          href: `/${orgId}/installs/${installId}/components`,
          text: 'Components',
        },
        WORKFLOWS
          ? {
              href: `/${orgId}/installs/${installId}/workflows`,
              text: 'Workflows',
            }
          : undefined,
      ]}
    />
  )
}

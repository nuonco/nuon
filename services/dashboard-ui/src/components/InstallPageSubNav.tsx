import React, { type FC } from 'react'
import { SubNav } from '@/components/Nav'
import { RUNNERS, WORKFLOWS } from '@/utils'

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
        { href: `/${orgId}/installs/${installId}`, text: 'Overview' },
        {
          href: `/${orgId}/installs/${installId}/history`,
          text: 'History',
        },
        {
          href: `/${orgId}/installs/${installId}/components`,
          text: 'Components',
        },
        RUNNERS
          ? {
              href: `/${orgId}/installs/${installId}/runner-group`,
              text: 'Runners',
            }
          : undefined,
        WORKFLOWS
          ? {
              href: `/${orgId}/installs/${installId}/actions`,
              text: 'Actions',
            }
          : undefined,
      ]}
    />
  )
}

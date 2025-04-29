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
        RUNNERS
          ? {
              href: `/${orgId}/installs/${installId}/runner`,
              text: 'Runner',
            }
          : undefined,
        {
          href: `/${orgId}/installs/${installId}/sandbox`,
          text: 'Sandbox',
        },
        {
          href: `/${orgId}/installs/${installId}/components`,
          text: 'Components',
        },
        WORKFLOWS
          ? {
              href: `/${orgId}/installs/${installId}/actions`,
              text: 'Actions',
            }
          : undefined,
        {
          href: `/${orgId}/installs/${installId}/history`,
          text: 'History',
        },
      ]}
    />
  )
}

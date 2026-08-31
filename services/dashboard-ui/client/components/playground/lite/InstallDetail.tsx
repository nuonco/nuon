import type { ReactNode } from 'react'
import { useParams } from 'react-router'
import { InstallSettingsButton } from './InstallSettingsButton'
import { Page } from './Page'
import { installTabs } from './nav'
import type { IDetailCrumb } from './types'

export interface IInstallDetail {
  crumbs?: IDetailCrumb[]
  actions?: string[]
  children: ReactNode
}

export const InstallDetail = ({
  crumbs = [],
  actions,
  children,
}: IInstallDetail) => {
  const { installId = '' } = useParams()
  const base = `/installs/${installId}`

  return (
    <Page
      tabs={installTabs(installId)}
      crumbs={[
        { label: 'Installs', path: '/installs' },
        { label: installId, path: base },
        ...crumbs.map((crumb) => ({
          label: crumb.label,
          path: crumb.slug ? `${base}/${crumb.slug}` : undefined,
        })),
      ]}
      actions={actions}
      actionsSlot={<InstallSettingsButton />}
    >
      {children}
    </Page>
  )
}

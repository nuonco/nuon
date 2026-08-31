import type { ReactNode } from 'react'
import { useParams } from 'react-router'
import { AppSettingsButton } from './AppSettingsButton'
import { BranchSwitcher } from './BranchSwitcher'
import { Page } from './Page'
import { appTabs } from './nav'
import type { IDetailCrumb } from './types'

export interface IAppDetail {
  crumbs?: IDetailCrumb[]
  actions?: string[]
  children: ReactNode
}

export const AppDetail = ({ crumbs = [], actions, children }: IAppDetail) => {
  const { appId = '', branchId = '' } = useParams()
  const base = `/apps/${appId}/branches/${branchId}`

  return (
    <Page
      tabs={appTabs(appId, branchId)}
      crumbs={[
        { label: 'Apps', path: '/apps' },
        { label: appId, path: base },
        ...crumbs.map((crumb) => ({
          label: crumb.label,
          path: crumb.slug ? `${base}/${crumb.slug}` : undefined,
        })),
      ]}
      actions={actions}
      actionsSlot={
        <>
          <BranchSwitcher />
          <AppSettingsButton />
        </>
      }
    >
      {children}
    </Page>
  )
}

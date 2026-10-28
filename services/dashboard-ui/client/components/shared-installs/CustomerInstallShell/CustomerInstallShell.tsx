import type { ReactNode } from 'react'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { SubNav } from '@/components/navigation/SubNav'
import { PageSidebarProvider } from '@/providers/page-sidebar-provider'
import { customerPortalNav } from '../customer-portal-nav'
import type { TInstallShareScope } from '../types'

export interface ICustomerInstallShell {
  installName: string
  installId: string
  basePath: string
  scopes: TInstallShareScope[]
  children: ReactNode
}

export const CustomerInstallShell = ({
  installName,
  installId,
  basePath,
  scopes,
  children,
}: ICustomerInstallShell) => (
  <PageSidebarProvider>
    <PageLayout>
      <PageHeader>
        <div className="flex flex-col gap-1.5">
          <Text variant="h3" weight="stronger" level={1}>
            {installName}
          </Text>
          <ID>{installId}</ID>
        </div>
      </PageHeader>
      <PageContent className="border-t" variant="row">
        <SubNav
          basePath={basePath}
          links={customerPortalNav(scopes)}
          storageKey="subnav:customer-install"
        />
        <div className="flex flex-col flex-1 min-w-0">{children}</div>
      </PageContent>
    </PageLayout>
  </PageSidebarProvider>
)

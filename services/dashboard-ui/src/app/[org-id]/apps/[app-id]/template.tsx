'use client'

import { usePathname } from 'next/navigation'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageHeadingGroup } from '@/components/layout/PageHeadingGroup'
import { SubNav } from '@/components/navigation/SubNav'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

// NOTE: old layout stuff
import { AppCreateInstallButton } from '@/components'

export default function Template({ children }: { children: React.ReactNode }) {
  const pathName = usePathname()
  const { org } = useOrg()
  const { app } = useApp()
  const isThirdLevel = pathName.split('/').length > 5

  return org?.features?.['stratus-layout'] ? (
    <PageLayout
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/${org?.id}/apps`,
            text: 'Apps',
          },
          {
            path: `/${org?.id}/apps/${app?.id}`,
            text: app?.name || 'App',
          },
        ],
      }}
    >
      {isThirdLevel ? (
        children
      ) : (
        <>
          <PageHeader>
            <PageHeadingGroup title={app.name} subtitle={<ID>{app.id}</ID>} />
            <div>
              <AppCreateInstallButton
                platform={app?.runner_config.app_runner_type}
              />
              {/* <Dropdown
               *   buttonText="Manage"
               *   id="app-manage"
               *   variant="primary"
               *   alignment="right"
               * >
               *   <Menu className="min-w-56">
               *     <Link href={`/${org.id}/apps/${app?.id}/configs`}>
               *       Config versions
               *       <Icon variant="GitDiff" />
               *     </Link>
               *     <Link href={`/${org.id}/apps/${app?.id}/workflows`}>
               *       Workflows
               *       <Icon variant="TreeStructure" />
               *     </Link>
               *   </Menu>
               * </Dropdown> */}
            </div>
          </PageHeader>
          <PageContent className="border-t" isScrollable variant="secondary">
            <SubNav
              basePath={`/${org?.id}/apps/${app?.id}`}
              links={[
                {
                  path: `/`,
                  iconVariant: 'HouseSimple',
                  text: 'Overview',
                },
                {
                  path: `/components`,
                  iconVariant: 'Cards',
                  text: 'Components',
                },
                {
                  path: `/actions`,
                  iconVariant: 'TerminalWindow',
                  text: 'Actions',
                },
                {
                  path: `/installs`,
                  iconVariant: 'Cube',
                  text: 'Installs',
                },
                /*{
              path: `/readme`,
              iconVariant: "BookOpen",
              text: "README",
            },*/
              ]}
            />
            {children}
          </PageContent>
        </>
      )}
    </PageLayout>
  ) : (
    children
  )
}

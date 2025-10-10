'use client'

import { usePathname } from 'next/navigation'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { SubNav } from '@/components/navigation/SubNav'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

// NOTE: old install components
import { InstallStatuses } from '@/components/InstallStatuses'
import { InstallManagementDropdown } from '@/components/Installs/ManagementDropdown'

export default function Template({ children }: { children: React.ReactNode }) {
  const pathName = usePathname()
  const { org } = useOrg()
  const { install } = useInstall()
  const isThirdLevel = pathName.split('/').length > 5

  return org?.features?.['stratus-layout'] ? (
    <PageLayout
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/${org?.id}/installs`,
            text: 'Installs',
          },
          {
            path: `/${org?.id}/installs/${install?.id}`,
            text: install?.name || 'Install',
          },
        ],
      }}
    >
      {isThirdLevel ? (
        children
      ) : (
        <>
          <PageHeader>
            <HeadingGroup>
              <Text variant="h3" weight="stronger" level={1}>
                {install.name}
              </Text>
              <ID>{install.id}</ID>
              <Text variant="subtext" theme="info">
                Last updated{' '}
                <Time
                  variant="subtext"
                  time={install?.updated_at}
                  format="relative"
                />
              </Text>
            </HeadingGroup>

            <div className="flex flex-wrap gap-4 md:gap-8">
              <LabeledValue label="App config">
                <Text variant="subtext">
                  <Link href={`/${org.id}/apps/${install.app_id}`}>
                    {install?.app?.name}
                  </Link>
                </Text>
              </LabeledValue>
              <InstallStatuses />
              <InstallManagementDropdown />
            </div>
          </PageHeader>
          <PageContent className="border-t" isScrollable variant="secondary">
            <SubNav
              basePath={`/${org?.id}/installs/${install?.id}`}
              links={[
                {
                  path: `/`,
                  iconVariant: 'HouseSimple',
                  text: 'Overview',
                },
                {
                  path: `/stacks`,
                  iconVariant: 'Stack',
                  text: 'Stacks',
                },
                {
                  path: `/runner`,
                  iconVariant: 'SneakerMove',
                  text: 'Runner',
                },
                {
                  path: '/sandbox', //`/sandbox/${install?.install_sandbox_runs?.at(0)?.id || ""}`,
                  iconVariant: 'ShippingContainer',
                  text: 'Sandbox',
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
                  path: `/workflows`,
                  iconVariant: 'TreeStructure',
                  text: 'Workflows',
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

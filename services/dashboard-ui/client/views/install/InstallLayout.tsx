import { Outlet, useParams, useMatch } from 'react-router'
import { LabelBadge } from '@/components/common/LabelBadge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Time } from '@/components/common/Time'
import { Text } from '@/components/common/Text'
import { DeprovisionBanner } from '@/components/installs/DeprovisionBanner'
import { DriftedSummary } from '@/components/installs/DriftedSummary'
import { InstallStatusesContainer } from '@/components/installs/InstallStatuses'
import { InstallManagementDropdown } from '@/components/installs/management/InstallManagementDropdown'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { SubNav } from '@/components/navigation/SubNav'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TNavItem } from '@/types'

import { PageSidebarProvider } from '@/providers/page-sidebar-provider'
import { InstallProvider } from '@/providers/install-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'

export const InstallLayout = () => {
  const params = useParams()

  return (
    <InstallProvider installId={params?.installId} shouldPoll>
      <PageSidebarProvider>
        <ToastProvider>
          <SurfacesProvider>
            <InstallTemplate />
          </SurfacesProvider>
        </ToastProvider>
      </PageSidebarProvider>
    </InstallProvider>
  )
}

const InstallTemplate = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const hasRunbooks = !!org?.features?.runbooks
  const hasNotebooks = !!org?.features?.notebooks

  const navLinks: TNavItem[] = [
    { type: 'section', label: 'Overview' },
    {
      path: `/`,
      iconVariant: 'HouseSimpleIcon' as const,
      text: 'Overview',
    },
    {
      path: `/workflows`,
      iconVariant: 'TreeStructureIcon' as const,
      text: 'Workflows',
    },
    { type: 'section', label: 'App' },
    {
      path: `/components`,
      iconVariant: 'CardsIcon' as const,
      text: 'Components',
    },
    {
      path: '/sandbox',
      iconVariant: 'ShippingContainerIcon' as const,
      text: 'Sandbox',
    },
    {
      path: `/roles`,
      iconVariant: 'FileLockIcon' as const,
      text: 'Roles',
    },
    {
      path: `/actions`,
      iconVariant: 'TerminalWindowIcon' as const,
      text: 'Actions',
    },
    ...(hasRunbooks
      ? [
          {
            path: `/runbooks`,
            iconVariant: 'BookIcon' as const,
            text: 'Runbooks',
          },
        ]
      : []),
    ...(hasNotebooks
      ? [
          {
            path: `/notebooks`,
            iconVariant: 'NotebookIcon' as const,
            text: 'Notebooks',
          },
        ]
      : []),
    { type: 'section', label: 'Customer' },
    {
      path: `/stacks`,
      iconVariant: 'StackIcon' as const,
      text: 'Stacks',
    },
    {
      path: `/policies`,
      iconVariant: 'ShieldCheckIcon' as const,
      text: 'Policy reports',
    },
    {
      path: `/inputs`,
      iconVariant: 'ListChecksIcon' as const,
      text: 'Current inputs',
    },
    {
      path: `/state`,
      iconVariant: 'CodeBlockIcon' as const,
      text: 'View state',
    },
    { type: 'section', label: 'Advanced' },
    {
      path: `/runner`,
      iconVariant: 'SneakerMoveIcon' as const,
      text: 'Install runner',
    },
  ]
  const isChildRoute = !!useMatch('/:orgId/installs/:installId/:section/:rest/*')

  if (!install) return null

  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  return (
    <PageLayout>    
      {isChildRoute ? (
        <PageContent className="border-t" variant="row">
          <SubNav
            basePath={`/${org?.id}/installs/${install?.id}`}
            links={navLinks}
          />
          <div className="flex flex-col flex-1 min-w-0">
            <Outlet />
          </div>
        </PageContent>
      ) : (
        <>
          <PageHeader>
                          <DeprovisionBanner />
            <div className="@container flex flex-col gap-6 w-full md:flex-row md:justify-between">
              <HeadingGroup className="gap-1.5">
                <div className="flex items-center gap-2 flex-wrap">
                  <Text variant="h3" weight="stronger" level={1}>
                    {install.name}
                  </Text>
                  {install.labels &&
                    Object.entries(install.labels).map(([key, value]) => (
                      <LabelBadge key={key} size="sm" variant="code" labelKey={key} labelValue={value} />
                    ))}
                </div>
                <ID>{install.id}</ID>
                <div className="flex items-center gap-3">
                  <Text variant="subtext" theme="info">
                    Last updated{' '}
                    <Time
                      variant="subtext"
                      time={install?.updated_at}
                      format="relative"
                    />
                  </Text>
                  <AdminDashboardLink
                    path={`/queues?owner_id=${install.id}`}
                    label="Admin panel"
                  />
                </div>
              </HeadingGroup>

              <div className="flex items-start flex-wrap gap-4 md:gap-8">
                {isManagedByConfig && (
                  <LabeledValue label="Managed By">
                    <Text variant="subtext">
                      <span className="flex items-center gap-1">
                        <Icon variant="FileCodeIcon" /> Install Config
                      </span>
                    </Text>
                  </LabeledValue>
                )}
                <LabeledValue label="App">
                  <Text variant="subtext">
                    <Link href={`/${org.id}/apps/${install.app_id}`}>
                      {install?.app?.name}
                    </Link>
                  </Text>
                </LabeledValue>
                <InstallStatusesContainer collapsible />
                <InstallManagementDropdown />
              </div>
            </div>
            {install?.drifted_objects?.length ? (
              <DriftedSummary
                className="mt-4"
                orgId={org.id}
                installId={install.id}
                driftedObjects={install.drifted_objects}
              />
            ) : null}
          </PageHeader>
          <PageContent className="border-t" variant="row">
            <SubNav
              basePath={`/${org?.id}/installs/${install?.id}`}
              links={navLinks}
            />
            <div className="flex flex-col flex-1 min-w-0">
              <Outlet />
            </div>
          </PageContent>
        </>
      )}
    </PageLayout>
  )
}

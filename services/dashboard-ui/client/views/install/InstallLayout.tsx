import { Outlet, useParams, useMatch } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Time } from '@/components/common/Time'
import { Text } from '@/components/common/Text'
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

const navLinks: TNavItem[] = [
  { type: 'section', label: 'Overview' },
  {
    path: `/`,
    iconVariant: 'HouseSimple' as const,
    text: 'Overview',
  },
  { type: 'section', label: 'App' },
  {
    path: `/stacks`,
    iconVariant: 'Stack' as const,
    text: 'Stacks',
  },
  {
    path: '/sandbox',
    iconVariant: 'ShippingContainer' as const,
    text: 'Sandbox',
  },
  {
    path: `/components`,
    iconVariant: 'Cards' as const,
    text: 'Components',
  },
  {
    path: `/roles`,
    iconVariant: 'FileLock' as const,
    text: 'Roles',
  },
  {
    path: `/policies`,
    iconVariant: 'ShieldCheck' as const,
    text: 'Policy reports',
  },
  { type: 'section', label: 'Day-2' },
  {
    path: `/actions`,
    iconVariant: 'TerminalWindow' as const,
    text: 'Actions',
  },
  {
    path: `/workflows`,
    iconVariant: 'TreeStructure' as const,
    text: 'Workflows',
  },
  {
    path: `/runner`,
    iconVariant: 'SneakerMove' as const,
    text: 'Install runner',
  },
  {
    path: `/inputs`,
    iconVariant: 'ListChecks' as const,
    text: 'Current inputs',
  },
  {
    path: `/state`,
    iconVariant: 'CodeBlock' as const,
    text: 'View state',
  },
]

const InstallTemplate = () => {
  const { org } = useOrg()
  const { install } = useInstall()
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
            <div className="@container flex flex-col gap-4 w-full md:flex-row md:justify-between">
              <HeadingGroup>
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
                    label="View in admin panel"
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
              <div className="self-center flex flex-col gap-2">
                <Text theme="warn">
                  <span className="flex items-center gap-2">
                    <Icon variant="WarningIcon" weight="bold" />
                    Drift detected
                  </span>
                </Text>
                <div className="self-center flex items-center gap-6">
                  {install?.drifted_objects?.map((drift) => (
                    <Badge size="sm" theme="warn" key={drift?.target_id}>
                      Drifted:{' '}
                      <Link
                        href={`/${org.id}/installs/${install?.id}/workflows/${drift?.install_workflow_id}`}
                        className="!leading-none"
                      >
                        {drift?.target_type === 'install_deploy'
                          ? drift?.component_name
                          : 'Sandbox'}
                      </Link>
                    </Badge>
                  ))}
                </div>
              </div>
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

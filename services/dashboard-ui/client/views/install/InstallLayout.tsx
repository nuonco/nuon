import {
  Outlet,
  useParams,
  useMatch,
  useSearchParams,
  useLocation,
} from 'react-router'
import { LabelBadge } from '@/components/common/LabelBadge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Time } from '@/components/common/Time'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { PageSection } from '@/components/layout/PageSection'
import { DeprovisionBanner } from '@/components/installs/DeprovisionBanner'
import { DriftedSummary } from '@/components/installs/DriftedSummary'
import { InstallStatusesContainer } from '@/components/installs/InstallStatuses'
import {
  InstallSettingsPanel,
  useOpenInstallSettings,
  INSTALL_SETTINGS_PANEL_KEY,
} from '@/components/installs/InstallSettingsPanel'
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
import { InstallAppConfigProvider } from '@/providers/install-app-config-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'

export const InstallLayout = () => {
  const params = useParams()

  return (
    <InstallProvider installId={params?.installId} shouldPoll>
      <InstallAppConfigProvider>
        <PageSidebarProvider>
          <ToastProvider>
            <SurfacesProvider>
              <InstallTemplate />
            </SurfacesProvider>
          </ToastProvider>
        </PageSidebarProvider>
      </InstallAppConfigProvider>
    </InstallProvider>
  )
}

const InstallContentError = () => (
  <PageSection>
    <EmptyState
      variant="404"
      emptyTitle="Something went wrong"
      emptyMessage="We couldn't load this page. Some of its data may be unavailable right now."
      action={
        <Button variant="secondary" onClick={() => window.location.reload()}>
          Try again
        </Button>
      }
    />
  </PageSection>
)

const InstallTemplate = () => {
  const { org } = useOrg()
  const { install, labelColors } = useInstall()
  const { pathname } = useLocation()
  const hasNotebooks = !!org?.features?.notebooks
  const openSettings = useOpenInstallSettings()
  const [searchParams] = useSearchParams()
  const isSettingsOpen =
    searchParams.get('panel') === INSTALL_SETTINGS_PANEL_KEY

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
    {
      path: `/versions`,
      iconVariant: 'ClockCounterClockwiseIcon' as const,
      text: 'App branch runs',
    },
    {
      type: 'action',
      key: 'settings',
      iconVariant: 'GearIcon' as const,
      text: 'Settings',
      onClick: openSettings,
      isActive: isSettingsOpen,
    },
    { type: 'section', label: 'App' },
    ...(org?.features?.['component-health']
      ? [
          {
            path: `/resources`,
            iconVariant: 'PulseIcon' as const,
            text: 'Resources',
          },
        ]
      : []),
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
    {
      path: `/runbooks`,
      iconVariant: 'BookIcon' as const,
      text: 'Runbooks',
    },
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
      path: `/configs`,
      iconVariant: 'GearIcon' as const,
      text: 'Configs',
    },
    {
      path: `/runner`,
      iconVariant: 'SneakerMoveIcon' as const,
      text: 'Install runner',
    },
  ]
  const isChildRoute = !!useMatch(
    '/:orgId/installs/:installId/:section/:rest/*'
  )

  if (!install) return null

  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  return (
    <>
      <InstallSettingsPanel />
      <PageLayout>
        {isChildRoute ? (
          <PageContent className="border-t" variant="row">
            <SubNav
              basePath={`/${org?.id}/installs/${install?.id}`}
              links={navLinks}
            />
            <div className="flex flex-col flex-1 min-w-0">
              <ErrorBoundary key={pathname} fallback={<InstallContentError />}>
                <Outlet />
              </ErrorBoundary>
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
                        <LabelBadge key={key} size="sm" labelKey={key} labelValue={value} customColor={labelColors?.[key]} />
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
                  {install?.app_branch && (
                    <LabeledValue label="Branch">
                      <Text variant="subtext">
                        <Link href={`/${org?.id}/apps/${install?.app_id}/branches/${install?.app_branch?.id}`}>
                          <span className="flex items-center gap-1">
                            <Icon variant="GitBranchIcon" size={14} />
                            {install.app_branch.name}
                          </span>
                        </Link>
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
                <ErrorBoundary key={pathname} fallback={<InstallContentError />}>
                  <Outlet />
                </ErrorBoundary>
              </div>
            </PageContent>
          </>
        )}
      </PageLayout>
    </>
  )
}

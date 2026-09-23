import { useEffect, useState, type ReactNode } from 'react'
import { useLocation, useNavigate, useSearchParams } from 'react-router'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { SubNav } from '@/components/navigation/SubNav'
import { TabNav } from '@/components/navigation/TabNav'
import type { TNavItem } from '@/types'
import {
  ActionDetail,
  ComponentDetail,
  ConfigFileTab,
  ConfigurationSummaryRow,
  ConfigurationVersionFeed,
  DEFAULT_DEPLOYMENT_FILTER,
  DeploymentsTab,
  getInstallStatusEntries,
  HealthChecksTab,
  ImageDetail,
  InputsTab,
  InstallBranchTrackingCard,
  OverviewTab,
  OverridesTab,
  PoliciesTab,
  RunbookDetail,
  RunnerTab,
  SandboxTab,
  StackTab,
  type TDeploymentFilter,
  type TTopTab,
} from '../InstallDetailPlayground/InstallDetailPlayground'
import type { TPlaygroundInstall } from '../InstallDetailPlayground/types'

const SECTION_PARAM = 'section'

type TTabNavigate = (
  tab: TTopTab,
  opts?: {
    resourcesTab?: string
    configurationTab?: string
  }
) => void

type TPageContext = {
  install: TPlaygroundInstall
  deploymentFilter: TDeploymentFilter
  onDeploymentFilterChange: (filter: TDeploymentFilter) => void
  sectionId?: string
  onNavigate: TTabNavigate
}

type TPlaygroundPage = {
  tab: TTopTab
  path: string
  text: string
  iconVariant: TIconVariant
  title: string
  description: string
  render: (ctx: TPageContext) => ReactNode
}

type TSubpage = {
  id: string
  label: string
  content: ReactNode
}

const EmptyCollection = ({ children }: { children: ReactNode }) => (
  <div className="p-4">
    <Text variant="subtext" theme="neutral">
      {children}
    </Text>
  </div>
)

const InstallSubpageTabs = ({
  activeId,
  basePath,
  pages,
}: {
  activeId?: string
  basePath: string
  pages: TSubpage[]
}) => {
  const activeIndex = Math.max(
    pages.findIndex((page) => page.id === activeId),
    0
  )

  return (
    <div className="flex flex-col min-w-0">
      <div className="px-4 md:px-6">
        <TabNav
          activeIndex={activeIndex}
          basePath={basePath}
          tabs={pages.map((page) => ({
            path: `?${SECTION_PARAM}=${page.id}`,
            text: page.label,
          }))}
        />
      </div>
      {pages[activeIndex]?.content}
    </div>
  )
}

const ResourcesPages = ({
  install,
  sectionId,
}: {
  install: TPlaygroundInstall
  sectionId?: string
}) => (
  <InstallSubpageTabs
    activeId={sectionId}
    basePath={`/${install.orgId}/installs/${install.id}/resources`}
    pages={[
      {
        id: 'stack',
        label: 'Stack',
        content: <StackTab versions={install.resources.stackVersions} />,
      },
      {
        id: 'sandbox',
        label: 'Sandbox',
        content: <SandboxTab sandbox={install.resources.sandbox} />,
      },
      {
        id: 'components',
        label: 'Components',
        content: install.resources.components.length ? (
          <div className="divide-y">
            {install.resources.components.map((component) => (
              <ComponentDetail key={component.id} component={component} />
            ))}
          </div>
        ) : (
          <EmptyCollection>No components configured.</EmptyCollection>
        ),
      },
      {
        id: 'images',
        label: 'Images',
        content: install.resources.images.length ? (
          <div className="divide-y">
            {install.resources.images.map((image) => (
              <ImageDetail key={image.id} image={image} />
            ))}
          </div>
        ) : (
          <EmptyCollection>No images configured.</EmptyCollection>
        ),
      },
    ]}
  />
)

const OperationsPages = ({
  install,
  sectionId,
}: {
  install: TPlaygroundInstall
  sectionId?: string
}) => (
  <InstallSubpageTabs
    activeId={sectionId}
    basePath={`/${install.orgId}/installs/${install.id}/operations`}
    pages={[
      {
        id: 'actions',
        label: 'Actions',
        content: install.operations.actions.length ? (
          <div className="divide-y">
            {install.operations.actions.map((action) => (
              <ActionDetail key={action.id} action={action} />
            ))}
          </div>
        ) : (
          <EmptyCollection>No actions configured.</EmptyCollection>
        ),
      },
      {
        id: 'runbooks',
        label: 'Runbooks',
        content: install.operations.runbooks.length ? (
          <div className="divide-y">
            {install.operations.runbooks.map((runbook) => (
              <RunbookDetail key={runbook.id} runbook={runbook} />
            ))}
          </div>
        ) : (
          <EmptyCollection>No runbooks configured.</EmptyCollection>
        ),
      },
      {
        id: 'policies',
        label: 'Policies',
        content: <PoliciesTab policies={install.operations.policies} />,
      },
      {
        id: 'runner',
        label: 'Runner',
        content: <RunnerTab runner={install.operations.runner} />,
      },
    ]}
  />
)

const ConfigurationPages = ({
  install,
  sectionId,
}: {
  install: TPlaygroundInstall
  sectionId?: string
}) => (
  <InstallSubpageTabs
    activeId={sectionId}
    basePath={`/${install.orgId}/installs/${install.id}/configuration`}
    pages={[
      {
        id: 'appBranch',
        label: 'App branch',
        content: (
          <div className="flex flex-col gap-4 p-4">
            <InstallBranchTrackingCard install={install} />
            <ConfigurationVersionFeed
              versions={install.configuration.appBranchVersions}
              idPrefix="layout-app-branch"
            />
          </div>
        ),
      },
      {
        id: 'inputs',
        label: 'Inputs',
        content: (
          <InputsTab
            inputs={install.configuration.inputs}
            versions={install.configuration.inputVersions}
          />
        ),
      },
      {
        id: 'configFile',
        label: 'Config file',
        content: (
          <ConfigFileTab
            configFile={install.configuration.configFile}
            versions={install.configuration.configFileVersions}
          />
        ),
      },
      {
        id: 'overrides',
        label: 'Overrides',
        content: <OverridesTab overrides={install.configuration.overrides} />,
      },
    ]}
  />
)

const PAGES: TPlaygroundPage[] = [
  {
    tab: 'overview',
    path: '/',
    text: 'Overview',
    iconVariant: 'HouseSimpleIcon',
    title: 'Install overview',
    description: 'Applied configuration, drift, and the install readme.',
    render: ({ install }) => <OverviewTab install={install} />,
  },
  {
    tab: 'deployments',
    path: '/deployments',
    text: 'Deployments',
    iconVariant: 'ArrowsClockwiseIcon',
    title: 'Deployments',
    description: 'Every change applied to this install, newest first.',
    render: ({ install, deploymentFilter, onDeploymentFilterChange }) => (
      <DeploymentsTab
        install={install}
        filter={deploymentFilter}
        onFilterChange={onDeploymentFilterChange}
      />
    ),
  },
  {
    tab: 'resources',
    path: '/resources',
    text: 'Resources',
    iconVariant: 'CardsIcon',
    title: 'Resources',
    description: 'Stack, sandbox, components, and images for this install.',
    render: ({ install, sectionId }) => (
      <ResourcesPages install={install} sectionId={sectionId} />
    ),
  },
  {
    tab: 'health',
    path: '/health',
    text: 'Health checks',
    iconVariant: 'PulseIcon',
    title: 'Health checks',
    description: 'Install and component health over the last 30 days.',
    render: ({ install, onNavigate }) => (
      <HealthChecksTab
        install={install}
        onSelectComponent={() =>
          onNavigate('resources', { resourcesTab: 'components' })
        }
      />
    ),
  },
  {
    tab: 'operations',
    path: '/operations',
    text: 'Operations',
    iconVariant: 'TerminalWindowIcon',
    title: 'Operations',
    description: 'Actions, runbooks, policy reports, and the install runner.',
    render: ({ install, sectionId }) => (
      <OperationsPages install={install} sectionId={sectionId} />
    ),
  },
  {
    tab: 'configuration',
    path: '/configuration',
    text: 'Configuration',
    iconVariant: 'FadersIcon',
    title: 'Configuration',
    description: 'App branch, inputs, config file, and overrides.',
    render: ({ install, sectionId }) => (
      <ConfigurationPages install={install} sectionId={sectionId} />
    ),
  },
]

const navLinks: TNavItem[] = PAGES.map(({ path, text, iconVariant }) => ({
  path,
  text,
  iconVariant,
}))

const TAB_PATHS = Object.fromEntries(
  PAGES.map(({ tab, path }) => [tab, path])
) as Record<TTopTab, string>

const driftItems = (install: TPlaygroundInstall): TContextTooltipItem[] =>
  install.driftedObjects.length
    ? install.driftedObjects.map((drift) => ({
        id: drift.id,
        title:
          drift.targetType === 'sandbox'
            ? 'Sandbox'
            : (drift.componentName ?? 'Component'),
        subtitle: 'Drift detected',
        leftContent: (
          <Status
            status="warn"
            isWithoutText
            variant="timeline"
            iconSize={16}
          />
        ),
      }))
    : [
        {
          id: 'no-drift',
          title: 'No drift',
          subtitle: 'This install has detected no drift',
          leftContent: (
            <Status
              status="active"
              isWithoutText
              variant="timeline"
              iconSize={16}
            />
          ),
        },
      ]

const InstallStatusSummaryCard = ({
  install,
  onNavigate,
}: {
  install: TPlaygroundInstall
  onNavigate: TTabNavigate
}) => {
  const hasDrift = install.driftedObjects.length > 0

  return (
    <Card className="!p-4" aria-label="Install status summary">
      <div className="flex flex-wrap gap-x-8 gap-y-3">
        {getInstallStatusEntries(install).map((entry) => {
          const value = (
            <Button
              variant="ghost"
              size="sm"
              className="!px-0"
              onClick={() => onNavigate(entry.tab, entry.opts)}
              aria-label={`${entry.label}: ${entry.statusLabel}. Navigate to ${entry.label.toLowerCase()}.`}
            >
              <span className="flex items-center gap-1.5">
                <Status status={entry.status} variant="badge">
                  {entry.statusLabel}
                </Status>
                {entry.id === 'resources' && hasDrift ? (
                  <Text as="span" theme="warn">
                    <Icon variant="FileDashedIcon" size={14} />
                  </Text>
                ) : null}
              </span>
            </Button>
          )

          return (
            <LabeledValue key={entry.id} label={entry.label}>
              {entry.id === 'resources' ? (
                <ContextTooltip
                  title="Drift detection"
                  items={driftItems(install)}
                  showCount={hasDrift}
                  width="w-56"
                >
                  {value}
                </ContextTooltip>
              ) : (
                value
              )}
            </LabeledValue>
          )
        })}
      </div>
    </Card>
  )
}

const PlaygroundPageShell = ({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) => (
  <PageSection flush>
    <div className="p-4 pb-0 md:px-6 md:pt-6">
      <SectionHeader title={title} description={description} />
    </div>
    {children}
  </PageSection>
)

const InstallPlaygroundHeader = ({
  install,
  onNavigate,
}: {
  install: TPlaygroundInstall
  onNavigate: TTabNavigate
}) => (
  <PageHeader>
    <div className="flex flex-col gap-4 w-full">
      <HeadingGroup className="gap-1.5">
        <Text variant="h3" weight="stronger" level={1}>
          {install.name}
        </Text>
        <div className="flex items-center gap-3 flex-wrap">
          <ID>{install.id}</ID>
          <Text variant="subtext" theme="info">
            Last updated{' '}
            <Time
              variant="subtext"
              time={install.updatedAt}
              format="relative"
            />
          </Text>
          <AdminDashboardLink
            path={`/queues?owner_id=${install.id}`}
            label="Admin panel"
          />
        </div>
      </HeadingGroup>
      <div className="flex items-center gap-2 flex-wrap">
        {Object.entries(install.labels).map(([key, value]) => (
          <LabelBadge key={key} size="sm" labelKey={key} labelValue={value} />
        ))}
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 w-full">
        <ConfigurationSummaryRow install={install} onNavigate={onNavigate} />
        <InstallStatusSummaryCard install={install} onNavigate={onNavigate} />
      </div>
    </div>
  </PageHeader>
)

export interface IInstallDetailLayoutPlayground {
  install: TPlaygroundInstall
}

export const InstallDetailLayoutPlayground = ({
  install,
}: IInstallDetailLayoutPlayground) => {
  const basePath = `/${install.orgId}/installs/${install.id}`
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [deploymentFilter, setDeploymentFilter] = useState<TDeploymentFilter>(
    DEFAULT_DEPLOYMENT_FILTER
  )

  // Ladle mounts stories at `/`, so nothing in the SubNav reads as active until
  // the router sits under the install's base path.
  useEffect(() => {
    if (!pathname.startsWith(basePath)) navigate(basePath, { replace: true })
  }, [basePath, navigate, pathname])

  const relativePath = pathname.startsWith(basePath)
    ? pathname.slice(basePath.length).replace(/\/$/, '') || '/'
    : '/'
  const page = PAGES.find(({ path }) => path === relativePath) ?? PAGES[0]
  const sectionId = searchParams.get(SECTION_PARAM) ?? undefined

  const onNavigate: TTabNavigate = (tab, opts) => {
    const params = new URLSearchParams()
    const section = opts?.resourcesTab ?? opts?.configurationTab
    if (section) params.set(SECTION_PARAM, section)
    const query = params.toString()
    navigate(`${basePath}${TAB_PATHS[tab]}${query ? `?${query}` : ''}`)
  }

  return (
    <PageLayout>
      <InstallPlaygroundHeader install={install} onNavigate={onNavigate} />
      <PageContent className="border-t" variant="row">
        <SubNav
          basePath={basePath}
          links={navLinks}
          storageKey="subnav:install-layout-playground"
        />
        <div className="flex flex-col flex-1 min-w-0">
          <PlaygroundPageShell
            key={`${page.path}-${sectionId}`}
            title={page.title}
            description={page.description}
          >
            {page.render({
              install,
              deploymentFilter,
              onDeploymentFilterChange: setDeploymentFilter,
              sectionId,
              onNavigate,
            })}
          </PlaygroundPageShell>
        </div>
      </PageContent>
    </PageLayout>
  )
}

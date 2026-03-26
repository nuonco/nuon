import { Outlet, useParams, useLocation } from 'react-router'
import { TemporalLink } from '@/components/admin/TemporalLink'
import { Badge } from '@/components/common/Badge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Time } from '@/components/common/Time'
import { Text } from '@/components/common/Text'
import { InstallStatusesContainer } from '@/components/installs/InstallStatuses'
import { InstallManagementDropdown } from '@/components/installs/management/InstallManagementDropdown'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { SubNav } from '@/components/navigation/SubNav'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

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
  const location = useLocation()
  const { org } = useOrg()
  const { install } = useInstall()
  const isThirdLevel = location.pathname.split('/').length > 5

  if (!install) return null

  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  return (
    <PageLayout>
      {isThirdLevel ? (
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
                text: 'Install runner',
              },
              {
                path: '/sandbox',
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
                path: `/roles`,
                iconVariant: 'FileLock',
                text: 'Roles',
              },
              {
                path: `/policies`,
                iconVariant: 'ShieldCheck',
                text: 'Policies',
              },
              {
                path: `/workflows`,
                iconVariant: 'TreeStructure',
                text: 'Workflows',
              },
              {
                path: `/readme`,
                iconVariant: 'BookOpen',
                text: 'README',
              },
            ]}
          />
          <div className="flex flex-col flex-1 min-w-0">
            <Outlet />
          </div>
        </PageContent>
      ) : (
        <>
          <PageHeader>
            <div className="flex justify-between w-full">
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

              <div className="flex items-start flex-wrap gap-4 md:gap-8">
                <TemporalLink namespace="installs" eventLoopId={install?.id} />
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
                <InstallStatusesContainer />
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
                  text: 'Install runner',
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
                  path: `/roles`,
                  iconVariant: 'FileLock',
                  text: 'Roles',
                },
                {
                  path: `/policies`,
                  iconVariant: 'ShieldCheck',
                  text: 'Policies',
                },
                {
                  path: `/workflows`,
                  iconVariant: 'TreeStructure',
                  text: 'Workflows',
                },
                {
                  path: `/readme`,
                  iconVariant: 'BookOpen',
                  text: 'README',
                },
              ]}
            />
            <Outlet />
          </PageContent>
        </>
      )}
    </PageLayout>
  )
}

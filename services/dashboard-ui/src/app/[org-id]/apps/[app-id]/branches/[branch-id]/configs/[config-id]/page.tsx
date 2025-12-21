import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Time } from '@/components/common/Time'
import { Card } from '@/components/common/Card'
import { Badge } from '@/components/common/Badge'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { BackLink } from '@/components/common/BackLink'
import { getApp, getAppBranch, getBranchConfig, getOrg } from '@/lib'
import type { TPageProps } from '@/types'

type TConfigPageProps = TPageProps<
  'org-id' | 'app-id' | 'branch-id' | 'config-id'
>

export async function generateMetadata({
  params,
}: TConfigPageProps): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['branch-id']: branchId,
    ['config-id']: configId,
  } = await params
  const { data: config } = await getBranchConfig({
    appId,
    branchId,
    configId,
    orgId,
  })

  return {
    title: `Config v${config?.config_number || ''} | Branch | Nuon`,
  }
}

export default async function BranchConfigDetailPage({
  params,
}: TConfigPageProps) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['branch-id']: branchId,
    ['config-id']: configId,
  } = await params

  const [
    { data: config, error: configError },
    { data: branch },
    { data: app },
    { data: org },
  ] = await Promise.all([
    getBranchConfig({ appId, branchId, configId, orgId }),
    getAppBranch({ appId, branchId, orgId }),
    getApp({ appId, orgId }),
    getOrg({ orgId }),
  ])

  if (configError || !config) {
    notFound()
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name || '',
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name || '',
          },
          {
            path: `/${orgId}/apps/${appId}/branches`,
            text: 'Branches',
          },
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}`,
            text: branch?.name || '',
          },
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}/configs`,
            text: 'Configurations',
          },
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}/configs/${configId}`,
            text: `v${config?.config_number || ''}`,
          },
        ]}
      />

      <BackLink className="mb-4">
        Back to configurations
      </BackLink>

      {/* Page Header */}
      <div className="flex items-start justify-between mb-6">
        <HeadingGroup>
          <div className="flex items-center gap-3">
            <Text variant="h3" weight="stronger">
              Configuration
            </Text>
            <Badge theme="info">v{config.config_number}</Badge>
          </div>
          <ID>{config.id}</ID>
          <Text variant="subtext" theme="info">
            Created <Time time={config?.created_at} format="relative" />
          </Text>
        </HeadingGroup>
      </div>

      {/* Configuration Details Card */}
      <Card className="mb-6">
        <div className="p-6">
          <Text variant="h4" weight="strong" className="mb-4">
            Configuration Details
          </Text>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <LabeledValue label="Version">
              <Text variant="base">v{config.config_number}</Text>
            </LabeledValue>

            <LabeledValue label="Created">
              <Time time={config.created_at} format="relative" />
            </LabeledValue>

            <LabeledValue label="Install Groups">
              <Text variant="base">
                {config.install_groups?.length || 0} group
                {config.install_groups?.length !== 1 ? 's' : ''}
              </Text>
            </LabeledValue>
          </div>

          {config.connected_github_vcs_config && (
            <div className="border-t pt-4">
              <Text variant="base" weight="strong" className="mb-3">
                VCS Configuration
              </Text>
              <div className="space-y-2">
                <div className="flex items-start gap-2">
                  <Text variant="subtext" theme="neutral" className="w-32">
                    Repository:
                  </Text>
                  <Text variant="base">
                    {config.connected_github_vcs_config.repo}
                  </Text>
                </div>
                <div className="flex items-start gap-2">
                  <Text variant="subtext" theme="neutral" className="w-32">
                    Branch:
                  </Text>
                  <code className="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
                    {config.connected_github_vcs_config.branch}
                  </code>
                </div>
                <div className="flex items-start gap-2">
                  <Text variant="subtext" theme="neutral" className="w-32">
                    Directory:
                  </Text>
                  <Text variant="base">
                    {config.connected_github_vcs_config.directory || '.'}
                  </Text>
                </div>
                {config.connected_github_vcs_config.path_filter && (
                  <div className="flex items-start gap-2">
                    <Text variant="subtext" theme="neutral" className="w-32">
                      Path Filter:
                    </Text>
                    <code className="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
                      {config.connected_github_vcs_config.path_filter}
                    </code>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Install Groups Section */}
      {config.install_groups && config.install_groups.length > 0 && (
        <Card>
          <div className="p-6">
            <Text variant="h4" weight="strong" className="mb-4">
              Install Groups
            </Text>
            <div className="space-y-3">
              {config.install_groups.map((group, idx) => (
                <div
                  key={group.id || idx}
                  className="p-4 bg-gray-50 dark:bg-gray-900 rounded-md border border-gray-200 dark:border-gray-700"
                >
                  <div className="flex items-center justify-between mb-3">
                    <Text variant="base" weight="strong">
                      {idx + 1}. {group.name}
                    </Text>
                    <div className="flex items-center gap-2">
                      {group.requires_approval && (
                        <Badge theme="warning" size="sm">
                          Requires Approval
                        </Badge>
                      )}
                      {group.rollback_on_failure && (
                        <Badge theme="info" size="sm">
                          Rollback on Failure
                        </Badge>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <LabeledValue label="Installs">
                      <Text variant="body">
                        {group.install_ids?.length || 0} install
                        {group.install_ids?.length !== 1 ? 's' : ''}
                      </Text>
                    </LabeledValue>
                    <LabeledValue label="Max Parallel">
                      <Text variant="body">{group.max_parallel || 1}</Text>
                    </LabeledValue>
                  </div>
                  {group.install_ids && group.install_ids.length > 0 && (
                    <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                      <Text variant="subtext" theme="neutral" className="mb-2">
                        Install IDs:
                      </Text>
                      <div className="flex flex-wrap gap-2">
                        {group.install_ids.map((installId) => (
                          <code
                            key={installId}
                            className="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded"
                          >
                            {installId.slice(0, 8)}...
                          </code>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </Card>
      )}
    </PageSection>
  )
}
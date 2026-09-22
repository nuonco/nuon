import type { ReactNode } from 'react'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { Card } from '@/components/common/Card'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageHeader } from '@/components/layout/PageHeader'
import type { TCloudPlatform, TInstall } from '@/types'

export interface INewInstallHeader {
  branchAction?: ReactNode
  install: TInstall
  labelColors?: Record<string, string>
  orgId?: string
  settingsAction?: ReactNode
  statuses?: ReactNode
}

export const NewInstallHeader = ({
  branchAction,
  install,
  labelColors,
  orgId,
  settingsAction,
  statuses,
}: INewInstallHeader) => {
  const installPath = `/${orgId}/installs/${install.id}`
  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'
  const labels = Object.entries(install.labels ?? {})
  const platform = (install?.cloud_platform as TCloudPlatform) || 'unknown'
  const region = install.aws_account?.region ?? install.gcp_account?.region
  const location = install.azure_account?.location

  return (
    <PageHeader>
      <div className="flex flex-col gap-4 w-full">
        <div className="flex items-start justify-between gap-4 w-full">
          <HeadingGroup className="gap-1.5">
            <div className="flex items-center gap-3 flex-wrap">
              <Text variant="h3" weight="stronger" level={1}>
                {install.name}
              </Text>
              {region || location ? (
                <span className="flex items-center gap-1.5">
                  <CloudPlatform
                    platform={platform}
                    colorVariant="color"
                    displayVariant="icon-only"
                    iconSize="16"
                  />
                  <CloudRegion
                    variant="subtext"
                    theme="neutral"
                    platform={platform}
                    region={region}
                    location={location}
                  />
                </span>
              ) : null}
            </div>
            <div className="flex items-center gap-3 flex-wrap">
              <ID>{install.id}</ID>
              <Text variant="subtext" theme="neutral">
                Created{' '}
                <Time
                  variant="subtext"
                  time={install?.created_at}
                  format="relative"
                />
              </Text>
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
          {settingsAction}
        </div>

        {labels.length ? (
          <div className="flex items-center gap-2 flex-wrap">
            {labels.map(([key, value]) => (
              <LabelBadge
                key={key}
                size="sm"
                labelKey={key}
                labelValue={value}
                customColor={labelColors?.[key]}
              />
            ))}
          </div>
        ) : null}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 w-full">
          <Card className="!p-4 !shadow-none w-full">
            <div className="flex items-start gap-x-8 gap-y-4 flex-wrap">
              <LabeledValue label="App">
                <Link
                  href={`/${orgId}/apps/${install.app_id}`}
                  textVariant="subtext"
                >
                  {install.app?.name}
                </Link>
              </LabeledValue>

              <LabeledValue label="App branch">
                <span className="flex items-center gap-1">
                  {install.app_branch ? (
                    <Link
                      href={`/${orgId}/apps/${install.app_id}/branches/${install.app_branch.id}`}
                      textVariant="subtext"
                    >
                      <Text as="span" variant="subtext" family="mono">
                        {install.app_branch.name}
                      </Text>
                    </Link>
                  ) : (
                    <Text as="span" variant="subtext" theme="neutral">
                      None
                    </Text>
                  )}
                  {branchAction}
                </span>
              </LabeledValue>

              <LabeledValue label="Managed by">
                {isManagedByConfig ? (
                  <span className="flex items-center gap-1.5">
                    <Icon
                      variant="FileCodeIcon"
                      size={13}
                      className="text-cool-grey-400"
                    />
                    <Link
                      href={`${installPath}/configuration/config-file`}
                      textVariant="subtext"
                    >
                      Install config
                    </Link>
                  </span>
                ) : (
                  <Text as="span" variant="subtext">
                    Dashboard
                  </Text>
                )}
              </LabeledValue>
            </div>
          </Card>

          <Card
            className="!p-4 !shadow-none"
            aria-label="Install status summary"
          >
            {statuses}
          </Card>
        </div>
      </div>
    </PageHeader>
  )
}

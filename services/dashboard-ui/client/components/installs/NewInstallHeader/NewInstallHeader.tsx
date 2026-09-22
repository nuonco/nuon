import type { ReactNode } from 'react'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageHeader } from '@/components/layout/PageHeader'
import type { TInstall } from '@/types'

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

  return (
    <PageHeader>
      <div className="flex flex-col gap-4 w-full">
        <div className="flex items-start justify-between gap-4 w-full">
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
          <Card className="!p-3 !gap-2 !shadow-none w-full">
            <div className="flex items-center gap-x-6 gap-y-1 flex-wrap">
              <span className="flex items-center gap-1.5">
                <Text as="span" variant="subtext" theme="neutral">
                  App
                </Text>
                <Link
                  href={`/${orgId}/apps/${install.app_id}`}
                  textVariant="subtext"
                >
                  {install.app?.name}
                </Link>
              </span>

              <span className="flex items-center gap-1.5">
                <Text as="span" variant="subtext" theme="neutral">
                  App branch
                </Text>
                {install.app_branch ? (
                  <>
                    <Icon
                      variant="GitBranchIcon"
                      size={13}
                      className="text-cool-grey-400"
                    />
                    <Link
                      href={`/${orgId}/apps/${install.app_id}/branches/${install.app_branch.id}`}
                      textVariant="subtext"
                    >
                      <Text as="span" variant="subtext" family="mono">
                        {install.app_branch.name}
                      </Text>
                    </Link>
                  </>
                ) : (
                  <Text as="span" variant="subtext" theme="neutral">
                    None
                  </Text>
                )}
                {branchAction}
              </span>
            </div>

            <div className="flex items-center gap-x-6 gap-y-1 flex-wrap">
              <span className="flex items-center gap-1.5">
                <Text as="span" variant="subtext" theme="neutral">
                  Managed by
                </Text>
                {isManagedByConfig ? (
                  <>
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
                  </>
                ) : (
                  <Text as="span" variant="subtext">
                    Dashboard
                  </Text>
                )}
              </span>
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

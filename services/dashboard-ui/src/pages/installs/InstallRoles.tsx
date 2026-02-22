import { useParams } from 'react-router-dom'
import { useQuery } from '@/hooks/use-query'
import { BackToTop } from '@/components/common/BackToTop'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { IAMRoles, IAMRolesSkeleton } from '@/components/roles/IAMRoles'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import type { TAppConfig } from '@/types'

const InstallRolesError = ({
  title = 'Unable to load roles',
  message = 'We encountered an issue loading your roles. Please try refreshing the page or contact support if the problem persists.',
}: {
  title?: string
  message?: string
}) => {
  return (
    <EmptyState variant="table" emptyMessage={message} emptyTitle={title} />
  )
}

const InstallRolesContent = ({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId: string
  appId: string
  orgId: string
}) => {
  const { data: config, error, isLoading } = useQuery<TAppConfig>({
    path: `/api/ctl-api/v1/apps/${appId}/configs/${appConfigId}?recurse=true`,
  })

  if (error) {
    return <InstallRolesError />
  }

  if (isLoading && !config) {
    return <IAMRolesSkeleton />
  }

  if (!config?.permissions?.aws_iam_roles?.length) {
    return (
      <InstallRolesError
        title="No roles found"
        message="You don't have any roles assigned yet. Contact your administrator to get access to roles."
      />
    )
  }

  return <IAMRoles appConfig={config} />
}

export default function InstallRoles() {
  const { installId, orgId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  if (!installId || !orgId || !install) {
    return null
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/roles`,
            text: 'Roles',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          IAM roles
        </Text>
        <Text variant="subtext" theme="neutral">
          View the IAM roles that your install uses to access customer AWS
          resources.
        </Text>
      </HeadingGroup>

      <InstallRolesContent
        appConfigId={install.app_config_id}
        appId={install.app_id}
        orgId={orgId}
      />

      <BackToTop />
    </PageSection>
  )
}
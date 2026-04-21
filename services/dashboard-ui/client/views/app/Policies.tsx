import { useQuery } from '@tanstack/react-query'
import { PoliciesTable, policiesTableColumns } from '@/components/policies/PoliciesTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppPoliciesConfigs } from '@/lib'

export const Policies = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: policiesConfigs, isLoading } = useQuery({
    queryKey: ['app-policies-configs', org?.id, app?.id],
    queryFn: () => getAppPoliciesConfigs({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const latestConfig = policiesConfigs
    ?.slice()
    .sort((a, b) => {
      const dateA = a.created_at ? new Date(a.created_at).getTime() : 0
      const dateB = b.created_at ? new Date(b.created_at).getTime() : 0
      return dateB - dateA
    })
    .at(0)
  const policies = latestConfig?.policies ?? []

  return (
    <PageSection>
      <PageTitle title={`Policies | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/policies`, text: 'Policies' },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App policies
        </Text>
        <Text variant="subtext" theme="neutral">
          Define validation rules that run against builds and deploys.
        </Text>
      </HeadingGroup>

      <div className="flex flex-auto">
        {isLoading ? (
          <TableSkeleton columns={policiesTableColumns} skeletonRows={5} />
        ) : (
          <PoliciesTable
            policies={policies}
            orgId={org?.id}
            appId={app?.id}
          />
        )}
      </div>
    </PageSection>
  )
}

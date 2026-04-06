import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useBranch } from '@/hooks/use-branch'
import { getBranchConfigs } from '@/lib'
import { BranchProvider } from '@/providers/branch-provider'

const BranchConfigsContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const branchId = params.branchId as string

  const { data: configs } = useQuery({
    queryKey: ['branch-configs', org.id, app.id, branchId],
    queryFn: () => getBranchConfigs({ orgId: org.id!, appId: app.id!, branchId }),
    enabled: !!org.id && !!app.id && !!branchId,
  })

  return (
    <PageSection>
      <PageTitle title={`Configs | ${branch?.name ?? 'Branch'} | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}`, text: branch?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}/configs`, text: 'Configs' },
        ]}
      />
      <Text variant="h3" weight="strong">
        Branch Configs
      </Text>
      <Text variant="body">
        {configs?.length || 0} config{configs?.length !== 1 ? 's' : ''} found
      </Text>
    </PageSection>
  )
}

export const BranchConfigs = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId}>
      <BranchConfigsContent />
    </BranchProvider>
  )
}

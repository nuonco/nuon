import { useParams } from 'react-router'
import { BranchConfigsTable } from '@/components/branches/BranchConfigsTable'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import { BranchProvider } from '@/providers/branch-provider'

const BranchConfigsContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <PageSection>
      <PageTitle
        segments={[branch?.name ? `${branch.name} configs` : 'Configs', app?.name]}
      />
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
      <SectionHeader
        title="Branch configs"
        description="App config versions synced from this branch. Select a version to see its contents."
      />

      <BranchConfigsTable branchId={branchId} />
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

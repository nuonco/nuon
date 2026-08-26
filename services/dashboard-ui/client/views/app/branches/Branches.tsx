import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { BranchCards } from '@/components/branches/BranchCards'
import { BranchesTable } from '@/components/branches/BranchesTable'
import { CreateBranchButton } from '@/components/branches/CreateBranchModal'

export const Branches = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const hasNewAppIA = useNewAppIA()

  const breadcrumbs = [
    { path: `/${org?.id}`, text: org?.name },
    { path: `/${org?.id}/apps`, text: 'Apps' },
    { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
    ...(hasNewAppIA
      ? []
      : [
          {
            path: `/${org?.id}/apps/${app?.id}/branches`,
            text: 'Branches',
          },
        ]),
  ]

  return (
    <PageSection>
      <PageTitle segments={['Branches', app?.name]} />
      <Breadcrumbs breadcrumbs={breadcrumbs} />
      <SectionHeader
        title="Branches"
        description="Manage app branches for version control and deployment"
        actions={<CreateBranchButton variant="primary" />}
      />
      {hasNewAppIA ? <BranchCards shouldPoll /> : <BranchesTable shouldPoll />}
    </PageSection>
  )
}

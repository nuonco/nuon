import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
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
      <div className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="strong">
            Branches
          </Text>
          <Text variant="subtext" theme="neutral">
            Manage app branches for version control and deployment
          </Text>
        </HeadingGroup>
        <CreateBranchButton variant="primary" />
      </div>
      {hasNewAppIA ? <BranchCards shouldPoll /> : <BranchesTable shouldPoll />}
    </PageSection>
  )
}

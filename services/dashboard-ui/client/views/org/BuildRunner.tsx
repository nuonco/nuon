import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageHeadingGroup } from '@/components/layout/PageHeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ControlPlaneRecentActivity } from '@/components/runners/ControlPlaneRecentActivity'
import { useOrg } from '@/hooks/use-org'

export const BuildRunner = () => {
  const { org } = useOrg()

  return (
    <PageLayout className="pb-6">
      <PageTitle title="Builds" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/runner`, text: 'Builds' },
        ]}
      />
      <PageHeader>
        <PageHeadingGroup
          title="Builds"
          subtitle="Component builds run on Nuon's control plane."
        />
      </PageHeader>
      <PageContent>
        <PageSection>
          <ControlPlaneRecentActivity shouldPoll />
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}

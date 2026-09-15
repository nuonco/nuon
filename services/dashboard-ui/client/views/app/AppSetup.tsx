import { PageSection } from '@/components/layout/PageSection'
import { PageLayout } from '@/components/layout/PageLayout'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const AppSetup = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="Set up app" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/setup`, text: 'Set up app' },
        ]}
      />
      <PageLayout>
        <PageSection>
          <SectionHeader
            title="Set up app"
            description="App setup will be added in a follow-up."
          />
        </PageSection>
      </PageLayout>
    </>
  )
}

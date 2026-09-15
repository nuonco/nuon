import { PageSection } from '@/components/layout/PageSection'
import { PageLayout } from '@/components/layout/PageLayout'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const InstallSetup = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="Set up install" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/setup`, text: 'Set up install' },
        ]}
      />
      <PageLayout>
        <PageSection>
          <SectionHeader
            title="Set up install"
            description="Install setup will be added in a follow-up."
          />
        </PageSection>
      </PageLayout>
    </>
  )
}

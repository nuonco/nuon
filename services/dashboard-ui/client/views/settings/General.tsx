import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { UpdateOrgName } from '@/components/orgs/UpdateOrgName'
import { useOrg } from '@/hooks/use-org'

export const GeneralSettings = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="General" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/settings`, text: 'Settings' },
          { path: `/${org.id}/settings/general`, text: 'General' },
        ]}
      />
      <PageSection>
        <SectionHeader
          title="General"
          description="Update the name of this organization."
        />
        <UpdateOrgName />
      </PageSection>
    </>
  )
}

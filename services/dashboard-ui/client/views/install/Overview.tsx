import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Overview = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
        ]}
      />

      {/* install overview content here */}
    </PageSection>
  )
}

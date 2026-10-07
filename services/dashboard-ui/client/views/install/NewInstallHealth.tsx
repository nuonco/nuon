import type { ReactNode } from 'react'
import { InstallHealth } from '@/components/installs/InstallHealth'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const NewInstallHealth = ({ children }: { children?: ReactNode }) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const path = `/${org?.id}/installs/${install?.id}/health`

  return (
    <PageSection>
      <PageTitle segments={['Health', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path, text: 'Health' },
        ]}
      />
      <SectionHeader
        title="Health"
        description="Inspect the latest health of components and managed resources."
      />
      {children ?? <InstallHealth />}
    </PageSection>
  )
}

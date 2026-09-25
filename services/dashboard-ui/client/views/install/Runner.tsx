import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallRunner } from '@/components/runners/InstallRunner'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Runner = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Install runner', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          {
            path: `/${org?.id}/installs/${install?.id}`,
            text: install?.name,
          },
          {
            path: `/${org?.id}/installs/${install?.id}/runner`,
            text: 'Install runner',
          },
        ]}
      />
      <PageSection>
        <InstallRunner />
      </PageSection>
    </>
  )
}

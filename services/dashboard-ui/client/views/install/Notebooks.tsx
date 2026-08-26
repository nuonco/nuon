import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { CreateNotebookButton } from '@/components/notebooks/CreateNotebook'
import { NotebooksTable } from '@/components/notebooks/NotebooksTable'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Notebooks = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle segments={['Notebooks', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/notebooks`,
            text: 'Notebooks',
          },
        ]}
      />
      <SectionHeader
        title="Notebooks"
        description="Run commands on the runner for this install."
        actions={<CreateNotebookButton />}
      />
      <NotebooksTable />
    </PageSection>
  )
}

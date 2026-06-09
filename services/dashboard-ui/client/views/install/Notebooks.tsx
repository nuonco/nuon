import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
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
      <PageTitle title={`Notebooks | ${install?.name}`} />
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
      <div className="flex items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Notebooks
          </Text>
          <Text variant="subtext" theme="neutral">
            Run commands on the runner for this install.
          </Text>
        </HeadingGroup>
        <CreateNotebookButton />
      </div>
      <NotebooksTable />
    </PageSection>
  )
}

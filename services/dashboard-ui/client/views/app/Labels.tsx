import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { AppLabels } from '@/components/apps/AppLabels'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Labels = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <PageSection>
      <PageTitle title={`Labels | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/labels`, text: 'Labels' },
        ]}
      />
      <AppLabels />
    </PageSection>
  )
}

import { RunbookBuilder as RunbookBuilderFeature } from '@/components/runbooks/RunbookBuilder'
import { Text } from '@/components/common/Text'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export function RunbookBuilder() {
  const { org } = useOrg()
  const { app } = useApp()
  return (
    <PageSection>
      <PageTitle title={`Build runbook | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/runbooks`, text: 'Runbooks' },
          {
            path: `/${org?.id}/apps/${app?.id}/runbooks/builder`,
            text: 'Build runbook',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="h3" weight="strong">
          Build runbook
        </Text>
        <Text variant="subtext" theme="neutral">
          Compose operations, then copy or download the TOML for the app
          runbooks directory.
        </Text>
      </HeadingGroup>
      <RunbookBuilderFeature />
    </PageSection>
  )
}

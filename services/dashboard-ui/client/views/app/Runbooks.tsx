import { RunbooksTable } from '@/components/runbooks/RunbooksTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Runbooks = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <PageSection>
      <PageTitle title={`Runbooks | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/runbooks`, text: 'Runbooks' },
        ]}
      />
      <div className="flex flex-wrap items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            App runbooks
          </Text>
          <Text variant="subtext" theme="neutral">
            Define and manage operational procedures for your installs.
          </Text>
        </HeadingGroup>
        <Button
          variant="primary"
          href={`/${org?.id}/apps/${app?.id}/runbooks/builder`}
        >
          Build runbook
        </Button>
      </div>
      <RunbooksTable />
    </PageSection>
  )
}

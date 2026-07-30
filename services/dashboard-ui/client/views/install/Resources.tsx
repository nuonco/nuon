import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { HealthTimeline } from '@/components/install-health/HealthTimeline'
import { InstallResourcesTable } from '@/components/install-resources/InstallResourcesTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Resources = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle title={`Resources | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/resources`,
            text: 'Resources',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Resources
        </Text>
        <Text variant="subtext" theme="neutral">
          Live resources managed by this install's components, with per-resource health.
        </Text>
      </HeadingGroup>

      <Card>
        <HealthTimeline shouldPoll />
      </Card>

      <InstallResourcesTable shouldPoll />
    </PageSection>
  )
}

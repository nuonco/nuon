import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const DriftEvaluation = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle title={`Drift evaluation | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/drift`,
            text: 'Drift evaluation',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Drift evaluation
        </Text>
        <Text variant="subtext" theme="neutral">
          Evaluate and monitor drift across your install components and sandbox.
        </Text>
      </HeadingGroup>
      <EmptyState
        emptyTitle="Coming soon"
        emptyMessage="Drift evaluation is being built. Check back soon."
        variant="search"
      />
    </PageSection>
  )
}

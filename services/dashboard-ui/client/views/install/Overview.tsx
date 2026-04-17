import { useQuery } from '@tanstack/react-query'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Markdown } from '@/components/common/Markdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallDetailsButton } from '@/components/installs/ArchitectureDiagram'
import { ViewCurrentInputsButton } from '@/components/installs/management/ViewCurrentInputs'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallReadme } from '@/lib'

export const Overview = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { data: readme } = useQuery({
    queryKey: ['install-readme', org?.id, install?.id],
    queryFn: () =>
      getInstallReadme({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <PageSection>
      <PageTitle title={`Overview | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
        ]}
      />

      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Install overview
          </Text>
          <Text variant="subtext" theme="neutral">
            View the install README, architecture, and current inputs.
          </Text>
        </HeadingGroup>
        <div className="flex items-center gap-2">
          <InstallDetailsButton variant="secondary" />
          <ViewCurrentInputsButton variant="secondary" />
        </div>
      </div>

      {readme?.readme ? (
        <div className="flex flex-col gap-4">
          {readme.warnings?.map((warn, i) => (
            <div
              key={i}
              className="p-3 rounded bg-yellow-50 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200 text-sm"
            >
              {warn}
            </div>
          ))}
          <Markdown content={readme.readme} mode="install" />
        </div>
      ) : (
        <EmptyState
          emptyTitle="No README"
          emptyMessage="No install README found."
          variant="diagram"
        />
      )}
    </PageSection>
  )
}

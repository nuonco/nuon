import { useQuery } from '@tanstack/react-query'
import { Markdown } from '@/components/common/Markdown'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { ReadmeWarnings } from '@/components/installs/ReadmeWarnings'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallReadme } from '@/lib'

export const Readme = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: result, isLoading } = useQuery({
    queryKey: ['install-readme', org?.id, install?.id],
    queryFn: () =>
      getInstallReadme({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const readme = result

  return (
    <PageSection>
      <PageTitle title={`README | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/readme`,
            text: 'Readme',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install readme
        </Text>
      </HeadingGroup>

      {isLoading ? null : readme?.readme ? (
        <div className="flex flex-col gap-3">
          <ReadmeWarnings warnings={readme.warnings} />
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

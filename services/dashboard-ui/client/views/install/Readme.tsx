import { useQuery } from '@tanstack/react-query'
import { Markdown } from '@/components/common/Markdown'
import { BackToTop } from '@/components/common/BackToTop'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallReadme } from '@/lib'

const CONTAINER_ID = 'install-readme-page'

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
    <PageSection id={CONTAINER_ID} isScrollable>
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
      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}

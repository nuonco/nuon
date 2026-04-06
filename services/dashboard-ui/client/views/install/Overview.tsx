import { useQuery } from '@tanstack/react-query'
import { Markdown } from '@/components/common/Markdown'
import { BackToTop } from '@/components/common/BackToTop'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallReadme } from '@/lib'

const CONTAINER_ID = 'install-overview-page'

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
    <PageSection id={CONTAINER_ID} className="!pt-0" isScrollable>
      <PageTitle title={`Overview | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
        ]}
      />

      <div className="py-6 flex flex-col gap-4">
        <Text variant="base" weight="strong">
          README
        </Text>
        {readme?.readme ? (
          <>
            {readme.warnings?.map((warn, i) => (
              <div
                key={i}
                className="p-3 rounded bg-yellow-50 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200 text-sm"
              >
                {warn}
              </div>
            ))}
            <Markdown content={readme.readme} mode="install" />
          </>
        ) : (
          <EmptyState
            emptyTitle="No README"
            emptyMessage="No install README found."
            variant="diagram"
          />
        )}
      </div>
      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}

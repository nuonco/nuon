import { useQuery } from '@tanstack/react-query'
import { Markdown } from '@/components/common/Markdown'
import { BackToTop } from '@/components/common/BackToTop'
import { EmptyState } from '@/components/common/EmptyState'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallReadme, getInstallCurrentInputs } from '@/lib'

const CONTAINER_ID = 'install-overview-page'

export const Overview = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: readmeResult } = useQuery({
    queryKey: ['install-readme', org?.id, install?.id],
    queryFn: () =>
      getInstallReadme({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const { data: inputsResult } = useQuery({
    queryKey: ['install-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const readme = readmeResult
  const inputs = inputsResult
  const inputEntries = inputs?.redacted_values
    ? Object.entries(inputs.redacted_values).map(([key, value]) => ({
        key,
        value,
      }))
    : []

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

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8 py-6 pr-6 flex flex-col gap-4 min-w-0">
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
              <Markdown content={readme.readme} />
            </>
          ) : (
            <EmptyState
              emptyTitle="No README"
              emptyMessage="No install README found."
              variant="diagram"
            />
          )}
        </div>

        <div className="md:col-span-4 py-6 pl-6 flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Current inputs
          </Text>
          {inputEntries.length > 0 ? (
            <PropertyGrid values={inputEntries} />
          ) : (
            <Text variant="subtext" theme="neutral">
              No inputs configured.
            </Text>
          )}
        </div>
      </div>
      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}

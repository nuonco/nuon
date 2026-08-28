import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Expand } from '@/components/common/Expand'
import { Markdown } from '@/components/common/Markdown'
import { ReadmeWarnings } from '@/components/installs/ReadmeWarnings'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallDetailsButton } from '@/components/installs/ArchitectureDiagram'
import { ViewCurrentInputsButton } from '@/components/installs/management/ViewCurrentInputs'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallReadme } from '@/lib'
import { isCustomerManagedInstall } from '@/utils/install-utils'
import { CustomerManagedSnapshotOverview } from '@/components/customer-managed-support/SnapshotOverview'

export const Overview = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const isCustomerManaged = isCustomerManagedInstall(install)
  const { data: readme } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-readme', org?.id, install?.id],
    queryFn: () => getInstallReadme({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id && !isCustomerManaged,
  })

  if (isCustomerManaged) {
    return (
      <PageSection>
        <PageTitle title={`Overview | ${install.name}`} />
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${org.id}`, text: org.name },
            { path: `/${org.id}/installs`, text: 'Installs' },
            { path: `/${org.id}/installs/${install.id}`, text: install.name },
          ]}
        />
        <CustomerManagedSnapshotOverview />
      </PageSection>
    )
  }

  return (
    <PageSection>
      <PageTitle segments={['Overview', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
        ]}
      />

      <SectionHeader
        title="Install overview"
        description="View the install README, architecture, and current inputs."
        actions={
          <>
            <InstallDetailsButton variant="secondary" />
            <ViewCurrentInputsButton variant="secondary" />
          </>
        }
      />

      {readme?.readme ? (
        <div className="flex flex-col gap-4">
          <ReadmeWarnings warnings={readme.warnings} />
          {readme.warnings?.length ? (
            <Expand
              id="incomplete-readme"
              heading="View incomplete README"
              className="border rounded-lg"
            >
              <div className="p-4 border-t max-h-[32rem] overflow-y-auto">
                <Markdown content={readme.readme} mode="install" />
              </div>
            </Expand>
          ) : (
            <Markdown content={readme.readme} mode="install" />
          )}
        </div>
      ) : (
        // Blue informative Banner (theme="info") replaces the previous
        // EmptyState when the rendered README is empty. The customer
        // hits this before the install reaches an active state — any
        // `original` README still needs live install data to template
        // against, so the right UX is to tell them when it'll show up
        // rather than imply "no README exists".
        <Banner theme="info">
          The readme will render after the install is active and live.
        </Banner>
      )}
    </PageSection>
  )
}

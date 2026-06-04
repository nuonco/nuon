import { useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { ReadmeWarnings } from '@/components/installs/ReadmeWarnings'
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
          <ReadmeWarnings warnings={readme.warnings} />
          <Markdown content={readme.readme} mode="install" />
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

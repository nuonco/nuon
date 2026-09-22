import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { CurrentAppBranchRun } from '@/components/install-updates/CurrentAppBranchRun'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import { ReadmeWarnings } from '@/components/installs/ReadmeWarnings'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useCurrentAppBranchRun } from '@/hooks/use-current-app-branch-run'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallReadme } from '@/lib'

export const Overview = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { data: readme } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-readme', org?.id, install?.id],
    queryFn: () => getInstallReadme({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })
  const { run: currentAppBranchRun, isLoading: isAppBranchRunLoading } =
    useCurrentAppBranchRun()

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
        description="Current install status and applied app configuration."
      />

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <div className="flex flex-col gap-1">
            <Text variant="h3">Install status</Text>
            <Text variant="subtext" theme="neutral">
              Runner, sandbox, and component status.
            </Text>
          </div>
          {install ? <InstallStatuses install={install} /> : null}
        </Card>
        <CurrentAppBranchRun
          run={currentAppBranchRun}
          orgId={org?.id}
          appId={install?.app_id}
          branchName={install?.app_branch?.name}
          isLoading={isAppBranchRunLoading}
        />
      </div>

      <SectionHeader
        title="README"
        description="Instructions and details rendered for this install."
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
        // An `original` README still needs live install data to template
        // against, so an empty render means "not ready yet", not "none exists".
        <Banner theme="info">
          The readme will render after the install is active and live.
        </Banner>
      )}
    </PageSection>
  )
}

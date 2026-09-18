import { Badge } from '@/components/common/Badge'
import { Link } from '@/components/common/Link'
import { CurrentAppBranchRun } from '@/components/install-updates/CurrentAppBranchRun'
import { InstallUpdatesTimeline } from '@/components/install-updates/InstallUpdatesTimeline'
import { ChangeAppBranchButton } from '@/components/installs/management/ChangeAppBranch'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useCurrentAppBranchRun } from '@/hooks/use-current-app-branch-run'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Updates = () => {
  const { org } = useOrg()
  const { install, refresh } = useInstall()
  const { run: currentAppBranchRun } = useCurrentAppBranchRun()
  const hasAppBranchesUI = !!org?.features?.['app-branches-ui']

  return (
    <PageSection>
      <PageTitle segments={['Updates', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/updates`,
            text: 'Updates',
          },
        ]}
      />
      <SectionHeader
        title="Updates"
        description="App config updates applied to this install, including stack and component impacts."
        status={
          install?.app_branch_id ? (
            <Badge size="sm" theme="info">
              {install.app_branch?.name || 'branch'}
            </Badge>
          ) : null
        }
        actions={
          install && hasAppBranchesUI ? (
            <div className="flex items-center gap-2">
              {install.app_branch_id ? (
                <Link
                  href={`/${org?.id}/apps/${install.app_id}/branches/${install.app_branch_id}`}
                >
                  View branch
                </Link>
              ) : null}
              <ChangeAppBranchButton
                compact
                install={install}
                onSuccess={refresh}
              />
            </div>
          ) : null
        }
      />

      <CurrentAppBranchRun
        run={currentAppBranchRun}
        orgId={org?.id}
        appId={install?.app_id}
      />

      <InstallUpdatesTimeline shouldPoll />
    </PageSection>
  )
}

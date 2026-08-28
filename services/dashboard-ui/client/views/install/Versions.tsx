import { Badge } from '@/components/common/Badge'
import { Link } from '@/components/common/Link'
import { InstallVersionsTimeline } from '@/components/install-versions/InstallVersionsTimeline'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Versions = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle segments={['Versions', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/versions`,
            text: 'Versions',
          },
        ]}
      />
      <SectionHeader
        title="App branch runs"
        description="History of app config updates applied to this install."
        status={
          install?.app_branch_id ? (
            <Badge size="sm" theme="info">
              {install.app_branch?.name || 'branch'}
            </Badge>
          ) : null
        }
        actions={
          install?.app_branch_id ? (
            <Link
              href={`/${org?.id}/apps/${install?.app_id}/branches/${install?.app_branch_id}`}
            >
              View branch
            </Link>
          ) : null
        }
      />

      <InstallVersionsTimeline />
    </PageSection>
  )
}

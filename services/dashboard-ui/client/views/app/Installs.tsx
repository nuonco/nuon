import { AppInstallsTable } from '@/components/apps/AppInstallsTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Installs = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <PageSection isScrollable>
      <PageTitle title={`Installs | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/installs`, text: 'Installs' },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App installs
        </Text>
      </HeadingGroup>
      <AppInstallsTable appId={app?.id} />
    </PageSection>
  )
}

import { CreateInstallButton } from '@/components/apps/CreateInstall'
import { CreateBranchButton } from '@/components/branches/CreateBranchModal'
import { AppFleet } from '@/components/fleet/AppFleet'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'

export const Fleet = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const hasNewAppIA = useNewAppIA()

  return (
    <>
      <PageTitle segments={['Fleet', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          ...(hasNewAppIA
            ? []
            : [
                { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Fleet' },
              ]),
        ]}
      />
      <ListPage
        title="Fleet"
        description="Installs grouped by app branch and install group, with rollout, health, and runner state."
        createAction={<CreateBranchButton variant="primary" />}
      >
        <AppFleet
          emptyAction={app?.runner_config ? <CreateInstallButton /> : undefined}
        />
      </ListPage>
    </>
  )
}

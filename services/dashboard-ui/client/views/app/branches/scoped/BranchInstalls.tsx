import { useParams } from 'react-router'
import { AppInstallsTable } from '@/components/apps/AppInstallsTable'
import { CreateInstallButton } from '@/components/installs/CreateInstall'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { Installs } from '../../Installs'
import { BranchTabPage } from '../tabs/BranchTabPage'

const BranchInstallsContent = () => {
  const { app } = useApp()
  const params = useParams()
  const branchId = params.branchId as string

  const createButton = app?.runner_config ? (
    <CreateInstallButton initialApp={app} variant="secondary" />
  ) : undefined

  return (
    <BranchTabPage
      tab="Installs"
      heading="Installs"
      subheading="Installs assigned to this branch's deployment plan."
      actions={createButton}
    >
      <AppInstallsTable
        appId={app?.id}
        filterBranchId={branchId}
        emptyTitle="No installs in this deployment plan"
        emptyMessage="Create an install and assign it to one of this branch's deployment groups."
        emptyAction={createButton}
      />
    </BranchTabPage>
  )
}

export const BranchInstalls = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchInstallsContent /> : <Installs />
}

import { InstallActionsList } from '@/components/actions/InstallActionsList'
import { ActivityList } from '@/components/operations/ActivityList'
import { ActivityPins } from '@/components/operations/ActivityPins'
import { InstallPolicyReports } from '@/components/policies/InstallPolicyReports'
import { InstallRunbooksList } from '@/components/runbooks/InstallRunbooksList'
import { InstallRunner } from '@/components/runners/InstallRunner'
import { InstallWorkflowPanelController } from '@/components/workflows/InstallWorkflowPanel'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'

export const NewInstallActivity = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Activity', install?.name]} />
      <div className="flex flex-col gap-6">
        <ActivityPins />
        <ActivityList shouldPoll />
      </div>
      <InstallWorkflowPanelController />
    </>
  )
}

export const NewInstallActions = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Actions', install?.name]} />
      <InstallActionsList />
      <InstallWorkflowPanelController />
    </>
  )
}

export const NewInstallRunbooks = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Runbooks', install?.name]} />
      <InstallRunbooksList />
      <InstallWorkflowPanelController />
    </>
  )
}

export const NewInstallPolicies = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Policies', install?.name]} />
      <InstallPolicyReports variant="cards" />
    </>
  )
}

export const NewInstallRunner = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Install runner', install?.name]} />
      <InstallRunner variant="embedded" />
    </>
  )
}

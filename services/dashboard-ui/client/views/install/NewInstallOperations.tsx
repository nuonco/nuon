import { InstallActionsList } from '@/components/actions/InstallActionsList'
import { Text } from '@/components/common/Text'
import { ActivityPins } from '@/components/operations/ActivityPins'
import { InstallPolicyReports } from '@/components/policies/InstallPolicyReports'
import { InstallRunbooksList } from '@/components/runbooks/InstallRunbooksList'
import { InstallRunner } from '@/components/runners/InstallRunner'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'

export const NewInstallActivity = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Activity', install?.name]} />
      <div className="flex flex-col gap-6">
        <ActivityPins />
        <Text theme="neutral">Placeholder content.</Text>
      </div>
    </>
  )
}

export const NewInstallActions = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Actions', install?.name]} />
      <InstallActionsList />
    </>
  )
}

export const NewInstallRunbooks = () => {
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Runbooks', install?.name]} />
      <InstallRunbooksList />
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

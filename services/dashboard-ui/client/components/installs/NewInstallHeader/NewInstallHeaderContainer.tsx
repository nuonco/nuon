import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { InstallStatusesContainer } from '@/components/installs/InstallStatuses'
import { useOpenInstallSettings } from '@/components/installs/InstallSettingsPanel'
import { ChangeAppBranchButton } from '@/components/installs/management/ChangeAppBranch'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { NewInstallHeader } from './NewInstallHeader'

export const NewInstallHeaderContainer = () => {
  const { org } = useOrg()
  const { install, labelColors, refresh } = useInstall()
  const openSettings = useOpenInstallSettings()

  if (!install) return null

  return (
    <NewInstallHeader
      install={install}
      labelColors={labelColors}
      orgId={org?.id}
      branchAction={
        <ChangeAppBranchButton compact install={install} onSuccess={refresh} />
      }
      settingsAction={
        <Button
          variant="secondary"
          onClick={openSettings}
          aria-label="Install settings"
        >
          <Icon variant="GearIcon" size={16} />
        </Button>
      }
      statuses={<InstallStatusesContainer />}
    />
  )
}

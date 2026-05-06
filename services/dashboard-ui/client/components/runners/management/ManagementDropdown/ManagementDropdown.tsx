import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { UpdateRunnerButton } from '../UpdateRunner'
import type { TRunnerSettings, TRunner } from '@/types'

interface IManagementDropdown {
  runner: TRunner
  isInstallRunner?: boolean
  settings: TRunnerSettings
}

export const ManagementDropdown = ({
  runner,
  isInstallRunner = false,
  settings,
}: IManagementDropdown) => {
  return (
    <Dropdown
      id={`runner-${runner.id}-mgmt`}
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> {isInstallRunner ? 'Manage runner' : 'Manage build runner'}
        </>
      }
      alignment="right"
      variant={!isInstallRunner ? 'primary' : 'secondary'}
    >
      <Menu>
        {!isInstallRunner ? (
          <UpdateRunnerButton
            settings={settings}
            field="container_image_tag"
            label="Update runner tag"
            modalHeading="Update runner tag"
            inputLabel="Enter the runner tag you'd like to update to."
            inputPlaceholder="runner tag"
            submitLabel="Update runner tag"
            isMenuButton
          />
        ) : (
          <>
            <UpdateRunnerButton
              settings={settings}
              field="binary_version"
              label="Update manager version"
              modalHeading="Update manager version"
              inputLabel="Enter the manager version you'd like to update to."
              inputPlaceholder="manager version"
              submitLabel="Update manager version"
              isMenuButton
            />
            <UpdateRunnerButton
              settings={settings}
              field="container_image_tag"
              label="Update instance version"
              modalHeading="Update instance version"
              inputLabel="Enter the instance version you'd like to update to."
              inputPlaceholder="instance version"
              submitLabel="Update instance version"
              isMenuButton
            />
          </>
        )}
      </Menu>
    </Dropdown>
  )
}

import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { UpdateRunnerButton } from '../UpdateRunner'
import type { TRunnerSettings, TRunner } from '@/types'

interface IManagementDropdown {
  runner: TRunner
  isInstallRunner?: boolean
  settings: TRunnerSettings
  hasMngProcess?: boolean
  hasInstanceProcess?: boolean
}

const buildRunnerTagAction = {
  field: 'container_image_tag' as const,
  label: 'Update runner tag',
  modalHeading: 'Update runner tag',
  inputLabel: "Enter the runner tag you'd like to update to.",
  inputPlaceholder: 'runner tag',
  submitLabel: 'Update runner tag',
}

const installManagerAction = {
  field: 'binary_version' as const,
  label: 'Update manager version',
  modalHeading: 'Update manager version',
  inputLabel: "Enter the manager version you'd like to update to.",
  inputPlaceholder: 'manager version',
  submitLabel: 'Update manager version',
}

const installInstanceAction = {
  field: 'container_image_tag' as const,
  label: 'Update instance version',
  modalHeading: 'Update instance version',
  inputLabel: "Enter the instance version you'd like to update to.",
  inputPlaceholder: 'instance version',
  submitLabel: 'Update instance version',
}

export const ManagementDropdown = ({
  runner,
  isInstallRunner = false,
  settings,
  hasMngProcess,
  hasInstanceProcess,
}: IManagementDropdown) => {
  if (!isInstallRunner) {
    return (
      <UpdateRunnerButton
        settings={settings}
        variant="primary"
        {...buildRunnerTagAction}
      />
    )
  }

  const showMng = hasMngProcess !== false
  const showInstance = hasInstanceProcess !== false
  const actions = [
    showMng ? installManagerAction : null,
    showInstance ? installInstanceAction : null,
  ].filter((a): a is NonNullable<typeof a> => !!a)

  if (actions.length === 0) return null

  if (actions.length === 1) {
    return (
      <UpdateRunnerButton
        settings={settings}
        variant="secondary"
        {...actions[0]}
      />
    )
  }

  return (
    <Dropdown
      id={`runner-${runner.id}-mgmt`}
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage install runner
        </>
      }
      alignment="right"
      variant="secondary"
    >
      <Menu>
        {actions.map((action) => (
          <UpdateRunnerButton
            key={action.field}
            settings={settings}
            isMenuButton
            {...action}
          />
        ))}
      </Menu>
    </Dropdown>
  )
}

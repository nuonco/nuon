import { useState } from 'react'
import { type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import {
  CreateInstallFromAppContainer,
  type ICreateFromAppState,
} from '@/components/installs/CreateInstall/CreateInstallFromAppContainer'
import { StackOnlyCheckbox } from '@/components/installs/forms/InstallForm'
import { useApp } from '@/hooks/use-app'
import { useSurfaces } from '@/hooks/use-surfaces'
import { CreateInstallButton as CreateInstallButtonComponent } from './CreateInstall'

const noop = () => {}

const INITIAL_STATE: ICreateFromAppState = {
  canSubmit: false,
  submit: noop,
  isSubmitting: false,
  phase: 'form',
}

const CreateInstallModalContainer = ({ ...props }: IModal) => {
  const { app } = useApp()
  const [state, setState] = useState<ICreateFromAppState>(INITIAL_STATE)
  const pickingBranch = state.phase === 'select-branch'
  const pickingGroup = state.phase === 'pick-group'

  if (!app) return null

  return (
    <Modal
      {...props}
      size="xl"
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      showFooter
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon
            variant={
              pickingBranch
                ? 'GitBranchIcon'
                : pickingGroup
                  ? 'UsersIcon'
                  : 'CubeIcon'
            }
            size="24"
          />
          {pickingBranch
            ? 'Select app branch'
            : pickingGroup
              ? 'Select install group'
              : 'Create install'}
        </Text>
      }
      primaryActionTrigger={{
        children: state.isSubmitting ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Creating install
          </span>
        ) : pickingBranch ? (
          'Continue'
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Create install
          </span>
        ),
        disabled: !state.canSubmit || state.isSubmitting,
        onClick: () => state.submit(),
        variant: 'primary',
      }}
      footerActions={
        !pickingBranch && !pickingGroup && state.form ? (
          <StackOnlyCheckbox form={state.form} />
        ) : undefined
      }
    >
      <CreateInstallFromAppContainer
        app={app}
        onStateChange={setState}
        modalId={props.modalId}
      />
    </Modal>
  )
}

export const CreateInstallButtonContainer = ({
  onClick: _onClick,
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  return (
    <CreateInstallButtonComponent
      onClick={() => addModal(<CreateInstallModalContainer />)}
      {...props}
    />
  )
}

import { useState } from 'react'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TApp } from '@/types'
import { AppSelectContainer as AppSelect } from './AppSelectContainer'
import {
  CreateInstallFromAppContainer,
  type ICreateFromAppState,
} from './CreateInstallFromAppContainer'

interface ICreateInstall {
  initialApp?: TApp
}

const INITIAL_STATE: ICreateFromAppState = {
  canSubmit: false,
  submit: () => {},
  isSubmitting: false,
  phase: 'form',
}

const phaseIcon = (phase: ICreateFromAppState['phase']): TIconVariant => {
  if (phase === 'select-branch') return 'GitBranchIcon'
  if (phase === 'pick-group') return 'UsersIcon'
  return 'CubeIcon'
}

const phaseHeading = (phase: ICreateFromAppState['phase']): string => {
  if (phase === 'select-branch') return 'Select app branch'
  if (phase === 'pick-group') return 'Select install group'
  return 'Create install'
}

const primaryLabel = (state: ICreateFromAppState): React.ReactNode => {
  if (state.isSubmitting) {
    return (
      <span className="flex items-center gap-2">
        <Icon variant="Loading" />
        Creating install
      </span>
    )
  }
  if (state.phase === 'select-branch' || state.phase === 'form') {
    return (
      <span className="flex items-center gap-2">
        Continue
        <Icon variant="CaretRightIcon" />
      </span>
    )
  }
  // pick-group
  return (
    <span className="flex items-center gap-2">
      <Icon variant="PlusIcon" />
      Create install
    </span>
  )
}

export const CreateInstallModal = ({
  initialApp,
  ...props
}: ICreateInstall & IModal) => {
  const [selectedApp, setSelectedApp] = useState<TApp | undefined>(initialApp)
  const [state, setState] = useState<ICreateFromAppState>(INITIAL_STATE)

  const showForm = !!selectedApp
  const showFooter = showForm

  return (
    <Modal
      {...props}
      size={showForm ? 'xl' : 'default'}
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      showFooter={showFooter}
      heading={
        <div className="flex flex-col gap-2">
          <Text flex className="gap-4" variant="h3" weight="strong">
            <Icon variant={phaseIcon(state.phase)} size="24" />
            {showForm ? phaseHeading(state.phase) : 'Create install'}
          </Text>
          {!selectedApp && (
            <Text
              variant="body"
              className="text-cool-grey-600 dark:text-cool-grey-400"
            >
              Select an app to create an install
            </Text>
          )}
        </div>
      }
      primaryActionTrigger={
        showFooter
          ? {
              children: primaryLabel(state),
              disabled: !state.canSubmit || state.isSubmitting,
              onClick: () => state.submit(),
              variant: 'primary',
            }
          : undefined
      }
    >
      {selectedApp ? (
        <CreateInstallFromAppContainer
          app={selectedApp}
          onBack={
            initialApp
              ? undefined
              : () => {
                  setSelectedApp(undefined)
                  setState(INITIAL_STATE)
                }
          }
          onStateChange={setState}
          modalId={props.modalId}
        />
      ) : (
        <AppSelect
          onSelectApp={setSelectedApp}
          onClose={() => props.onClose?.()}
        />
      )}
    </Modal>
  )
}

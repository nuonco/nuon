import { useState } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import type { TApp } from '@/types'
import { AppSelectContainer as AppSelect } from './AppSelectContainer'
import {
  CreateInstallFromAppContainer,
  type ICreateFromAppState,
} from './CreateInstallFromAppContainer'
import { InstallationProfileWizard } from './InstallationProfileWizard'

interface ICreateInstall {
  initialApp?: TApp
}

const INITIAL_STATE: ICreateFromAppState = {
  canSubmit: false,
  submit: () => {},
  isSubmitting: false,
  phase: 'form',
}

export const CreateInstallModal = ({
  initialApp,
  ...props
}: ICreateInstall & IModal) => {
  const { org } = useOrg()
  const [selectedApp, setSelectedApp] = useState<TApp | undefined>(initialApp)
  const [state, setState] = useState<ICreateFromAppState>(INITIAL_STATE)
  const [profileSelected, setProfileSelected] = useState(false)

  const showForm = !!selectedApp
  const showBranches = state.phase === 'branches'
  const profileEnabled = !!org?.features?.['customer-managed-installs']
  const showProfile = profileEnabled && !!selectedApp && !profileSelected
  const showFooter = showForm && !showProfile && state.phase === 'form'

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
            <Icon
              variant={showBranches ? 'GitBranchIcon' : 'CubeIcon'}
              size="24"
            />
            {showBranches ? 'Connect to app branches' : 'Create install'}
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
              children: state.isSubmitting ? (
                <span className="flex items-center gap-2">
                  <Icon variant="Loading" />
                  Creating install
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  <Icon variant="PlusIcon" />
                  Create install
                </span>
              ),
              disabled: !state.canSubmit || state.isSubmitting,
              onClick: () => state.submit(),
              variant: 'primary',
            }
          : undefined
      }
    >
      {selectedApp ? (
        showProfile ? (
          <InstallationProfileWizard
            app={selectedApp}
            onBack={
              initialApp
                ? undefined
                : () => {
                    setSelectedApp(undefined)
                    setState(INITIAL_STATE)
                  }
            }
            onUseNuon={() => setProfileSelected(true)}
          />
        ) : (
          <CreateInstallFromAppContainer
            app={selectedApp}
            onBack={
              initialApp
                ? undefined
                : () => {
                    if (profileEnabled) {
                      setProfileSelected(false)
                    } else {
                      setSelectedApp(undefined)
                    }
                    setState(INITIAL_STATE)
                  }
            }
            onStateChange={setState}
            modalId={props.modalId}
          />
        )
      ) : (
        <AppSelect
          onSelectApp={setSelectedApp}
          onClose={() => props.onClose?.()}
        />
      )}
    </Modal>
  )
}

import { useState } from 'react'
import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TComponent } from '@/types'

interface IDeployComponentModal extends Omit<IModal, 'onSubmit'> {
  component: TComponent
  currentBuildId?: string
  currentDeployStatus?: string
  installId: string
  isPending: boolean
  error?: { error?: string } | null
  onSubmit: (params: {
    buildId: string
    deployDependents: boolean
    role: string
  }) => void
  onClose: () => void
  buildSelect: (props: {
    selectedBuildId?: string
    onSelectBuild: (buildId: string) => void
    onClose: () => void
  }) => ReactNode
  roleSelector: (props: {
    value: string
    onChange: (value: string) => void
  }) => ReactNode
}

export const DeployComponentModal = ({
  component,
  currentBuildId,
  currentDeployStatus,
  installId,
  isPending,
  error,
  onSubmit,
  onClose,
  buildSelect,
  roleSelector,
  ...props
}: IDeployComponentModal) => {
  const [buildId, setBuildId] = useState<string>()
  const [deployDependents, setDeployDependents] = useState(false)
  const [selectedRole, setSelectedRole] = useState<string>('')

  const isDeployDisabled = !buildId || isPending

  const handleClose = () => {
    setBuildId(undefined)
    onClose()
  }

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
            variant="h3"
            weight="strong"
            theme="info"
          >
            <Icon variant="CloudArrowUp" size="24" />
            Deploy {component.name} component
          </Text>
          <Text
            variant="body"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            Select a build to deploy to your install
          </Text>
        </div>
      }
      size="half"
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      onClose={handleClose}
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Deploying build
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CloudArrowUp" />
            Deploy build
          </span>
        ),
        disabled: isDeployDisabled,
        onClick: () => {
          onSubmit({
            buildId: buildId!,
            deployDependents,
            role: selectedRole,
          })
        },
        variant: 'primary' as const,
      }}
      footerActions={
        <div className="flex flex-col gap-1 pl-4">
          <CheckboxInput
            checked={deployDependents}
            onChange={(e) => setDeployDependents(e.target.checked)}
            labelProps={{
              className:
                'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-1 gap-4 max-w-none',
              labelText: 'Deploy dependents',
              labelTextProps: { variant: 'base', weight: 'stronger' },
            }}
          />
          <Text variant="subtext" theme="neutral" className="ml-8 leading-none">
            Deploy all dependents as well as the selected build.
          </Text>
        </div>
      }
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to deploy component'}
          </Banner>
        ) : null}

        {buildSelect({
          selectedBuildId: buildId,
          onSelectBuild: (selectedBuildId: string) => setBuildId(selectedBuildId),
          onClose: handleClose,
        })}

        {roleSelector({
          value: selectedRole,
          onChange: setSelectedRole,
        })}
      </div>
    </Modal>
  )
}

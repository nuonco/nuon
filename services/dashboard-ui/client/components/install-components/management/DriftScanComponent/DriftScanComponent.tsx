import { useState } from 'react'
import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TComponent } from '@/types'

interface IDriftScanComponentModal extends Omit<IModal, 'onSubmit'> {
  component: TComponent
  currentBuildId?: string
  isPending: boolean
  error?: { error?: string } | null
  onSubmit: (params: { buildId: string }) => void
  onClose: () => void
  buildSelect: (props: {
    selectedBuildId?: string
    onSelectBuild: (buildId: string) => void
    onClose: () => void
  }) => ReactNode
}

export const DriftScanComponentModal = ({
  component,
  currentBuildId,
  isPending,
  error,
  onSubmit,
  onClose,
  buildSelect,
  ...props
}: IDriftScanComponentModal) => {
  const [buildId, setBuildId] = useState<string>()

  const isDriftScanDisabled = !buildId || isPending

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
            <Icon variant="Scan" size="24" />
            Drift scan {component.name} component
          </Text>
          <Text
            variant="body"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            Select a build to scan for drift
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
            Scanning build
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="MagnifyingGlass" />
            Drift scan build
          </span>
        ),
        disabled: isDriftScanDisabled,
        onClick: () => {
          onSubmit({ buildId: buildId! })
        },
        variant: 'primary' as const,
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to drift scan component'}
          </Banner>
        ) : null}

        {buildSelect({
          selectedBuildId: buildId,
          onSelectBuild: (selectedBuildId: string) => setBuildId(selectedBuildId),
          onClose: handleClose,
        })}
      </div>
    </Modal>
  )
}

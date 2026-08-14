import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IRecoverHelmReleaseModal extends Omit<IModal, 'onSubmit'> {
  componentName: string
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
  onClose: () => void
}

export const RecoverHelmReleaseModal = ({
  componentName,
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IRecoverHelmReleaseModal) => {
  const [confirm, setConfirm] = useState('')
  const canSubmit = confirm === componentName && !isPending

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text flex className="gap-4" variant="h3" weight="strong">
            <Icon variant="ArrowCounterClockwiseIcon" size="24" />
            Recover Helm release?
          </Text>
          <Text variant="body" className="text-cool-grey-600 dark:text-cool-grey-400">
            Unstick the Helm release for {componentName}
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Recovering release
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ArrowCounterClockwiseIcon" />
            Recover release
          </span>
        ),
        disabled: !canSubmit,
        onClick: onSubmit,
        variant: 'danger' as const,
      }}
      onClose={onClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to recover the Helm release'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <Text variant="body" theme="neutral">
            Helm marks a release as pending before it starts changing the cluster and clears
            that when the operation finishes. A release left pending is a rollout whose runner
            went away, and Helm refuses every further operation on it until it is recovered.
          </Text>

          <Banner theme="warn">
            <Text variant="body">
              If an earlier revision of {componentName} rolled out successfully, the release is
              rolled back to it. If none ever did, the stuck release is{' '}
              <strong>removed</strong> so the next deploy can start clean. Nothing is deployed
              either way — deploy {componentName} afterwards to roll out the version you want.
            </Text>
          </Banner>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {componentName}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-recover-component-name"
              placeholder="component name"
              type="text"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              error={confirm.length > 0 && confirm !== componentName}
              errorMessage={
                confirm.length > 0 && confirm !== componentName
                  ? "Component name doesn't match"
                  : undefined
              }
            />
          </div>
        </div>
      </div>
    </Modal>
  )
}

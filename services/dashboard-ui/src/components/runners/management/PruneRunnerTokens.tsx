'use client'

import { pruneRunnerTokens } from '@/actions/runners/prune-runner-tokens'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'

export const PruneRunnerTokensButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <PruneRunnerTokensModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Key" />}
      Prune old tokens
      {props?.isMenuButton ? <Icon variant="Key" /> : null}
    </Button>
  )
}

export const PruneRunnerTokensModal = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { runner } = useRunner()
  const { removeModal } = useSurfaces()
  const runnerId = runner?.id

  const { data, error, execute, isLoading } = useServerAction({
    action: pruneRunnerTokens,
  })

  const prunedCount = data?.invalidated_count ?? 0

  useServerActionToast({
    data: data,
    error,
    errorContent: <Text>Unable to prune runner tokens.</Text>,
    errorHeading: 'Token pruning failed',
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: (
      <Text>
        Pruned {prunedCount} old token{prunedCount !== 1 ? 's' : ''}.
      </Text>
    ),
    successHeading: 'Tokens pruned',
  })

  return (
    <Modal
      heading="Prune old runner tokens?"
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Pruning...
          </span>
        ) : (
          'Prune tokens'
        ),
        disabled: isLoading,
        onClick: () => {
          execute({
            runnerId,
            orgId: org.id,
          })
        },
        variant: 'primary' as const,
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Text variant="body" theme="neutral">
          This will remove old authentication tokens for this runner, keeping
          only the most recent token.
        </Text>
        <Text variant="body" theme="neutral">
          Use this to clean up accumulated tokens without disrupting the active
          runner.
        </Text>
      </div>
    </Modal>
  )
}

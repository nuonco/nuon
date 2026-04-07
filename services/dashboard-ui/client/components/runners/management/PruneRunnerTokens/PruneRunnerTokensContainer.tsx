import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { pruneRunnerTokens } from '@/lib'
import { PruneRunnerTokensModal } from './PruneRunnerTokens'

export const PruneRunnerTokensButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <PruneRunnerTokensModalContainer />
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

export const PruneRunnerTokensModalContainer = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { runner } = useRunner()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending: isLoading } = useMutation({
    mutationFn: () => pruneRunnerTokens({ runnerId: runner.id, orgId: org.id }),
    onSuccess: (data) => {
      const prunedCount = data?.invalidated_count ?? 0
      addToast(
        <Toast heading="Tokens pruned" theme="success">
          <Text>
            Pruned {prunedCount} old token{prunedCount !== 1 ? 's' : ''}.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Token pruning failed" theme="error">
          <Text>Unable to prune runner tokens.</Text>
        </Toast>
      )
    },
  })

  return (
    <PruneRunnerTokensModal
      isLoading={isLoading}
      onConfirm={() => mutate()}
      {...props}
    />
  )
}

import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { useAuth } from '@/hooks/use-auth'
import { adminGracefulRunnerShutdown, adminForceRunnerShutdown, adminInvalidateRunnerToken } from '@/lib'
import type { TRunner } from '@/types'
import { RunnerCard } from './RunnerCard'

interface RunnerCardContainerProps {
  runner: TRunner
  href: string
  isInstallRunner?: boolean
  onAction?: () => void
}

export const RunnerCardContainer = ({ runner, href, isInstallRunner = false, onAction }: RunnerCardContainerProps) => {
  const { addToast } = useToast()
  const { user } = useAuth()
  const adminEmail = user?.email ?? ''
  const runnerId = runner.id

  const { mutate: gracefulShutdown, isPending: isGracefulLoading } = useMutation({
    mutationFn: () => adminGracefulRunnerShutdown({ runnerId, adminEmail }),
    onSuccess: () => {
      addToast(<Toast heading="Runner shutting down" theme="success"><Text>Graceful shutdown initiated.</Text></Toast>)
      onAction?.()
    },
    onError: () => {
      addToast(<Toast heading="Shutdown failed" theme="error"><Text>Failed to initiate graceful shutdown.</Text></Toast>)
    },
  })

  const { mutate: forceShutdown, isPending: isForceLoading } = useMutation({
    mutationFn: () => adminForceRunnerShutdown({ runnerId, adminEmail }),
    onSuccess: () => {
      addToast(<Toast heading="Runner force stopped" theme="success"><Text>Force shutdown initiated.</Text></Toast>)
      onAction?.()
    },
    onError: () => {
      addToast(<Toast heading="Force shutdown failed" theme="error"><Text>Failed to force shutdown runner.</Text></Toast>)
    },
  })

  const { mutate: invalidateToken, isPending: isInvalidateLoading } = useMutation({
    mutationFn: () => adminInvalidateRunnerToken({ runnerId, adminEmail }),
    onSuccess: () => {
      addToast(<Toast heading="Token invalidated" theme="success"><Text>Runner token has been invalidated.</Text></Toast>)
      onAction?.()
    },
    onError: () => {
      addToast(<Toast heading="Token invalidation failed" theme="error"><Text>Failed to invalidate runner token.</Text></Toast>)
    },
  })

  return (
    <RunnerCard
      runner={runner}
      href={href}
      isInstallRunner={isInstallRunner}
      isGracefulLoading={isGracefulLoading}
      isForceLoading={isForceLoading}
      isInvalidateLoading={isInvalidateLoading}
      onGracefulShutdown={() => gracefulShutdown()}
      onForceShutdown={() => forceShutdown()}
      onInvalidateToken={() => invalidateToken()}
    />
  )
}

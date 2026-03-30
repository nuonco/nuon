import { useMutation } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Status } from '@/components/common/Status'
import { Link } from '@/components/common/Link'
import { Expand } from '@/components/common/Expand'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { useAuth } from '@/hooks/use-auth'
import { adminGracefulRunnerShutdown, adminForceRunnerShutdown, adminInvalidateRunnerToken } from '@/lib'
import { LoadRunnerHeartbeat } from './LoadRunnerHeartbeat'
import { LoadRunnerJob } from './LoadRunnerJob'
import type { TRunner } from '@/types'

interface RunnerCardProps {
  runner: TRunner
  href: string
  isInstallRunner?: boolean
  onAction?: () => void
}

export const RunnerCard = ({ runner, href, isInstallRunner = false, onAction }: RunnerCardProps) => {
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
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between w-full">
        <div className="flex items-center gap-2">
          <Status status={runner?.status} />
          <Text variant="base" weight="strong">
            {runner?.display_name || runner.id}
          </Text>
        </div>
        <Link href={href} className="text-sm flex items-center gap-1">
          Details <Icon variant="CaretRight" />
        </Link>
      </div>

      <Expand
        className="border rounded"
        id={runner.id}
        heading="Runner details"
      >
        <div className="px-3 flex flex-col gap-4">
          <div className="py-3 flex flex-col gap-3">
            <Text variant="base" weight="strong">Heartbeat</Text>
            <LoadRunnerHeartbeat runnerId={runner.id} />
          </div>
          <div className="py-3 flex flex-col gap-3 border-t">
            <Text variant="base" weight="strong">Last shut-down job</Text>
            <LoadRunnerJob
              runnerId={runner.id}
              groups={['operations']}
              title="Last shut-down job"
            />
          </div>
          <div className="py-3 flex flex-col gap-3 border-t">
            <Text variant="base" weight="strong">Recent job</Text>
            <LoadRunnerJob
              runnerId={runner.id}
              statuses={['finished', 'failed', 'timed-out', 'cancelled', 'not-attempted']}
              title="Recent job"
            />
          </div>
        </div>
      </Expand>

      <div className="flex flex-wrap gap-2 w-full">
        <Button
          size="sm"
          variant="secondary"
          onClick={() => invalidateToken()}
          disabled={isInvalidateLoading}
        >
          {isInvalidateLoading ? (
            <>
              <Icon variant="Loading" className="animate-spin" />
              Invalidating...
            </>
          ) : (
            <>
              <Icon variant="Key" />
              Invalidate token
            </>
          )}
        </Button>
        <Button
          size="sm"
          variant="secondary"
          onClick={() => gracefulShutdown()}
          disabled={isGracefulLoading}
        >
          {isGracefulLoading ? (
            <>
              <Icon variant="Loading" className="animate-spin" />
              Shutting down...
            </>
          ) : (
            <>
              <Icon variant="Power" />
              Graceful shutdown
            </>
          )}
        </Button>
        <Button
          size="sm"
          variant="danger"
          onClick={() => forceShutdown()}
          disabled={isForceLoading}
        >
          {isForceLoading ? (
            <>
              <Icon variant="Loading" className="animate-spin" />
              Forcing...
            </>
          ) : (
            <>
              <Icon variant="Stop" />
              Force shutdown
            </>
          )}
        </Button>
      </div>
    </div>
  )
}

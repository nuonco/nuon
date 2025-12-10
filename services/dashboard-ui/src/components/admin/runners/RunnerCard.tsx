'use client'

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Status } from '@/components/common/Status'
import { Link } from '@/components/common/Link'
import { Expand } from '@/components/common/Expand'
import { useServerAction } from '@/hooks/use-server-action'
import { gracefulOrgRunnerShutdown } from '@/actions/admin/graceful-org-runner-shutdown'
import { forceOrgRunnerShutdown } from '@/actions/admin/force-org-runner-shutdown'
import { invalidateOrgRunnerToken } from '@/actions/admin/invalidate-org-runner-token'
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
  const { execute: gracefulShutdown, isLoading: isGracefulLoading } = useServerAction({
    action: async () => {
      const result = await gracefulOrgRunnerShutdown(runner.id)
      return { data: result, error: null, headers: {}, status: 200 }
    },
  })

  const { execute: forceShutdown, isLoading: isForceLoading } = useServerAction({
    action: async () => {
      const result = await forceOrgRunnerShutdown(runner.id)
      return { data: result, error: null, headers: {}, status: 200 }
    },
  })

  const { execute: invalidateToken, isLoading: isInvalidateLoading } = useServerAction({
    action: async () => {
      const result = await invalidateOrgRunnerToken(runner.id)
      return { data: result, error: null, headers: {}, status: 200 }
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
              statuses={['finished', 'error', 'timed-out', 'cancelled', 'not-attempted']}
              title="Recent job"
            />
          </div>
        </div>
      </Expand>

      <div className="flex flex-wrap gap-2 w-full">
        <Button
          size="sm"
          variant="secondary"
          onClick={() => {
            invalidateToken()
            onAction?.()
          }}
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
          onClick={() => {
            gracefulShutdown()
            onAction?.()
          }}
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
          onClick={() => {
            forceShutdown()
            onAction?.()
          }}
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
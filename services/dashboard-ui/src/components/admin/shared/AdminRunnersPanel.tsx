'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { Card } from '@/components/common/Card'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { restartOrgRunners } from '@/actions/admin/restart-org-runners'
import { RunnerCard } from '../runners/RunnerCard'
import { LoadRunnerCard } from '../runners/LoadRunnerCard'
import type { TInstall } from '@/types'

interface AdminRunnersPanelProps extends IPanel {
  orgId: string
}

export const AdminRunnersPanel = ({
  orgId,
  size = 'half',
  ...props
}: AdminRunnersPanelProps) => {
  const { org } = useOrg()
  const [installs, setInstalls] = useState<TInstall[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const {
    execute: restartAll,
    isLoading: isRestarting,
    data,
    error: restartError,
  } = useServerAction({
    action: async () => {
      const result = await restartOrgRunners(orgId)
      return { data: result, error: null, headers: {}, status: 200 }
    },
  })

  useServerActionToast({
    data,
    error: restartError,
    successContent: <Text>All runners restarted successfully</Text>,
    successHeading: 'Runners Restarted',
    errorContent: <Text>Failed to restart runners. Please try again.</Text>,
    errorHeading: 'Restart Failed',
  })

  const fetchInstalls = async () => {
    setIsLoading(true)
    setError(undefined)

    try {
      const res = await fetch(`/api/orgs/${orgId}/installs`)
      const { data, error } = await res.json()

      if (error) {
        setError('Unable to load org installs')
      } else if (data) {
        setInstalls(data)
      }
    } catch (err) {
      setError('Unable to load org installs')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchInstalls()
  }, [orgId])

  const handleRestartAll = async () => {
    await restartAll()
    await fetchInstalls() // Refresh data after restart
  }

  return (
    <Panel
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="SlidersHorizontal" size="24" />
          <Text weight="strong" variant="h2">
            All {org?.name} Runners
          </Text>
        </div>
      }
      size={size}
      {...props}
    >
      <div className="flex flex-col gap-6 p-6">
        <div className="flex items-center justify-between">
          <Text variant="body" className="text-gray-600 dark:text-gray-300">
            Manage all runners for organization:{' '}
            <span className="font-mono">{org?.name}</span>
          </Text>
          <Button
            onClick={handleRestartAll}
            disabled={isRestarting}
            variant="secondary"
          >
            {isRestarting ? (
              <>
                <Icon variant="Loading" className="animate-spin" />
                Restarting...
              </>
            ) : (
              <>
                <Icon variant="ArrowClockwise" />
                Restart all runners
              </>
            )}
          </Button>
        </div>

        {error && (
          <div className="p-4 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
            <Text variant="subtext" className="text-red-700 dark:text-red-300">
              {error}
            </Text>
          </div>
        )}

        {/* Org Runners */}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Organization runners
          </Text>
          {org?.runner_group?.runners?.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-1">
              {org.runner_group.runners.map((runner) => (
                <RunnerCard
                  key={runner.id}
                  runner={runner}
                  href={`/${orgId}/runner`}
                  onAction={fetchInstalls}
                />
              ))}
            </div>
          ) : (
            <Card className="p-6 text-center">
              <Icon
                variant="Warning"
                size="48"
                className="text-gray-400 mb-4 mx-auto"
              />
              <Text variant="base" weight="strong" className="mb-2">
                No organization runners
              </Text>
              <Text variant="subtext">
                No runners are currently configured for this organization.
              </Text>
            </Card>
          )}
        </div>

        {/* Install Runners */}
        <div className="flex flex-col gap-4 border-t pt-6">
          <Text variant="base" weight="strong">
            Installation runners
          </Text>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="flex items-center gap-3">
                <Icon variant="Loading" className="animate-spin" />
                <Text>Loading install runners...</Text>
              </div>
            </div>
          ) : installs.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-1">
              {installs.map((install) => (
                <div key={install.id} className="rounded border p-3">
                  <div className="flex flex-col gap-3">
                    <Text variant="base" weight="strong">{install.name} runner</Text>
                    {install.runner_id ? (
                      <LoadRunnerCard
                        runnerId={install.runner_id}
                        installId={install.id}
                      />
                    ) : (
                      <Text variant="subtext">No runner assigned to this install</Text>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <Card className="p-6 text-center">
              <Icon
                variant="Warning"
                size="48"
                className="text-gray-400 mb-4 mx-auto"
              />
              <Text variant="base" weight="strong" className="mb-2">
                No installation runners
              </Text>
              <Text variant="subtext">
                No installation runners found for this organization.
              </Text>
            </Card>
          )}
        </div>
      </div>
    </Panel>
  )
}

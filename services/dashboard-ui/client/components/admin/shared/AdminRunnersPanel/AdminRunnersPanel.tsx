import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Pagination } from '@/components/common/Pagination'
import { PaginationProvider } from '@/providers/pagination-provider'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { Card } from '@/components/common/Card'
import { RunnerCard } from '../../runners/RunnerCard'
import { LoadRunnerCard } from '../../runners/LoadRunnerCard'
import type { TInstall, TRunner } from '@/types'
import type { TPaginationMeta } from '@/lib/api'

const PARAM = 'runners_offset'

interface IAdminRunnersPanel extends IPanel {
  orgId: string
  orgName: string
  orgRunners: TRunner[]
  installs: TInstall[]
  isLoading: boolean
  error?: string
  isRestarting: boolean
  onRestartAll: () => void
  onRefreshInstalls: () => void
  pagination: TPaginationMeta
}

export const AdminRunnersPanel = ({
  orgId,
  orgName,
  orgRunners,
  installs,
  isLoading,
  error,
  isRestarting,
  onRestartAll,
  onRefreshInstalls,
  pagination,
  size = 'half',
  ...props
}: IAdminRunnersPanel) => {
  return (
    <Panel
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="SlidersHorizontalIcon" size="24" />
          <Text weight="strong" variant="h2">
            All {orgName} Runners
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
            <span className="font-mono">{orgName}</span>
          </Text>
          <Button
            onClick={onRestartAll}
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
                <Icon variant="ArrowClockwiseIcon" />
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

        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Organization runners
          </Text>
          {orgRunners.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-1">
              {orgRunners.map((runner) => (
                <RunnerCard
                  key={runner.id}
                  runner={runner}
                  href={`/${orgId}/runner`}
                  onAction={onRefreshInstalls}
                />
              ))}
            </div>
          ) : (
            <Card className="p-6 text-center">
              <Icon
                variant="WarningIcon"
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
                variant="WarningIcon"
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

          <PaginationProvider>
            <Pagination
              hasNext={pagination.hasNext}
              offset={pagination.offset}
              limit={pagination.limit}
              param={PARAM}
            />
          </PaginationProvider>
        </div>
      </div>
    </Panel>
  )
}

import { Button } from '@/components/common/Button'
import type { TAppFleetInstall } from '@/types'
import { FleetInstallsTable } from './FleetInstallsTable'

export interface IFleetGroupInstalls {
  installs: TAppFleetInstall[]
  orgId?: string
  appId?: string
  labelColors?: Record<string, string>
  isLoading?: boolean
  isLoadingMore?: boolean
  onLoadMore?: () => void
}

export const FleetGroupInstalls = ({
  installs,
  orgId,
  appId,
  labelColors,
  isLoading,
  isLoadingMore,
  onLoadMore,
}: IFleetGroupInstalls) => (
  <div className="flex flex-col gap-3">
    <FleetInstallsTable
      installs={installs}
      orgId={orgId}
      appId={appId}
      labelColors={labelColors}
      showLabels={false}
      isLoading={isLoading}
    />
    {onLoadMore ? (
      <div className="flex justify-center">
        <Button onClick={onLoadMore} disabled={isLoadingMore}>
          {isLoadingMore ? 'Loading…' : 'Load more'}
        </Button>
      </div>
    ) : null}
  </div>
)

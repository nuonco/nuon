import { useEffect, useState, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import type { TCollectionView } from '@/components/common/CollectionViewToggle'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TAPIError, TInstallAlignmentResource } from '@/types'
import { cn } from '@/utils/classnames'
import {
  driftCounts,
  filterDrifted,
  hasOngoingDeployment,
  isDrifted,
  matchesSearch,
  type TDriftFilter,
} from '../alignment-utils'
import { OngoingDeployments } from './OngoingDeployments'

export interface IDriftedResources {
  error?: TAPIError | null
  initialDriftFilter?: TDriftFilter
  initiallyExpanded?: boolean
  loading?: boolean
  noun: string
  onRetry?: () => void
  pageSize?: number
  renderCard: (resource: TInstallAlignmentResource) => ReactNode
  resources: TInstallAlignmentResource[]
  search?: string
  view?: TCollectionView
}

const notComparedSummary = (resources: TInstallAlignmentResource[]) => {
  const unknown = resources.filter(
    (resource) => resource.category === 'unknown' || !resource.category
  ).length
  const disabled = resources.filter(
    (resource) => resource.category === 'disabled'
  ).length
  return [
    unknown ? `${unknown} could not be compared` : '',
    disabled ? `${disabled} disabled` : '',
  ]
    .filter(Boolean)
    .join(' · ')
}

export const DriftedResources = ({
  error,
  initialDriftFilter = 'all',
  initiallyExpanded = true,
  loading = false,
  noun,
  onRetry,
  pageSize = 4,
  renderCard,
  resources,
  search,
  view = 'list',
}: IDriftedResources) => {
  const [page, setPage] = useState(0)
  const [driftFilter, setDriftFilter] = useState(initialDriftFilter)
  const [expanded, setExpanded] = useState(initiallyExpanded)
  const imagesOnly = noun === 'images'
  const sectionTitle = imagesOnly ? 'Pending sync' : 'Drifted'

  useEffect(() => setPage(0), [search])

  if (error && !resources.length) {
    return (
      <Banner theme="error" data-testid="alignment-error">
        <div className="flex flex-wrap items-center justify-between gap-3 w-full">
          <Text variant="subtext">
            Unable to load {imagesOnly ? 'sync status' : 'drift'} for these{' '}
            {noun}. {error.description || error.error}
          </Text>
          {onRetry ? <Button onClick={onRetry}>Try again</Button> : null}
        </div>
      </Banner>
    )
  }

  if (loading && !resources.length) {
    return (
      <div
        className="flex items-center gap-2"
        role="status"
        data-testid="alignment-loading"
        aria-busy
      >
        <Loading />
        <Text variant="subtext" theme="neutral">
          Loading {imagesOnly ? 'image sync status' : 'resource alignment'}
        </Text>
      </div>
    )
  }

  const refreshError = error ? (
    <div role="alert">
      <Text variant="subtext" theme="error">
        Refreshing resources failed. Showing the last loaded results.{' '}
        {error.description || error.error}
      </Text>
    </div>
  ) : null

  const searched = resources.filter((resource) =>
    matchesSearch(resource, search)
  )
  const ongoing = searched.filter(hasOngoingDeployment)
  const drifted = searched.filter(
    (resource) => !hasOngoingDeployment(resource) && isDrifted(resource)
  )
  const counts = driftCounts(drifted)
  const visible = filterDrifted(drifted, driftFilter)
  const lastPage = Math.max(0, Math.ceil(visible.length / pageSize) - 1)
  const currentPage = Math.min(page, lastPage)
  const start = currentPage * pageSize
  const notCompared = notComparedSummary(resources.filter((r) => !isDrifted(r)))

  if (!drifted.length && !ongoing.length) {
    if (search || !resources.length || imagesOnly) return refreshError
    const allInSync = resources.every(
      (resource) => resource.category === 'in-sync'
    )
    return (
      <div className="flex flex-col gap-3">
        <div
          className="flex flex-wrap items-center gap-2"
          data-testid="alignment-summary"
        >
          <Icon
            variant={allInSync ? 'CheckCircleIcon' : 'InfoIcon'}
            size={18}
            className={
              allInSync
                ? 'text-green-800 dark:text-green-500'
                : 'text-cool-grey-600 dark:text-cool-grey-500'
            }
          />
          <Text variant="subtext">
            {allInSync
              ? `All ${noun} match the desired configuration and inputs. No changes are needed.`
              : `Compared ${noun} match the desired configuration and inputs. ${notCompared}.`}
          </Text>
        </div>
        {refreshError}
      </div>
    )
  }

  const selectDrift = (value: TDriftFilter) => {
    setDriftFilter(value)
    setPage(0)
  }

  return (
    <div className="flex flex-col gap-6">
      <OngoingDeployments
        resources={ongoing}
        renderCard={renderCard}
        view={view}
        pageSize={pageSize}
        search={search}
      />
      {drifted.length ? (
        <section aria-label={sectionTitle} data-testid="drifted-section">
          <Card className="!p-3 !gap-3">
            <Expand
              id="drifted-resources"
              className="-mx-3 !w-auto"
              interactiveHeading
              isOpen={expanded}
              onExpandedChange={setExpanded}
              headerClassName="!p-0 mx-3 !w-auto !items-center [&>button]:whitespace-nowrap [&_button:focus-visible]:outline [&_button:focus-visible]:outline-primary-400"
              toggleLabel={`${expanded ? 'Hide' : 'Show'} ${imagesOnly ? 'unsynced images' : `drifted ${noun}`}`}
              toggleContent={
                <Text theme="brand" variant="subtext">
                  {expanded ? 'Hide' : 'Show'} {noun}
                </Text>
              }
              heading={
                <div className="flex items-center gap-2 min-w-0">
                  <Icon
                    variant="WarningIcon"
                    size={16}
                    className="text-orange-800 dark:text-orange-500 shrink-0"
                  />
                  <SectionHeader
                    title={sectionTitle}
                    status={
                      <Badge size="sm" theme="warn">
                        {counts.all}
                      </Badge>
                    }
                  />
                </div>
              }
            >
              <div className="flex flex-col gap-3 mt-3 px-3">
                {!imagesOnly ? (
                  <div
                    className="overflow-x-auto p-1 [&_button:focus-visible]:[--tw-outline-style:solid] [&_button:focus-visible]:outline-2 [&_button:focus-visible]:-outline-offset-2 [&_button:focus-visible]:outline-primary-400/80"
                    onFocusCapture={(event) =>
                      event.target.scrollIntoView({
                        block: 'nearest',
                        inline: 'nearest',
                      })
                    }
                  >
                    <ToggleButton<TDriftFilter>
                      label="Filter drift"
                      size="md"
                      className="w-max"
                      value={driftFilter}
                      onChange={selectDrift}
                      options={(
                        [
                          { value: 'all', label: 'All', count: counts.all },
                          {
                            value: 'changes-not-applied',
                            label: 'Changes not applied',
                            count: counts.unapplied,
                          },
                          {
                            value: 'state-drift-detected',
                            label: 'State drift detected',
                            count: counts.stateDrift,
                          },
                        ] as const
                      ).map(({ value, label, count }) => ({
                        value,
                        label: (
                          <span className="flex items-center gap-2">
                            {label}
                            <Badge size="sm" theme="neutral">
                              {count}
                            </Badge>
                          </span>
                        ),
                      }))}
                    />
                  </div>
                ) : null}
                {visible.length ? (
                  <div
                    className="flex flex-col gap-4"
                    data-testid="drifted-cards"
                    data-view={view}
                  >
                    {(
                      ['changes-not-applied', 'state-drift-detected'] as const
                    ).map((category) => {
                      const group = visible
                        .slice(start, start + pageSize)
                        .filter((resource) => resource.category === category)
                      if (!group.length) return null
                      return (
                        <section key={category} className="flex flex-col gap-2">
                          <div className="flex flex-col gap-1">
                            {!imagesOnly ? (
                              <Text weight="strong">
                                {category === 'changes-not-applied'
                                  ? 'Changes not applied'
                                  : 'State drift detected'}
                              </Text>
                            ) : null}
                            <Text variant="subtext" theme="neutral">
                              {category === 'changes-not-applied'
                                ? imagesOnly
                                  ? 'Image versions specified by this install’s configuration haven’t been synced to its registry.'
                                  : 'Saved configuration or input changes haven’t been deployed.'
                                : 'Live resources differ from what was last applied.'}
                            </Text>
                          </div>
                          <div
                            className={cn(
                              view === 'grid'
                                ? 'flex flex-col gap-2 md:grid md:grid-cols-2 md:items-start'
                                : 'flex flex-col divide-y -mx-3'
                            )}
                          >
                            {group.map((resource) => (
                              <div
                                key={resource.id}
                                className={cn(
                                  'min-w-0',
                                  view === 'list' && 'px-3',
                                  view === 'grid' && 'md:h-full'
                                )}
                              >
                                {renderCard(resource)}
                              </div>
                            ))}
                          </div>
                        </section>
                      )
                    })}
                  </div>
                ) : (
                  <div className="flex flex-col items-start gap-3">
                    <Text variant="subtext" theme="neutral">
                      {driftFilter === 'state-drift-detected'
                        ? `No state drift detected for these ${noun}.`
                        : `No changes not applied for these ${noun}.`}
                    </Text>
                    <Button onClick={() => selectDrift('all')}>
                      View all drifted
                    </Button>
                  </div>
                )}
                {visible.length > pageSize ? (
                  <div className="flex flex-wrap gap-3 items-center justify-between">
                    <Text variant="subtext" aria-live="polite">
                      Showing {start + 1}–
                      {Math.min(start + pageSize, visible.length)} of{' '}
                      {visible.length}
                    </Text>
                    <div className="flex gap-2">
                      <Button
                        disabled={currentPage === 0}
                        onClick={() => setPage(currentPage - 1)}
                        aria-label="Previous drifted page"
                      >
                        <Icon variant="CaretLeftIcon" size={14} />
                        Previous
                      </Button>
                      <Button
                        disabled={currentPage >= lastPage}
                        onClick={() => setPage(currentPage + 1)}
                        aria-label="Next drifted page"
                      >
                        Next
                        <Icon variant="CaretRightIcon" size={14} />
                      </Button>
                    </div>
                  </div>
                ) : null}
              </div>
            </Expand>
          </Card>
        </section>
      ) : null}
      {refreshError}
    </div>
  )
}

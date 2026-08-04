import type React from 'react'
import { Divider } from '@/components/common/Divider'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Duration } from '@/components/common/Duration'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Tooltip } from '@/components/common/Tooltip'
import type {
  THealthTimelineDay,
  TInstallComponentHealthTransition,
  TInstallHealthTimelineComponent,
} from '@/types'
import { cn } from '@/utils/classnames'
import { kebabToWords, toSentenceCase } from '@/utils/string-utils'
import { formatToRelativeDay } from '@/utils/timeline-utils'

const BAR_NEUTRAL_CLASS = 'bg-cool-grey-200 dark:bg-dark-grey-700'

// Severity shading, statuspage-style: the bar encodes how much of the day was
// bad (unhealthy + degraded over observed time), not just the worst moment —
// a 2-minute blip and a 20-hour outage should not be the same red. Unknown
// time is blindness, not downtime, so it never darkens a bar; an all-unknown
// day stays neutral.
function dayBarClass(day: THealthTimelineDay): string {
  const observed = day?.observed_seconds ?? 0
  if (observed <= 0) return BAR_NEUTRAL_CLASS

  const bad = (day?.unhealthy_seconds ?? 0) + (day?.degraded_seconds ?? 0)
  const fraction = bad / observed

  if (fraction <= 0) return 'bg-green-600 dark:bg-green-500'
  if (fraction < 0.05) return 'bg-lime-500 dark:bg-lime-400'
  if (fraction < 0.25) return 'bg-yellow-500 dark:bg-yellow-400'
  if (fraction < 0.5) return 'bg-orange-500 dark:bg-orange-400'
  return 'bg-red-600 dark:bg-red-500'
}

function dayDowntimePercent(day: THealthTimelineDay): number {
  const observed = day?.observed_seconds ?? 0
  if (observed <= 0) return 0
  const bad = (day?.unhealthy_seconds ?? 0) + (day?.degraded_seconds ?? 0)
  return Math.round((bad / observed) * 1000) / 10
}

function formatHealth(health?: string): string {
  return toSentenceCase(kebabToWords(health || 'unknown'))
}

// A component nobody observed has 0% uptime arithmetically, which reads as
// total downtime. Absence of data is not downtime, so it gets a dash.
function formatUptime(uptimePercent?: number, observedSeconds?: number): string {
  if (!observedSeconds) return '—'
  return typeof uptimePercent === 'number' ? `${uptimePercent.toFixed(2)}%` : '—'
}

function DayTooltipContent({ day }: { day: THealthTimelineDay }) {
  const hasData = (day?.observed_seconds ?? 0) > 0
  const healthySeconds = Math.max(
    0,
    (day?.observed_seconds ?? 0) -
      (day?.unhealthy_seconds ?? 0) -
      (day?.degraded_seconds ?? 0) -
      (day?.unknown_seconds ?? 0)
  )

  return (
    <div className="flex flex-col gap-1 w-48">
      <div className="flex items-center justify-between gap-2">
        <Text variant="subtext" weight="strong">
          {day?.date ? formatToRelativeDay(day.date) : 'Unknown day'}
        </Text>
        <Status status={day?.health || 'unknown'} variant="badge" />
      </div>
      {hasData ? (
        <div className="flex flex-col gap-0.5">
          <Text variant="label" theme="neutral">
            {(100 - dayDowntimePercent(day)).toFixed(2)}% uptime
          </Text>
          {healthySeconds > 0 ? (
            <Text variant="label" theme="success">
              Healthy{' '}
              <Duration
                as="span"
                variant="label"
                theme="success"
                nanoseconds={healthySeconds * 1e9}
              />
            </Text>
          ) : null}
          {day?.unhealthy_seconds ? (
            <Text variant="label" theme="error">
              Unhealthy{' '}
              <Duration
                as="span"
                variant="label"
                theme="error"
                nanoseconds={day.unhealthy_seconds * 1e9}
              />
            </Text>
          ) : null}
          {day?.degraded_seconds ? (
            <Text variant="label" theme="warn">
              Degraded{' '}
              <Duration
                as="span"
                variant="label"
                theme="warn"
                nanoseconds={day.degraded_seconds * 1e9}
              />
            </Text>
          ) : null}
          {day?.unknown_seconds ? (
            <Text variant="label" theme="neutral">
              Unknown{' '}
              <Duration
                as="span"
                variant="label"
                theme="neutral"
                nanoseconds={day.unknown_seconds * 1e9}
              />
            </Text>
          ) : null}
        </div>
      ) : (
        <Text variant="label" theme="neutral">
          No data reported this day.
        </Text>
      )}
    </div>
  )
}

function HealthBar({ day }: { day: THealthTimelineDay }) {
  const hasData = (day?.observed_seconds ?? 0) > 0
  const downtime = dayDowntimePercent(day)
  const label = `${day?.date ? formatToRelativeDay(day.date) : 'Unknown day'}: ${
    hasData
      ? downtime > 0
        ? `${downtime}% downtime`
        : formatHealth(day?.health)
      : 'No data'
  }`

  return (
    <Tooltip
      position="top"
      // The tooltip wrapper span is the actual flex child of the bar row, so
      // it must grow — `grow` on the inner div alone leaves the bars at their
      // 6px base width instead of filling the container.
      className="flex grow min-w-0"
      tipContentClassName="!whitespace-normal !w-auto !p-2"
      tipContent={<DayTooltipContent day={day} />}
    >
      <div
        aria-label={label}
        tabIndex={0}
        className={cn(
          'h-6 w-1.5 shrink-0 grow rounded-[2px] focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-1 focus-visible:outline-primary-400/80',
          dayBarClass(day)
        )}
      />
    </Tooltip>
  )
}

export interface IHealthTimeline {
  className?: string
  headerAction?: React.ReactNode
  clusterAccessError?: string
  scope?: 'install' | 'component'
  days: number
  daily?: THealthTimelineDay[]
  uptimePercent?: number
  observedSeconds?: number
  currentHealth?: string
  components?: TInstallHealthTimelineComponent[]
  componentBasePath?: string
  transitions?: TInstallComponentHealthTransition[]
  deployBasePath?: string
  isLoading?: boolean
}

export const HealthTimeline = ({
  className,
  headerAction,
  clusterAccessError,
  scope = 'install',
  days,
  daily,
  uptimePercent,
  observedSeconds,
  currentHealth,
  components,
  componentBasePath,
  transitions,
  deployBasePath,
  isLoading = false,
}: IHealthTimeline) => {
  if (isLoading) {
    return (
      <div className={cn('flex flex-col gap-4', className)}>
        <div className="flex items-center justify-between gap-4">
          <Skeleton lines={2} width={['8rem', '14rem']} />
          <Skeleton width="4rem" height="1.25rem" />
        </div>
        <Skeleton height="1.5rem" />
      </div>
    )
  }

  const hasOverallData = (observedSeconds ?? 0) > 0
  const hasDaily = !!daily?.length

  return (
    <div className={cn('flex flex-col gap-4', className)}>
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <HeadingGroup>
          <Text variant="body" weight="strong">
            {days}-day health
          </Text>
          <Text variant="subtext" theme="neutral">
            {hasOverallData
              ? `${formatUptime(uptimePercent, observedSeconds)} uptime over the last ${days} days`
              : `No uptime data for the last ${days} days`}
          </Text>
        </HeadingGroup>
        {headerAction}
      </div>

      {clusterAccessError ? (
        <Banner theme="warn">
          <Text>Health cannot inspect this install's cluster: {clusterAccessError}</Text>
        </Banner>
      ) : null}

      {hasDaily ? (
        <div className="overflow-x-auto">
          <div className="flex gap-0.5 min-w-[36rem]">
            {daily!.map((day, idx) => (
              <HealthBar key={day?.date || idx} day={day} />
            ))}
          </div>
        </div>
      ) : (
        <EmptyState
          variant="history"
          size="sm"
          emptyTitle="No health data yet"
          emptyMessage="Health history will appear here once the component-health engine records observations."
        />
      )}

      {components?.length ? (
        <>
          <Divider dividerWord="Components" />
          <div className="flex flex-col gap-2">
            {components.map((component) => (
              <div
                key={component.install_component_id}
                className="flex items-center justify-between gap-3"
              >
                {/* Component routes are keyed by component_id; linking with the
                    install-component id dead-ends on an empty page. */}
                {componentBasePath && component.component_id ? (
                  <Link href={`${componentBasePath}/${component.component_id}`}>
                    {component.component_name || component.component_id}
                  </Link>
                ) : (
                  <Text>{component.component_name || component.install_component_id}</Text>
                )}
                <div className="flex items-center gap-3 shrink-0">
                  <Status
                    status={component.current_health || 'unknown'}
                    variant="badge"
                  />
                  <Text
                    variant="subtext"
                    theme="neutral"
                    className="w-16 text-right"
                  >
                    {formatUptime(
                      component.uptime_percent,
                      component.observed_seconds
                    )}
                  </Text>
                </div>
              </div>
            ))}
          </div>
        </>
      ) : null}

      {scope === 'component' ? (
        <>
          <Divider dividerWord="Transitions" />
          {transitions?.length ? (
            <Timeline<TInstallComponentHealthTransition & { created_at: string }>
              events={transitions.map((transition) => ({
                ...transition,
                created_at: transition.observed_at,
              }))}
              pagination={{ hasNext: false, offset: 0, limit: transitions.length }}
              getEventKey={(transition, idx) => `${transition.observed_at}-${idx}`}
              renderEvent={(transition) => (
                <TimelineEvent
                  status={transition.to_health || 'unknown'}
                  createdAt={transition.observed_at}
                  title={`${formatHealth(transition.from_health)} → ${formatHealth(transition.to_health)}`}
                  caption={transition.message}
                  additionalCaption={
                    [
                      transition.root_resource_kind,
                      transition.root_resource_namespace,
                      transition.root_resource_name,
                    ]
                      .filter(Boolean)
                      .join(' / ') || undefined
                  }
                  underline={transition.diagnosis}
                  actions={
                    transition.correlated_deploy_id && deployBasePath ? (
                      <Link
                        href={`${deployBasePath}/${transition.correlated_deploy_id}`}
                      >
                        View deploy
                      </Link>
                    ) : undefined
                  }
                />
              )}
            />
          ) : (
            <EmptyState
              variant="table"
              size="sm"
              emptyTitle="No transitions yet"
              emptyMessage="Health transitions will appear here once this component's health status changes."
            />
          )}
        </>
      ) : null}
    </div>
  )
}

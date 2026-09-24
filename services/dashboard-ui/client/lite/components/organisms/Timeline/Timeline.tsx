import { DateTime } from 'luxon'
import { Fragment, type HTMLAttributes, type Key, type ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { formatToRelativeDay } from '@/utils/timeline-utils'
import { Text } from '../../atoms/Text'
import { TimelineItem } from '../../molecules/TimelineItem'

export interface ITimeline<T>
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  events: T[]
  getEventKey: (event: T, index: number) => Key
  getEventTime: (event: T) => string | undefined
  group?: boolean
  loading?: boolean
  loadingItems?: number
  loadingLabel?: string
  emptyState?: ReactNode
  children: (event: T, index: number) => ReactNode
}

const UNDATED = 'undated'

type TEventGroup<T> = {
  day: string
  events: Array<{ event: T; index: number }>
}

const groupByDay = <T,>(
  events: T[],
  getEventTime: (event: T) => string | undefined
): TEventGroup<T>[] =>
  events.reduce<TEventGroup<T>[]>((groups, event, index) => {
    const time = getEventTime(event)
    const parsed = time ? DateTime.fromISO(time) : undefined
    const day = (parsed?.isValid ? parsed.toISODate() : null) ?? UNDATED
    const current = groups.at(-1)

    if (current?.day === day) current.events.push({ event, index })
    else groups.push({ day, events: [{ event, index }] })

    return groups
  }, [])

export const Timeline = <T,>({
  events,
  getEventKey,
  getEventTime,
  group = true,
  loading = false,
  loadingItems = 5,
  loadingLabel = 'Loading history',
  emptyState = 'No events yet',
  className,
  children,
  ...props
}: ITimeline<T>) => {
  const groups = group ? groupByDay(events, getEventTime) : null

  const renderEvents = (items: Array<{ event: T; index: number }>) => (
    <ol className="flex min-w-0 flex-col">
      {items.map(({ event, index }) => (
        <Fragment key={getEventKey(event, index)}>
          {children(event, index)}
        </Fragment>
      ))}
    </ol>
  )

  return (
    <div
      aria-busy={loading || undefined}
      className={cn('@container flex w-full min-w-0 flex-col gap-5', className)}
      {...props}
    >
      {loading ? (
        <>
          <span role="status" aria-label={loadingLabel} className="sr-only">
            {loadingLabel}
          </span>
          <section className="flex min-w-0 flex-col gap-2">
            {group ? <Text variant="label" loading loadingWidth={9} /> : null}
            <ol className="flex min-w-0 flex-col">
              {Array.from({ length: Math.max(1, loadingItems) }, (_, index) => (
                <TimelineItem
                  key={index}
                  data-timeline-loading-item
                  title=""
                  loading
                />
              ))}
            </ol>
          </section>
        </>
      ) : events.length === 0 ? (
        <div className="flex min-h-24 items-center justify-center px-4 py-8 text-center">
          {typeof emptyState === 'string' ? (
            <Text color="secondary">{emptyState}</Text>
          ) : (
            emptyState
          )}
        </div>
      ) : groups ? (
        groups.map(({ day, events: grouped }) => (
          <section key={day} className="flex min-w-0 flex-col gap-2">
            <Text as="h3" variant="label" color="tertiary">
              {day === UNDATED ? 'Undated' : formatToRelativeDay(day)}
            </Text>
            {renderEvents(grouped)}
          </section>
        ))
      ) : (
        renderEvents(events.map((event, index) => ({ event, index })))
      )}
    </div>
  )
}

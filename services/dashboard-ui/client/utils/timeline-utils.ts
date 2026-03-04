import { DateTime } from 'luxon'

export function formatToRelativeDay(dateString: string) {
  const inputDate = DateTime.fromISO(dateString).startOf('day')
  const today = DateTime.now().startOf('day')
  const diffDays = inputDate.diff(today, 'days').days

  if (diffDays === 0) {
    return 'Today'
  } else if (diffDays === -1) {
    return 'Yesterday'
  } else {
    return inputDate.toLocaleString(DateTime.DATE_MED_WITH_WEEKDAY)
  }
}

export interface IHasCreatedAt {
  created_at?: string
}

export type TActivityTimeline<T extends IHasCreatedAt> = Record<
  string,
  Array<T>
>

export function parseActivityTimeline<T extends IHasCreatedAt>(
  items: Array<T>
): TActivityTimeline<T> {
  return items.reduce<TActivityTimeline<T>>((acc, item) => {
    // Skip items without a valid created_at
    if (!item?.created_at) {
      console.warn('Skipping item without created_at:', item)
      return acc
    }

    const date = item.created_at.split('T')[0]

    if (!acc[date]) {
      acc[date] = []
    }
    acc[date].push(item)

    return acc
  }, {})
}

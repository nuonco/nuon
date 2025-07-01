import { DateTime } from "luxon"

export function formatToRelativeDay(isoDate: string) {
  const inputDate = DateTime.fromISO(isoDate).startOf('day')
  const today = DateTime.now().startOf('day')

  const diffDays = inputDate.diff(today, 'days').days

  if (diffDays === 0) {
    return 'Today'
  } else if (diffDays === -1) {
    return 'Yesterday'
  } else {
    return inputDate.toLocaleString(DateTime.DATETIME_MED_WITH_WEEKDAY)
  }
}

export interface IHasCreatedAt {
  created_at?: string
}

export type TActivityTimeline<T extends IHasCreatedAt> = Record<string, Array<T>>

export function parseActivityTimeline<T extends IHasCreatedAt>(
  items: Array<T>
): TActivityTimeline<T> {
  return items.reduce<TActivityTimeline<T>>((acc, item) => {
    const date = item?.created_at?.split('T')[0]

    if (!acc[date]) {
      acc[date] = []
    }
    acc[date].push(item)

    return acc
  }, {})
}

import { useMemo, type ReactNode } from 'react'
import { useSearchParams } from 'react-router'
import { Text } from '@/components/common/Text'

export const useMockPage = <T,>(
  rows: T[],
  pageSize: number,
  matches?: (row: T, query: string) => boolean
) => {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q')?.toLowerCase() ?? ''
  const offset = Number(searchParams.get('offset') ?? 0)

  return useMemo(() => {
    const filtered =
      query && matches ? rows.filter((row) => matches(row, query)) : rows
    return {
      data: filtered.slice(offset, offset + pageSize),
      pagination: {
        limit: pageSize,
        offset,
        hasNext: offset + pageSize < filtered.length,
      },
    }
  }, [rows, query, offset, pageSize, matches])
}

export const StoryFrame = ({
  children,
  note,
}: {
  children: ReactNode
  note?: string
}) => (
  <div className="flex flex-col gap-4 p-6">
    {note ? (
      <Text variant="subtext" theme="neutral">
        {note}
      </Text>
    ) : null}
    {children}
  </div>
)

export const hoursAgo = (hours: number) =>
  new Date(Date.UTC(2026, 10, 18, 12) - hours * 3600 * 1000).toISOString()

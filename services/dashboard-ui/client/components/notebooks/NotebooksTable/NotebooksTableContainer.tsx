import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getNotebooks } from '@/lib'
import { NotebooksTable, parseNotebooksToTableData } from './NotebooksTable'

const LIMIT = 20

export const NotebooksTableContainer = () => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined

  const { data: result, isLoading } = useQuery({
    queryKey: ['notebooks', org?.id, install?.id, offset, q],
    queryFn: () =>
      getNotebooks({
        orgId: org!.id,
        installId: install!.id,
        offset,
        limit: LIMIT,
        q,
      }),
    placeholderData: keepPreviousData,
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <NotebooksTable
      data={parseNotebooksToTableData(
        result?.data ?? [],
        org?.id ?? '',
        install?.id ?? ''
      )}
      isLoading={isLoading}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}

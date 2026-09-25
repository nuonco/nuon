import { useEffect, useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { SearchInput } from '@/components/common/SearchInput'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallActionsLatestRuns, getInstallRunbooks } from '@/lib'
import { ActivityPinList } from './ActivityPinList'
import type { TActivityPin } from './storage'
import { useActivityPins } from './use-activity-pins'

const LIMIT = 8

const useDebouncedValue = (value: string, delay = 300) => {
  const [debounced, setDebounced] = useState(value)

  useEffect(() => {
    const timeout = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timeout)
  }, [value, delay])

  return debounced
}

export const ActivityPinListContainer = ({
  kind,
}: {
  kind: TActivityPin['kind']
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { atLimit, isPinned, toggle } = useActivityPins()
  const [search, setSearch] = useState('')
  const [offset, setOffset] = useState(0)
  const q = useDebouncedValue(search.trim())

  useEffect(() => {
    setOffset(0)
  }, [q])

  const enabled = !!org?.id && !!install?.id
  const runbooks = useQuery({
    queryKey: ['activity-pin-picker', 'runbook', org?.id, install?.id, offset, q],
    queryFn: () =>
      getInstallRunbooks({
        orgId: org!.id,
        installId: install!.id,
        limit: LIMIT,
        offset,
        q: q || undefined,
      }),
    placeholderData: keepPreviousData,
    enabled: enabled && kind === 'runbook',
  })
  const actions = useQuery({
    queryKey: ['activity-pin-picker', 'action', org?.id, install?.id, offset, q],
    queryFn: () =>
      getInstallActionsLatestRuns({
        orgId: org!.id,
        installId: install!.id,
        limit: LIMIT,
        offset,
        q: q || undefined,
      }),
    placeholderData: keepPreviousData,
    enabled: enabled && kind === 'action',
  })
  const result = kind === 'runbook' ? runbooks : actions
  const isLoading = result.isLoading

  const items =
    kind === 'runbook'
      ? (runbooks.data?.data ?? []).flatMap((installRunbook) => {
          const id = installRunbook.runbook_id ?? installRunbook.id
          if (!id) return []
          const pin: TActivityPin = { kind: 'runbook', id }
          const checked = isPinned(pin)
          return [
            {
              id,
              name: installRunbook.runbook?.name ?? 'Runbook',
              description: installRunbook.runbook?.description,
              checked,
              disabled: atLimit && !checked,
            },
          ]
        })
      : (actions.data?.data ?? []).flatMap((installAction) => {
          const id = installAction.action_workflow_id ?? installAction.id
          const isManual = !!installAction.action_workflow?.configs?.[0]?.triggers?.some(
            (trigger) => trigger.type === 'manual'
          )
          if (!id || !isManual) return []
          const pin: TActivityPin = { kind: 'action', id }
          const checked = isPinned(pin)
          return [
            {
              id,
              name: installAction.action_workflow?.name ?? 'Action',
              checked,
              disabled: atLimit && !checked,
            },
          ]
        })

  return (
    <ActivityPinList
      title={kind === 'runbook' ? 'Runbooks' : 'Actions'}
      hint={
        kind === 'action'
          ? 'Only actions with a manual trigger are listed.'
          : undefined
      }
      items={items}
      loading={isLoading}
      filtered={!!q}
      emptyTitle={
        q
          ? `No ${kind === 'runbook' ? 'runbooks' : 'actions'} found`
          : `No ${kind === 'runbook' ? 'runbooks' : 'actions'} yet`
      }
      emptyMessage={
        q
          ? 'Try a different name or ID.'
          : kind === 'runbook'
            ? 'Runbooks show up here once they are defined on the app and synced to this install.'
            : 'Actions with a manual trigger show up here once they are on this install.'
      }
      search={
        <SearchInput
          className="w-full"
          labelClassName="w-full"
          placeholder="Search by name or ID..."
          value={search}
          onChange={(value) => setSearch(value)}
        />
      }
      pagination={{
        hasNext: result.data?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
      onOffsetChange={setOffset}
      onToggle={(id) => toggle({ kind, id })}
    />
  )
}

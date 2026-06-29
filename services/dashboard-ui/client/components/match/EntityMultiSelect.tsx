import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { getActions, getAppBranches, getComponents, getInstalls } from '@/lib'
import { cn } from '@/utils/classnames'
import { TARGET_KIND_LABELS_PLURAL, type TargetKind } from './types'

const PAGE_LIMIT = 50

export interface EntityOption {
  id: string
  name: string
}

// EntityMultiSelect renders a search-as-you-type multi-select bound to one
// of the three TargetKind taxonomies. Mirrors the slack subscribe modal's
// multi_external_select for parity — installs/components/actions all share
// the same `q` ILIKE-on-name search semantics.
//
// Installs are listed org-wide (they aren't app-owned the same way the
// other kinds are). Components and actions are listed *per app*, so an
// `appId` is required when `kind` is `components` or `actions` — the
// MatchPicker gates rendering on the user picking an app first to avoid
// org-wide enumeration.
//
// The component also batch-loads names for any preselected ids that
// aren't in the current search results so the chip area always shows
// friendly "Name (id)" labels (mirrors lookupEntityNames in the slack
// modal).
export const EntityMultiSelect = ({
  kind,
  appId,
  selectedIds,
  onChange,
  disabled,
}: {
  kind: TargetKind
  appId?: string
  selectedIds: string[]
  onChange: (next: string[]) => void
  disabled?: boolean
}) => {
  const { org } = useOrg()
  const [search, setSearch] = useState('')
  // names cache for ids that aren't in the current search results — keeps
  // the chip labels stable across re-renders.
  const [resolvedNames, setResolvedNames] = useState<Record<string, string>>(
    {}
  )

  const needsApp = kind === 'components' || kind === 'actions' || kind === 'app_branches'
  const fetchEnabled = !needsApp || !!appId

  // Search results for the current `search` query.
  const searchQuery = useQuery({
    queryKey: ['match-picker-search', kind, org.id, appId ?? '', search],
    queryFn: () => fetchEntities({ kind, orgId: org.id, appId, q: search }),
    enabled: fetchEnabled,
  })

  // Lookup pass for preselected ids that aren't in the search results.
  // Same per-app scoping rule as the search query — names for ids that
  // belong to a different app than the currently picked one stay
  // unresolved and chips fall back to bare ids (which round-trips fine
  // through the wire format).
  const missingIds = useMemo(
    () => selectedIds.filter((id) => !resolvedNames[id]),
    [selectedIds, resolvedNames]
  )
  const lookupQuery = useQuery({
    queryKey: [
      'match-picker-lookup',
      kind,
      org.id,
      appId ?? '',
      missingIds.sort().join(','),
    ],
    queryFn: () => fetchEntities({ kind, orgId: org.id, appId }),
    enabled: fetchEnabled && missingIds.length > 0,
  })

  useEffect(() => {
    const items = searchQuery.data ?? []
    if (items.length === 0) return
    setResolvedNames((prev) => {
      let changed = false
      const next = { ...prev }
      for (const it of items) {
        if (!next[it.id]) {
          next[it.id] = it.name
          changed = true
        }
      }
      return changed ? next : prev
    })
  }, [searchQuery.data])

  useEffect(() => {
    const items = lookupQuery.data ?? []
    if (items.length === 0) return
    setResolvedNames((prev) => {
      let changed = false
      const next = { ...prev }
      for (const it of items) {
        if (!next[it.id]) {
          next[it.id] = it.name
          changed = true
        }
      }
      return changed ? next : prev
    })
  }, [lookupQuery.data])

  const results = searchQuery.data ?? []
  const selectedSet = useMemo(() => new Set(selectedIds), [selectedIds])

  const toggle = (id: string) => {
    if (selectedSet.has(id)) {
      onChange(selectedIds.filter((x) => x !== id))
    } else {
      onChange([...selectedIds, id])
    }
  }

  const chips = selectedIds.map((id) => ({
    id,
    name: resolvedNames[id] ?? '',
  }))

  return (
    <div
      className={cn('flex flex-col gap-2', {
        'opacity-60 pointer-events-none': disabled,
      })}
    >
      <SearchInput
        value={search}
        onChange={setSearch}
        onClear={() => setSearch('')}
        placeholder={`Search ${TARGET_KIND_LABELS_PLURAL[kind]}…`}
        labelClassName="w-full"
        className="w-full"
      />

      {chips.length > 0 ? (
        <div className="flex flex-wrap gap-1">
          {chips.map((c) => (
            <Badge
              key={c.id}
              theme="info"
              className="flex items-center gap-1"
            >
              <span>{entityPickerLabel(c.id, c.name)}</span>
              <button
                type="button"
                aria-label={`Remove ${c.name || c.id}`}
                onClick={() => toggle(c.id)}
                className="hover:opacity-80"
              >
                <Icon variant="XIcon" size="12" />
              </button>
            </Badge>
          ))}
        </div>
      ) : null}

      <div className="border border-neutral-200 dark:border-neutral-700 rounded-md max-h-56 overflow-y-auto">
        {searchQuery.isLoading ? (
          <div className="p-3">
            <Text variant="subtext" theme="neutral">Loading…</Text>
          </div>
        ) : results.length === 0 ? (
          <div className="p-3">
            <Text variant="subtext" theme="neutral">
              {search.trim()
                ? `No ${TARGET_KIND_LABELS_PLURAL[kind]} match "${search.trim()}".`
                : needsApp
                  ? `No ${TARGET_KIND_LABELS_PLURAL[kind]} in this app.`
                  : `No ${TARGET_KIND_LABELS_PLURAL[kind]} in this org.`}
            </Text>
          </div>
        ) : (
          <ul role="listbox" aria-multiselectable="true">
            {results.map((r) => {
              const checked = selectedSet.has(r.id)
              return (
                <li key={r.id}>
                  <button
                    type="button"
                    role="option"
                    aria-selected={checked}
                    onClick={() => toggle(r.id)}
                    className={cn(
                      'flex items-center justify-between w-full text-left px-3 py-2 hover:bg-black/5 dark:hover:bg-white/5',
                      { 'bg-primary-50 dark:bg-primary-900/30': checked }
                    )}
                  >
                    <span className="truncate">
                      <Text variant="base" weight="strong">
                        {r.name || r.id}
                      </Text>
                    </span>
                    {checked ? <Icon variant="CheckIcon" size="16" /> : null}
                  </button>
                </li>
              )
            })}
          </ul>
        )}
      </div>

      {results.length === PAGE_LIMIT ? (
        <Text variant="subtext" theme="neutral">
          Showing first {PAGE_LIMIT} results — refine your search to see more.
        </Text>
      ) : null}
    </div>
  )
}

// entityPickerLabel mirrors the Slack modal helper of the same name —
// "Name (id)" when both are known, bare id otherwise.
const entityPickerLabel = (id: string, name: string): string => {
  if (!name) return id
  return `${name} (${id})`
}

// fetchEntities is the single place this component reaches for entity
// listings. Installs are org-scoped; components and actions are scoped to
// the supplied `appId` (callers gate rendering until one is picked).
const fetchEntities = async ({
  kind,
  orgId,
  appId,
  q,
}: {
  kind: TargetKind
  orgId: string
  appId?: string
  q?: string
}): Promise<EntityOption[]> => {
  switch (kind) {
    case 'installs': {
      const res = await getInstalls({ orgId, q, limit: PAGE_LIMIT })
      return (res.data ?? []).map((i) => ({
        id: i.id ?? '',
        name: i.name ?? '',
      }))
    }
    case 'components': {
      if (!appId) return []
      const res = await getComponents({ orgId, appId, q, limit: PAGE_LIMIT })
      return (res.data ?? []).map((c) => ({
        id: c.id ?? '',
        name: c.name ?? '',
      }))
    }
    case 'actions': {
      if (!appId) return []
      const res = await getActions({ orgId, appId, q, limit: PAGE_LIMIT })
      return (res.data ?? []).map((a) => ({
        id: a.id ?? '',
        name: a.name ?? '',
      }))
    }
    case 'app_branches': {
      if (!appId) return []
      const res = await getAppBranches({ orgId, appId, limit: PAGE_LIMIT })
      return (res.data ?? []).map((b) => ({
        id: b.id ?? '',
        name: b.name ?? '',
      }))
    }
  }
}

// Re-export so MatchPicker stories / consumers can render the helper.
export { entityPickerLabel }

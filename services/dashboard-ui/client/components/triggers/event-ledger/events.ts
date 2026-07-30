import type { TTriggerEventDispatchPage, TTriggerEvent } from '@/types'

export type TEventOutcome =
  | 'ok'
  | 'ignored'
  | 'rejected'
  | 'processing'
  | 'failed'

export const eventOutcome = (event: TTriggerEvent): TEventOutcome => {
  if (event?.routing_status === 'rejected') return 'rejected'
  if (event?.routing_status === 'routing_failed') return 'failed'
  if (event?.routing_status === 'ignored') return 'ignored'
  if (event?.routing_status === 'matched') return 'ok'
  return 'processing'
}

export type TEventPath = {
  path: string
  value: unknown
}

export const EVENT_PATH_MAX_DEPTH = 20
export const EVENT_PATH_MAX_COUNT = 1000
export const EVENT_PATH_TRUNCATION_MARKER = 'Additional payload paths omitted.'

const identifierPattern = /^[A-Za-z_][A-Za-z0-9_]*$/

const memberPath = (key: string) =>
  identifierPattern.test(key) ? `.${key}` : `[${JSON.stringify(key)}]`

export const flattenEventPaths = (
  value: unknown,
  maxDepth = EVENT_PATH_MAX_DEPTH,
  maxPathCount = EVENT_PATH_MAX_COUNT
): TEventPath[] => {
  const paths: TEventPath[] = []
  const pending: { current: unknown; depth: number; path: string }[] = [
    { current: value, depth: 0, path: '$' },
  ]
  let truncated = false

  while (pending.length > 0 && paths.length < maxPathCount) {
    const { current, depth, path } = pending.pop()!
    if (depth >= maxDepth && current !== null && typeof current === 'object') {
      truncated = true
      continue
    }
    if (Array.isArray(current)) {
      if (current.length === 0) paths.push({ path, value: current })
      const available = Math.max(
        0,
        maxPathCount - paths.length - pending.length
      )
      const childCount = Math.min(current.length, available)
      if (childCount < current.length) truncated = true
      for (let index = childCount - 1; index >= 0; index -= 1) {
        pending.push({
          current: current[index],
          depth: depth + 1,
          path: `${path}[${index}]`,
        })
      }
      continue
    }
    if (current !== null && typeof current === 'object') {
      const entries = Object.entries(current as Record<string, unknown>).sort(
        ([a], [b]) => a.localeCompare(b)
      )
      if (entries.length === 0) paths.push({ path, value: current })
      const available = Math.max(
        0,
        maxPathCount - paths.length - pending.length
      )
      const childCount = Math.min(entries.length, available)
      if (childCount < entries.length) truncated = true
      for (let index = childCount - 1; index >= 0; index -= 1) {
        const [key, item] = entries[index]
        pending.push({
          current: item,
          depth: depth + 1,
          path: `${path}${memberPath(key)}`,
        })
      }
      continue
    }
    paths.push({ path, value: current })
  }

  if (pending.length > 0) truncated = true
  if (truncated) {
    paths.push({ path: '…', value: EVENT_PATH_TRUNCATION_MARKER })
  }
  return paths
}

const terminalRoutingStatuses = new Set([
  'ignored',
  'matched',
  'rejected',
  'routing_failed',
])
const terminalDispatchStatuses = new Set([
  'cancelled',
  'dead_lettered',
  'triggered',
])

export const isTriggerEventTerminal = (event?: TTriggerEvent) =>
  !!event?.routing_status &&
  terminalRoutingStatuses.has(event.routing_status) &&
  !event?.dispatches_truncated &&
  (event?.dispatches ?? []).every(
    (dispatch) =>
      !!dispatch?.status && terminalDispatchStatuses.has(dispatch.status)
  )

export const shouldPollTriggerEventDispatches = ({
  event,
  pages = [],
}: {
  event?: TTriggerEvent
  pages?: TTriggerEventDispatchPage[]
}) => {
  const dispatches = pages.flatMap((page) => page?.items ?? [])
  return (
    !isTriggerEventTerminal({
      ...event,
      dispatches: [],
      dispatches_truncated: false,
    }) ||
    !!pages.at(-1)?.next_cursor ||
    dispatches.some(
      (dispatch) =>
        !dispatch?.status || !terminalDispatchStatuses.has(dispatch.status)
    ) ||
    (event?.dispatch_count ?? 0) > dispatches.length
  )
}

export const markTriggerEventDispatchPending = (
  pages: TTriggerEventDispatchPage[] = [],
  dispatchId: string
) =>
  pages.map((page) => ({
    ...page,
    items: page?.items?.map((dispatch) =>
      dispatch?.id === dispatchId
        ? { ...dispatch, status: 'pending' }
        : dispatch
    ),
  }))

export const displayEventValue = (value: unknown, limit = 160) => {
  const encoded = typeof value === 'string' ? value : JSON.stringify(value)
  const result = encoded ?? String(value)
  return result.length > limit ? `${result.slice(0, limit)}…` : result
}

export const decodeRawBody = (encoded?: string) => {
  if (!encoded) return ''
  try {
    const bytes = Uint8Array.from(atob(encoded), (character) =>
      character.charCodeAt(0)
    )
    return new TextDecoder().decode(bytes)
  } catch {
    return 'Unable to decode the stored request body.'
  }
}

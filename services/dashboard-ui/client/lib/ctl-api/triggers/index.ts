import { api } from '@/lib/api'
import type {
  TTriggerEvent,
  TTriggerEventDispatchRetryResponse,
  TTriggerEventRaw,
  TTriggerEventReplayResponse,
  TTriggerEventPage,
  TTriggerEventDispatchPage,
  TCreateTriggerBody,
  TCreateTriggerResponse,
  TTrigger,
  TTriggerIngressURLResponse,
  TTriggerRule,
  TEventTypeFacet,
  TRevealTriggerSecretResponse,
  TRevokeTriggerSecretResponse,
  TRotateTriggerIngressResponse,
  TRotateTriggerSecretResponse,
} from '@/types'

export type TTriggerEventQuery = {
  cursor?: string
  eventType?: string
  limit?: number
  order?: 'asc' | 'desc'
  outcome?: string
  query?: string
}

export const triggerEventsPath = (
  triggerId: string,
  {
    cursor,
    eventType,
    limit = 100,
    order = 'desc',
    outcome,
    query,
  }: TTriggerEventQuery
) => {
  const params = new URLSearchParams({ limit: String(limit), order })
  if (cursor) params.set('cursor', cursor)
  if (eventType) params.set('event_type', eventType)
  if (outcome) params.set('outcome', outcome)
  if (query) params.set('query', query)
  return `triggers/${encodeURIComponent(triggerId)}/events?${params.toString()}`
}

export const getTriggers = ({ orgId }: { orgId: string }) =>
  api<TTrigger[]>({ orgId, path: 'triggers' })

export const createTrigger = ({
  body,
  orgId,
}: {
  body: TCreateTriggerBody
  orgId: string
}) =>
  api<TCreateTriggerResponse>({
    body,
    method: 'POST',
    orgId,
    path: 'triggers',
  })

export const getTrigger = ({
  triggerId,
  orgId,
}: {
  triggerId: string
  orgId: string
}) =>
  api<TTrigger>({
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}`,
  })

export const getTriggerIngressURL = ({
  triggerId,
  orgId,
}: {
  triggerId: string
  orgId: string
}) =>
  api<TTriggerIngressURLResponse>({
    method: 'PATCH',
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}/ingress-url`,
  })

export const rotateTriggerIngressURL = ({
  triggerId,
  orgId,
}: {
  triggerId: string
  orgId: string
}) =>
  api<TRotateTriggerIngressResponse>({
    method: 'POST',
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}/rotate-ingress-url`,
  })

export const rotateTriggerSecret = ({
  triggerId,
  orgId,
}: {
  triggerId: string
  orgId: string
}) =>
  api<TRotateTriggerSecretResponse>({
    method: 'POST',
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}/rotate-secret`,
  })

export const revokeTriggerSecret = ({
  triggerId,
  orgId,
  secretId,
}: {
  triggerId: string
  orgId: string
  secretId: string
}) =>
  api<TRevokeTriggerSecretResponse>({
    method: 'POST',
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}/secrets/${encodeURIComponent(secretId)}/revoke`,
  })

export const revealTriggerSecret = ({
  triggerId,
  orgId,
  secretId,
}: {
  triggerId: string
  orgId: string
  secretId: string
}) =>
  api<TRevealTriggerSecretResponse>({
    method: 'PATCH',
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}/secrets/${encodeURIComponent(secretId)}/reveal`,
  })

export const getTriggerEventsForTrigger = ({
  triggerId,
  orgId,
  ...query
}: TTriggerEventQuery & { triggerId: string; orgId: string }) =>
  api<TTriggerEventPage>({
    orgId,
    path: triggerEventsPath(triggerId, query),
  })

export const getTriggerEventTypes = ({
  triggerId,
  orgId,
}: {
  triggerId: string
  orgId: string
}) =>
  api<TEventTypeFacet[]>({
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}/event-types`,
  })

export const getTriggerRules = ({
  triggerId,
  orgId,
}: {
  triggerId: string
  orgId: string
}) =>
  api<TTriggerRule[]>({
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}/rules`,
  })

export const getTriggerRule = ({
  triggerId,
  orgId,
  ruleId,
}: {
  triggerId: string
  orgId: string
  ruleId: string
}) =>
  api<TTriggerRule>({
    orgId,
    path: `triggers/${encodeURIComponent(triggerId)}/rules/${encodeURIComponent(ruleId)}`,
  })

export const getTriggerEvent = ({
  eventId,
  orgId,
}: {
  eventId: string
  orgId: string
}) =>
  api<TTriggerEvent>({
    orgId,
    path: `triggers/events/${encodeURIComponent(eventId)}`,
  })

export const getTriggerEventDispatches = ({
  cursor,
  eventId,
  limit = 100,
  orgId,
}: {
  cursor?: string
  eventId: string
  limit?: number
  orgId: string
}) =>
  api<TTriggerEventDispatchPage>({
    orgId,
    path: `triggers/dispatches?event_id=${encodeURIComponent(eventId)}&limit=${limit}${cursor ? `&cursor=${encodeURIComponent(cursor)}` : ''}`,
  })

export const getTriggerEventRaw = ({
  eventId,
  orgId,
}: {
  eventId: string
  orgId: string
}) =>
  api<TTriggerEventRaw>({
    orgId,
    path: `triggers/events/${encodeURIComponent(eventId)}/raw`,
  })

export const replayTriggerEvent = ({
  eventId,
  orgId,
}: {
  eventId: string
  orgId: string
}) =>
  api<TTriggerEventReplayResponse>({
    method: 'POST',
    orgId,
    path: `triggers/events/${encodeURIComponent(eventId)}/replay`,
  })

export const retryTriggerEventDispatch = ({
  dispatchId,
  orgId,
}: {
  dispatchId: string
  orgId: string
}) =>
  api<TTriggerEventDispatchRetryResponse>({
    method: 'POST',
    orgId,
    path: `triggers/dispatches/${encodeURIComponent(dispatchId)}/retry`,
  })

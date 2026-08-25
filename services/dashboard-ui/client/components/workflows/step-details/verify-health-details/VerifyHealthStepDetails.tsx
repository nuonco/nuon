import { DateTime } from 'luxon'
import { EmptyState } from '@/components/common/EmptyState'
import { RemovedFromAppConfigBadge } from '@/components/installs/RemovedFromAppConfig/RemovedFromAppConfig'
import { Badge } from '@/components/common/Badge'
import { Tooltip } from '@/components/common/Tooltip'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TWorkflowStep } from '@/types'

export type TVerifyHealthCheck = {
  kind?: string
  name?: string
  health?: string
  message?: string
  removed?: boolean
}

type TStatusEntry = {
  created_at_ts?: number
  status?: string
  status_human_description?: string
  metadata?: Record<string, unknown>
  history?: TStatusEntry[]
}

const HEALTH_ORDER = ['unhealthy', 'degraded', 'progressing', 'unknown', 'healthy']

function healthRank(health?: string): number {
  const idx = HEALTH_ORDER.indexOf(health || 'unknown')
  return idx === -1 ? HEALTH_ORDER.length : idx
}

function checksFrom(entry?: TStatusEntry): TVerifyHealthCheck[] {
  const checks = entry?.metadata?.checks
  if (!Array.isArray(checks)) return []
  return (checks as TVerifyHealthCheck[]).filter(
    (check) => check?.name || check?.kind
  )
}

// The latest snapshot wins: while the gate runs the current status carries the
// live checks; once it closes (success or error overwrote the description) the
// last narration in history holds the locked snapshot.
export function latestChecks(step?: TWorkflowStep): TVerifyHealthCheck[] {
  const status = step?.status as TStatusEntry | undefined
  const current = checksFrom(status)
  const history = (status?.history ?? []) as TStatusEntry[]

  let checks = current
  for (let i = history.length - 1; checks.length === 0 && i >= 0; i--) {
    checks = checksFrom(history[i])
  }

  return [...checks].sort(
    (a, b) =>
      Number(!!a.removed) - Number(!!b.removed) ||
      healthRank(a.health) - healthRank(b.health) ||
      (a.kind || '').localeCompare(b.kind || '') ||
      (a.name || '').localeCompare(b.name || '')
  )
}

export function narrationHistory(step?: TWorkflowStep): TStatusEntry[] {
  const status = step?.status as TStatusEntry | undefined
  const history = (status?.history ?? []) as TStatusEntry[]
  const all = [...history, ...(status ? [status] : [])]

  const out: TStatusEntry[] = []
  for (const entry of all) {
    const description = entry?.status_human_description
    if (!description) continue
    if (out.length > 0 && out[out.length - 1].status_human_description === description) continue
    out.push(entry)
  }
  return out
}

function summarize(checks: TVerifyHealthCheck[]): string {
  const counts = new Map<string, number>()
  let removed = 0
  for (const check of checks) {
    if (check.removed) {
      removed++
      continue
    }
    const health = check.health || 'unknown'
    counts.set(health, (counts.get(health) ?? 0) + 1)
  }
  const parts = HEALTH_ORDER.filter((health) => counts.has(health)).map(
    (health) => `${counts.get(health)} ${health}`
  )
  if (removed > 0) parts.push(`${removed} removed`)
  return parts.join(' · ')
}

export interface IVerifyHealthStepDetails {
  step?: TWorkflowStep
  orgId?: string
  componentId?: string
}

export const VerifyHealthStepDetails = ({
  step,
  orgId,
  componentId,
}: IVerifyHealthStepDetails) => {
  const checks = latestChecks(step)
  const narrations = narrationHistory(step)
  const finished = !!step?.finished

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <div className="flex items-center gap-4">
          <Text variant="base" weight="strong">
            Health checks
          </Text>
          {orgId && componentId ? (
            <Link href={`/${orgId}/installs/${step?.owner_id}/components/${componentId}`}>
              View component
            </Link>
          ) : null}
          {finished && checks.length > 0 ? (
            <Tooltip
              position="top"
              tipContent={
                <Text variant="subtext">
                  These checks are locked to what was observed when this step
                  closed — they do not update.
                </Text>
              }
            >
              <Badge size="sm" theme="neutral">
                Snapshot at close
              </Badge>
            </Tooltip>
          ) : null}
        </div>
        {checks.length > 0 ? (
          <Text variant="subtext" theme="neutral">
            {summarize(checks)}
          </Text>
        ) : null}
      </div>

      {checks.length > 0 ? (
        <div className="flex flex-col rounded-md border">
          {checks.map((check, idx) => (
            <div
              key={`${check.kind}-${check.name}-${idx}`}
              className={`flex items-center justify-between gap-3 p-2 ${idx > 0 ? 'border-t' : ''}`}
            >
              <div className="flex min-w-0 flex-col">
                <Text variant="body" weight="strong" className="truncate">
                  {check.name || check.kind}
                </Text>
                <div className="flex min-w-0 items-center gap-2">
                  <Text variant="label" theme="neutral">
                    {check.kind}
                  </Text>
                  {check.message ? (
                    <Text
                      variant="label"
                      theme="neutral"
                      className="line-clamp-1 min-w-0"
                    >
                      {check.message}
                    </Text>
                  ) : null}
                </div>
              </div>
              {check.removed ? (
                <RemovedFromAppConfigBadge kind="probe" />
              ) : (
                <Status variant="badge" status={check.health || 'unknown'} />
              )}
            </div>
          ))}
        </div>
      ) : (
        <EmptyState
          variant="history"
          size="sm"
          emptyTitle="No checks reported yet"
          emptyMessage="The check snapshot appears once the first health reading lands."
        />
      )}

      {narrations.length > 0 ? (
        <div className="flex flex-col gap-2">
          <Text variant="base" weight="strong">
            Timeline
          </Text>
          <div className="flex flex-col gap-2">
            {narrations.map((entry, idx) => (
              <div key={idx} className="flex items-start justify-between gap-3">
                <Text variant="subtext" className="min-w-0">
                  {entry.status_human_description}
                </Text>
                {entry.created_at_ts ? (
                  <Time
                    variant="label"
                    className="shrink-0"
                    time={DateTime.fromSeconds(entry.created_at_ts).toISO() ?? ''}
                    format="time-only"
                  />
                ) : null}
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  )
}

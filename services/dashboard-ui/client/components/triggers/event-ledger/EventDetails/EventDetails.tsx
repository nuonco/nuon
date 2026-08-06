import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import type {
  TTriggerEvent,
  TTriggerEventRaw,
  TTriggerEventRuleEvaluation,
  TTriggerEventWaiterMatch,
} from '@/types'
import {
  decodeRawBody,
  displayEventValue,
  eventOutcome,
  flattenEventPaths,
  type TEventPath,
} from '../events'

const IngressChecks = ({ event }: { event: TTriggerEvent }) => {
  const rejected = event?.routing_status === 'rejected'
  const reason = event?.routing_error || ''
  const envelopeFailed = rejected && reason.startsWith('envelope decoding')
  const authReasons = [
    'invalid signature',
    'invalid SNS signature',
    'invalid API key',
    'invalid basic authentication',
    'invalid bearer token',
  ]
  const authFailed =
    rejected && authReasons.some((authReason) => reason.startsWith(authReason))
  const unknownRejection = rejected && !envelopeFailed && !authFailed
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
      <LabeledValue label="Envelope decoding">
        <Status
          status={
            envelopeFailed ? 'error' : unknownRejection ? 'neutral' : 'success'
          }
          variant="badge"
        >
          {envelopeFailed ? 'Failed' : unknownRejection ? 'Unknown' : 'Decoded'}
        </Status>
      </LabeledValue>
      <LabeledValue label="Authentication">
        <Status
          status={
            authFailed
              ? 'error'
              : envelopeFailed || unknownRejection
                ? 'neutral'
                : 'success'
          }
          variant="badge"
        >
          {authFailed
            ? 'Failed'
            : envelopeFailed || unknownRejection
              ? 'Not evaluated'
              : 'Accepted'}
        </Status>
      </LabeledValue>
    </div>
  )
}

const EventPaths = ({ payload }: { payload: unknown }) => {
  const data = useMemo(() => flattenEventPaths(payload), [payload])
  const columns: ColumnDef<TEventPath>[] = useMemo(
    () => [
      {
        header: 'JSONPath',
        accessorKey: 'path',
        cell: ({ getValue }) => (
          <ClickToCopy>
            <Text family="mono" variant="subtext">
              {getValue<string>()}
            </Text>
          </ClickToCopy>
        ),
      },
      {
        header: 'Value',
        accessorFn: (path) => displayEventValue(path.value),
        cell: ({ getValue }) => (
          <Text family="mono" variant="subtext">
            {getValue<string>()}
          </Text>
        ),
      },
    ],
    []
  )
  return (
    <Table
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No payload paths',
        emptyMessage: 'This event does not contain a normalized payload.',
      }}
    />
  )
}

const RuleEvaluation = ({
  evaluation,
  href,
}: {
  evaluation: TTriggerEventRuleEvaluation
  href?: string
}) => {
  type TRuleCheck = {
    actual: string
    check: string
    expected: string
    passed: boolean
  }

  const checks: TRuleCheck[] = [
    {
      check: 'Event type',
      expected:
        evaluation?.allowed_event_types?.join(', ') || 'All event types',
      actual: evaluation?.event_type || 'Unknown',
      passed: !!evaluation?.event_type_matched,
    },
    ...(evaluation?.filters ?? []).map((filter) => ({
      check: `${filter?.from || 'payload'} ${filter?.path || '—'} ${filter?.op || '—'}`,
      expected: displayEventValue(filter?.expected),
      actual:
        filter?.error ||
        `${displayEventValue(filter?.selected?.length === 1 ? filter.selected[0] : (filter?.selected ?? []))}${filter?.truncated ? ' (truncated)' : ''}`,
      passed: !!filter?.matched,
    })),
  ]
  const columns: ColumnDef<TRuleCheck>[] = [
    {
      header: 'Check',
      accessorKey: 'check',
      cell: ({ getValue }) => (
        <Text family="mono" variant="subtext">
          {getValue<string>()}
        </Text>
      ),
    },
    {
      header: 'Expected',
      accessorKey: 'expected',
      cell: ({ getValue }) => (
        <Text family="mono" variant="subtext" className="break-all">
          {getValue<string>()}
        </Text>
      ),
    },
    {
      header: 'Actual',
      accessorKey: 'actual',
      cell: ({ getValue }) => (
        <Text family="mono" variant="subtext" className="break-all">
          {getValue<string>()}
        </Text>
      ),
    },
    {
      header: 'Result',
      accessorKey: 'passed',
      cell: ({ getValue }) => {
        const passed = getValue<boolean>()
        return (
          <Status status={passed ? 'success' : 'error'}>
            {passed ? 'Passed' : 'Failed'}
          </Status>
        )
      },
    },
  ]

  return (
    <Card className="!p-4 !gap-4">
      <div className="flex items-center justify-between gap-4">
        <div className="flex flex-col gap-1 min-w-0">
          {href ? (
            <Link href={href}>
              {evaluation?.rule_name || evaluation?.rule_id || 'Unnamed rule'}
            </Link>
          ) : (
            <Text weight="strong">
              {evaluation?.rule_name || evaluation?.rule_id || 'Unnamed rule'}
            </Text>
          )}
          <Text family="mono" variant="subtext" theme="neutral">
            {evaluation?.app_id || 'Unknown app'}
          </Text>
        </div>
        <Status
          status={evaluation?.matched ? 'success' : 'neutral'}
          variant="badge"
        >
          {evaluation?.matched ? 'Rule matched' : 'Rule did not match'}
        </Status>
      </div>
      <Table
        columns={columns}
        data={checks}
        enableSearch={false}
        enableSorting={false}
      />
    </Card>
  )
}

const WaiterMatch = ({
  match,
  orgId,
}: {
  match: TTriggerEventWaiterMatch
  orgId: string
}) => {
  const workflowHref =
    match?.install_id && match?.workflow_id
      ? `/${orgId}/installs/${match.install_id}/workflows/${match.workflow_id}${match?.workflow_step_id ? `?panel=${encodeURIComponent(match.workflow_step_id)}` : ''}`
      : undefined
  const runbookHref =
    match?.install_id && match?.runbook_id
      ? `/${orgId}/installs/${match.install_id}/runbooks/${match.runbook_id}`
      : undefined
  return (
    <Card className="!p-4 !gap-4">
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <Text weight="strong">
            {match?.notified_at
              ? 'Resumed runbook step'
              : 'Matched runbook step'}
          </Text>
          {runbookHref ? (
            <Link href={runbookHref}>
              {match?.runbook_name || match?.runbook_id || 'Runbook'}
            </Link>
          ) : (
            <Text variant="subtext">{match?.runbook_name || 'Runbook'}</Text>
          )}
          <Text variant="subtext" theme="neutral">
            {match?.workflow_step_name ||
              match?.workflow_step_id ||
              'Unknown step'}
          </Text>
        </div>
        <Status
          status={match?.notified_at ? 'success' : 'info'}
          variant="badge"
        >
          {match?.notified_at ? 'Workflow resumed' : 'Matched'}
        </Status>
      </div>
      <div className="grid gap-4 md:grid-cols-3">
        <LabeledValue label="Trigger">
          {match?.trigger_name || match?.trigger_id || 'Unknown'}
        </LabeledValue>
        <LabeledValue label="Accepted event types">
          {match?.event_types?.join(', ') || 'All event types'}
        </LabeledValue>
        <LabeledValue label="Activated">
          {match?.activated_at ? (
            <Time
              time={match.activated_at}
              format="long-datetime"
              variant="subtext"
            />
          ) : (
            '—'
          )}
        </LabeledValue>
      </div>
      {match?.filters?.length ? (
        <div className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            Filters
          </Text>
          {match.filters.map((filter, index) => (
            <CodeBlock key={`${filter?.path}-${index}`} language="text">
              {`${filter?.from || 'payload'} ${filter?.path || '—'} ${filter?.op || '—'} ${JSON.stringify(filter?.value)}`}
            </CodeBlock>
          ))}
        </div>
      ) : null}
      {workflowHref ? (
        <Link href={workflowHref}>
          View resumed workflow step <Icon variant="CaretRightIcon" />
        </Link>
      ) : null}
    </Card>
  )
}

export const EventDetails = ({
  event,
  orgId,
  hasError = false,
  hasRawError = false,
  hasDispatchError = false,
  isRawLoading,
  isRetrying = false,
  isReplaying,
  onRevealRaw,
  onLoadMoreDispatches,
  onRetry,
  onReplay,
  onRetryDispatch,
  rawRequest,
  retryingDispatchId,
  isLoadingMoreDispatches = false,
}: {
  event?: TTriggerEvent
  orgId: string
  hasError?: boolean
  hasRawError?: boolean
  hasDispatchError?: boolean
  isRawLoading: boolean
  isRetrying?: boolean
  isReplaying: boolean
  isLoadingMoreDispatches?: boolean
  onLoadMoreDispatches?: () => void
  onRevealRaw: () => void
  onRetry?: () => void
  onReplay: () => void
  onRetryDispatch: (dispatchId: string) => void
  rawRequest?: TTriggerEventRaw
  retryingDispatchId?: string
}) => {
  if (!event) {
    return (
      <div className="flex flex-col items-start gap-3">
        <Text theme="error">Event loading failed.</Text>
        <Button variant="secondary" disabled={isRetrying} onClick={onRetry}>
          <Icon variant={isRetrying ? 'Loading' : 'ArrowClockwiseIcon'} />
          {isRetrying ? 'Retrying event' : 'Retry loading event'}
        </Button>
      </div>
    )
  }

  const outcome = eventOutcome(event)
  const replayable =
    event?.routing_status === 'matched' ||
    event?.routing_status === 'ignored' ||
    event?.routing_status === 'routing_failed'
  const cannotReplay = !replayable
  const replayUnavailableReason =
    event?.routing_status === 'rejected'
      ? 'Cannot replay — authentication or envelope decoding failed.'
      : 'Replay is available after routing completes.'
  const rawBody = decodeRawBody(rawRequest?.raw_body_base64)
  const triggerActionCount =
    (event?.dispatch_count ?? 0) + (event?.waiter_match_count ?? 0)
  const replayButton = (
    <Button
      variant="primary"
      disabled={cannotReplay || isReplaying}
      onClick={onReplay}
    >
      {isReplaying ? 'Replaying event' : 'Replay event'}
    </Button>
  )

  return (
    <div className="flex flex-col gap-6">
      {hasError ? (
        <Banner theme="warn">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex flex-col gap-1">
              <Text weight="strong">Event refresh failed</Text>
              <Text variant="subtext">
                Showing the most recently loaded event details.
              </Text>
            </div>
            <Button
              variant="secondary"
              disabled={isRetrying}
              onClick={onRetry}
            >
              <Icon variant={isRetrying ? 'Loading' : 'ArrowClockwiseIcon'} />
              {isRetrying ? 'Refreshing event' : 'Refresh event'}
            </Button>
          </div>
        </Banner>
      ) : null}
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-3">
            <Text variant="h3" weight="stronger" level={1}>
              {event?.event_type || 'Event'}
            </Text>
            <Status
              status={
                outcome === 'ok'
                  ? 'success'
                  : outcome === 'rejected' || outcome === 'failed'
                    ? 'error'
                    : outcome === 'processing'
                      ? 'info'
                      : 'neutral'
              }
              variant="badge"
            >
              {outcome}
            </Status>
          </div>
          <ClickToCopy>
            <Text family="mono" variant="subtext" theme="neutral">
              {event?.id || 'Unknown event ID'}
            </Text>
          </ClickToCopy>
        </div>
        {cannotReplay ? (
          <Tooltip tipContent={replayUnavailableReason}>{replayButton}</Tooltip>
        ) : (
          replayButton
        )}
      </div>

      <Card className="!p-4 !gap-4">
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <LabeledValue label="Trigger">
            {event?.trigger_name || event?.trigger_id || 'Unknown'}
          </LabeledValue>
          <LabeledValue label="External ID" className="min-w-0">
            <ClickToCopy className="max-w-full">
              <Text family="mono" variant="subtext" className="truncate">
                {event?.external_id || '—'}
              </Text>
            </ClickToCopy>
          </LabeledValue>
          <LabeledValue label="Received">
            {event?.received_at ? (
              <Time
                time={event.received_at}
                format="long-datetime"
                variant="subtext"
              />
            ) : (
              '—'
            )}
          </LabeledValue>
          <LabeledValue label="Trigger actions">
            {String(triggerActionCount)}
          </LabeledValue>
        </div>
        <IngressChecks event={event} />
        {event?.routing_error ? (
          <Text theme="error" variant="subtext">
            {event.routing_error}
          </Text>
        ) : null}
      </Card>

      <div className="flex flex-col gap-3">
        <Text variant="base" weight="strong">
          Normalized payload
        </Text>
        <JSONViewer data={event?.payload ?? null} expanded={2} />
      </div>

      <div className="flex flex-col gap-3">
        <Text variant="base" weight="strong">
          Filterable paths
        </Text>
        <EventPaths payload={event?.payload} />
      </div>

      <div className="flex flex-col gap-3">
        <Text variant="base" weight="strong">
          Rule evaluations
        </Text>
        {event?.explanations_truncated ? (
          <Banner theme="warn">
            <Text>
              Some match explanations were omitted because they exceeded the
              storage limit.
            </Text>
          </Banner>
        ) : null}
        {(event?.match_explanations ?? []).length > 0 ? (
          (event?.match_explanations ?? []).map((evaluation, index) => (
            <RuleEvaluation
              key={evaluation?.rule_id || index}
              evaluation={evaluation}
              href={
                evaluation?.rule_id && event?.trigger_id
                  ? `/${orgId}/settings/triggers/${event.trigger_id}/rules/${evaluation.rule_id}`
                  : undefined
              }
            />
          ))
        ) : (
          <Text theme="neutral">No rules were evaluated for this event.</Text>
        )}
      </div>

      <div className="flex flex-col gap-3">
        <Text variant="base" weight="strong">
          Trigger actions
        </Text>
        {event?.dispatches_truncated ? (
          <Banner theme="info">
            <Text>
              More dispatches are available. Load them to see the complete event
              state.
            </Text>
          </Banner>
        ) : null}
        {hasDispatchError ? (
          <Banner theme="warn">
            <Text>Unable to refresh the trigger actions.</Text>
          </Banner>
        ) : null}
        {(event?.dispatches ?? []).map((dispatch) => (
          <Card key={dispatch?.id} className="!p-4 !gap-4">
            <div className="flex items-start justify-between gap-4">
              <div className="flex flex-col gap-1 min-w-0">
                <Text weight="strong">
                  {dispatch?.target_type === 'runbook'
                    ? 'Started runbook'
                    : dispatch?.target_type === 'app_branch_run'
                      ? 'Started app branch run'
                      : 'Started trigger target'}
                </Text>
                <ClickToCopy>
                  <Text family="mono" variant="subtext">
                    {dispatch?.id || 'Unknown dispatch ID'}
                  </Text>
                </ClickToCopy>
              </div>
              <div className="flex items-center gap-2">
                <Status
                  status={
                    dispatch?.status === 'triggered'
                      ? 'success'
                      : dispatch?.status === 'dead_lettered'
                        ? 'error'
                        : dispatch?.status === 'retryable_failed'
                          ? 'warn'
                          : 'info'
                  }
                  variant="badge"
                >
                  {dispatch?.status || 'unknown'}
                </Status>
                {dispatch?.status === 'dead_lettered' && dispatch?.id ? (
                  <Button
                    variant="secondary"
                    disabled={retryingDispatchId === dispatch.id}
                    onClick={() => {
                      if (dispatch?.id) onRetryDispatch(dispatch.id)
                    }}
                  >
                    {retryingDispatchId === dispatch.id
                      ? 'Retrying dispatch'
                      : 'Retry dispatch'}
                  </Button>
                ) : null}
              </div>
            </div>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
              <LabeledValue label="Rule">
                {dispatch?.trigger_rule_id || '—'}
              </LabeledValue>
              <LabeledValue label="Attempts">
                {String(dispatch?.attempts ?? 0)}
              </LabeledValue>
              <LabeledValue label="Result">
                {dispatch?.target_type === 'runbook' &&
                dispatch?.install_id &&
                dispatch?.workflow_id ? (
                  <Link
                    href={`/${orgId}/installs/${dispatch.install_id}/workflows/${dispatch.workflow_id}`}
                  >
                    {dispatch?.runbook_name || 'Runbook workflow'}
                  </Link>
                ) : dispatch?.target_type === 'app_branch_run' &&
                  dispatch?.app_id &&
                  dispatch?.target_id &&
                  dispatch?.workflow_id ? (
                  <Link
                    href={`/${orgId}/apps/${dispatch.app_id}/branches/${dispatch.target_id}/runs/${dispatch.workflow_id}`}
                  >
                    View app branch run
                  </Link>
                ) : dispatch?.result_resource_id ? (
                  `${dispatch?.result_resource_type || 'resource'} · ${dispatch.result_resource_id}`
                ) : (
                  '—'
                )}
              </LabeledValue>
            </div>
            {dispatch?.error ? (
              <Text theme="error" variant="subtext">
                {dispatch.error}
              </Text>
            ) : null}
          </Card>
        ))}
        {(event?.waiter_matches ?? []).map((match, index) => (
          <WaiterMatch key={match?.id || index} match={match} orgId={orgId} />
        ))}
        {(event?.dispatches ?? []).length === 0 &&
        (event?.waiter_matches ?? []).length === 0 ? (
          <Text theme="neutral">
            No trigger actions were taken for this event.
          </Text>
        ) : null}
        {onLoadMoreDispatches ? (
          <Button
            variant="secondary"
            disabled={isLoadingMoreDispatches}
            onClick={onLoadMoreDispatches}
          >
            {isLoadingMoreDispatches
              ? 'Loading dispatches'
              : 'Load more dispatches'}
          </Button>
        ) : null}
      </div>

      <Expand
        id="raw-event-request"
        heading="Raw request"
        headerClassName="border rounded-md"
        isIconBeforeHeading
        onExpandedChange={(isExpanded) => {
          if (isExpanded && !rawRequest && !hasRawError) onRevealRaw()
        }}
      >
        <div className="flex flex-col gap-4 pt-4">
          <div className="flex flex-col gap-2">
            <Text variant="subtext" weight="strong">
              Headers
            </Text>
            <JSONViewer data={event?.headers ?? {}} expanded={1} />
          </div>
          <div className="flex flex-col gap-2">
            <Text variant="subtext" weight="strong">
              Body
            </Text>
            {isRawLoading ? (
              <Text theme="neutral">Loading raw request...</Text>
            ) : rawRequest ? (
              <CodeBlock
                language={
                  rawRequest?.raw_content_type?.includes('json')
                    ? 'json'
                    : 'text'
                }
                showCopy
              >
                {rawBody || 'No raw request body was stored.'}
              </CodeBlock>
            ) : hasRawError ? (
              <div className="flex flex-col items-start gap-3">
                <Text theme="error">Raw request loading failed.</Text>
                <Button variant="secondary" onClick={onRevealRaw}>
                  <Icon variant="ArrowClockwiseIcon" />
                  Retry loading raw request
                </Button>
              </div>
            ) : (
              <Text theme="neutral">Raw request has not been loaded.</Text>
            )}
          </div>
        </div>
      </Expand>
    </div>
  )
}

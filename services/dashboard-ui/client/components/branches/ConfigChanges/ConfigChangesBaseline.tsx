import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

export interface IConfigChangesBaselineRun {
  href?: string
  status?: string
  message?: string
  sha?: string
}

export interface IConfigChangesBaseline {
  run?: IConfigChangesBaselineRun | null
}

const unfinishedStatuses = new Set(['pending', 'running'])

const shortSha = (sha?: string) => {
  const trimmed = sha?.trim()
  return trimmed ? trimmed.slice(0, 7) : undefined
}

export const configChangesBaseline = (
  comparison?: {
    base_run?: {
      id?: string
      status?: string
      workflow_id?: string
      vcs_connection_commit?: { sha?: string; message?: string } | null
    } | null
  } | null,
  ids?: { orgId?: string; appId?: string; branchId?: string }
): IConfigChangesBaseline | undefined => {
  if (!comparison) return undefined
  const base = comparison.base_run
  if (!base?.id) return { run: null }

  const message = base.vcs_connection_commit?.message?.split('\n')[0]?.trim()
  const sha = base.vcs_connection_commit?.sha?.trim()
  const workflowId = base.workflow_id
  const href =
    ids?.orgId && ids.appId && ids.branchId && workflowId
      ? `/${ids.orgId}/apps/${ids.appId}/branches/${ids.branchId}/runs/${workflowId}`
      : undefined

  return {
    run: {
      href,
      status: base.status || undefined,
      message: message || undefined,
      sha: sha || undefined,
    },
  }
}

export const ConfigChangesBaseline = ({
  baseline,
  hasChanges = true,
}: {
  baseline?: IConfigChangesBaseline | null
  hasChanges?: boolean
}) => {
  if (!baseline) return null
  if (!baseline.run) {
    if (!hasChanges) return null
    return (
      <Text variant="subtext" theme="neutral">
        No previous run to compare against. This shows the full config as
        created.
      </Text>
    )
  }

  const subject = baseline.run.message?.trim()
  const sha = shortSha(baseline.run.sha)
  const label = subject || sha
  const unfinished =
    !!baseline.run.status && unfinishedStatuses.has(baseline.run.status)
  const showStatus = !!baseline.run.status && baseline.run.status !== 'success'
  if (!label && !baseline.run.href && !showStatus) return null

  return (
    <span className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
      <Text variant="subtext" theme="neutral">
        Compared to{' '}
        {label ? (
          baseline.run.href ? (
            <Link
              href={baseline.run.href}
              variant="inline"
              textVariant="subtext"
              className={
                subject
                  ? 'inline-block max-w-md truncate align-bottom'
                  : 'font-mono'
              }
            >
              {label}
            </Link>
          ) : (
            <Text
              variant="subtext"
              theme="neutral"
              family={subject ? undefined : 'mono'}
            >
              {label}
            </Text>
          )
        ) : (
          'this run'
        )}
      </Text>
      {showStatus ? <Status status={baseline.run.status} /> : null}
      {unfinished ? (
        <Text variant="subtext" theme="neutral">
          This run hasn't finished.
        </Text>
      ) : null}
    </span>
  )
}

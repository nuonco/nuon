import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { getAppConfigs, getAppConfigDiff } from '@/lib'
import type { TAppConfigDiffResponse, TDiffNode } from '@/lib/ctl-api/apps/get-app-config-diff'
import type { TInstallWorkflowStep } from '@/types'
import { useState } from 'react'

function statusTheme(status?: string) {
  if (status === 'success') return 'success'
  if (status === 'error') return 'error'
  if (status === 'in-progress') return 'info'
  return 'neutral'
}

function miniStatusIcon(status?: string): string {
  if (status === 'success') return 'CheckCircleIcon'
  if (status === 'error') return 'XCircleIcon'
  if (status === 'in-progress') return 'ArrowsClockwiseIcon'
  if (status === 'skipped') return 'MinusCircleIcon'
  return 'ClockIcon'
}

function miniStatusColor(status?: string) {
  if (status === 'success') return 'text-green-500'
  if (status === 'error') return 'text-red-500'
  if (status === 'in-progress') return 'text-blue-500 animate-spin'
  return 'text-cool-grey-400'
}

function diffOpColor(op?: string) {
  if (op === 'add') return 'text-green-600 dark:text-green-400'
  if (op === 'remove') return 'text-red-600 dark:text-red-400'
  if (op === 'change') return 'text-yellow-600 dark:text-yellow-400'
  return 'text-cool-grey-500 dark:text-cool-grey-400'
}

function diffOpPrefix(op?: string) {
  if (op === 'add') return '+'
  if (op === 'remove') return '-'
  if (op === 'change') return '~'
  return ' '
}

const formatDuration = (ns?: number | null): string => {
  if (!ns) return ''
  const secs = ns / 1_000_000_000
  if (secs < 60) return `${secs.toFixed(2)}s`
  const mins = Math.floor(secs / 60)
  const rem = Math.round(secs % 60)
  return `${mins}m ${rem}s`
}

const DetailStatusIcon = ({ status }: { status?: string }) => {
  if (status === 'success') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-green-500 flex items-center justify-center shrink-0">
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M2.5 6.5L5.5 9.5L10.5 4" stroke="white" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    )
  }

  if (status === 'error') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-red-500 flex items-center justify-center shrink-0">
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M4 4L9 9M9 4L4 9" stroke="white" strokeWidth="1.8" strokeLinecap="round" />
        </svg>
      </div>
    )
  }

  if (status === 'in-progress') {
    return (
      <div className="w-[26px] h-[26px] rounded-full bg-blue-500 flex items-center justify-center shrink-0">
        <svg className="animate-spin" width="16" height="16" viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke="rgba(255,255,255,0.3)" strokeWidth="2" />
          <path d="M8 2 A6 6 0 0 1 14 8" stroke="white" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </div>
    )
  }

  return (
    <div
      className="w-[26px] h-[26px] rounded-full flex items-center justify-center shrink-0"
      style={{ boxShadow: 'inset 0 0 0 1.5px rgba(150,150,170,0.35)' }}
    >
      <div className="w-[5px] h-[5px] rounded-full bg-cool-grey-400 dark:bg-dark-grey-500" />
    </div>
  )
}

const InstallStatusIcon = ({ status }: { status?: string }) => {
  if (status === 'success' || status === 'deployed') {
    return (
      <div className="w-[17px] h-[17px] rounded-full border-2 border-green-500 flex items-center justify-center shrink-0">
        <div className="w-[5px] h-[5px] rounded-full bg-green-500" />
      </div>
    )
  }

  if (status === 'in-progress') {
    return (
      <div className="w-[17px] h-[17px] rounded-full bg-blue-500 flex items-center justify-center shrink-0">
        <svg className="animate-spin" width="11" height="11" viewBox="0 0 11 11" fill="none">
          <circle cx="5.5" cy="5.5" r="4" stroke="rgba(255,255,255,0.3)" strokeWidth="1.5" />
          <path d="M5.5 1.5 A4 4 0 0 1 9.5 5.5" stroke="white" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      </div>
    )
  }

  if (status === 'error') {
    return (
      <div className="w-[17px] h-[17px] rounded-full border-2 border-red-500 flex items-center justify-center shrink-0">
        <div className="w-[5px] h-[5px] rounded-full bg-red-500" />
      </div>
    )
  }

  return (
    <div
      className="w-[17px] h-[17px] rounded-full flex items-center justify-center shrink-0"
      style={{ boxShadow: 'inset 0 0 0 1.5px rgba(150,150,170,0.35)' }}
    >
      <div className="w-[4px] h-[4px] rounded-full bg-cool-grey-400 dark:bg-dark-grey-500" />
    </div>
  )
}

interface IWorkflowStepDetail {
  step: TInstallWorkflowStep
  onClose: () => void
}

const DiffNodeView = ({ node, depth = 0 }: { node: TDiffNode; depth?: number }) => {
  if (node.diff) {
    if (node.diff.op === 'noop' && node.diff.diff === "'' (unchanged)") return null
    if (node.diff.op === 'noop') return null
    return (
      <div className={`flex gap-2 font-mono text-xs py-0.5 ${diffOpColor(node.diff.op)}`} style={{ paddingLeft: `${depth * 16}px` }}>
        <span className="w-3 shrink-0">{diffOpPrefix(node.diff.op)}</span>
        <span className="font-medium">{node.key}:</span>
        <span className="opacity-80">{node.diff.diff}</span>
      </div>
    )
  }

  const changedChildren = (node.children || []).filter((c) => {
    if (c.diff) return c.diff.op !== 'noop'
    return (c.children || []).length > 0
  })
  if (changedChildren.length === 0) return null

  return (
    <div>
      <div className="font-mono text-xs font-medium py-0.5 text-cool-grey-700 dark:text-cool-grey-200" style={{ paddingLeft: `${depth * 16}px` }}>
        {node.key}:
      </div>
      {changedChildren.map((child, i) => (
        <DiffNodeView key={`${child.key}-${i}`} node={child} depth={depth + 1} />
      ))}
    </div>
  )
}

export const WorkflowStepDetail = ({ step, onClose }: IWorkflowStepDetail) => {
  const { org } = useOrg()
  const { app } = useApp()
  const metadata = step.status?.metadata || {}

  const isCommitStep = step.name?.toLowerCase().includes('commit')
  const isBuildStep = step.name?.toLowerCase().includes('build')
  const isConfigStep = step.name?.toLowerCase().includes('config') && !step.name?.toLowerCase().includes('diff')
  const isDeployGroupStep = step.name?.toLowerCase().includes('deploy install group')

  const commitSha = metadata.commit_sha as string | undefined
  const commitMessage = metadata.commit_message as string | undefined
  const authorName = metadata.author_name as string | undefined
  const authorEmail = metadata.author_email as string | undefined
  const repo = metadata.repo as string | undefined
  const branch = metadata.branch as string | undefined

  const builds = (metadata.builds as any[]) || []

  const appConfigId = metadata.app_config_id as string | undefined
  const componentCount = metadata.component_count as number | undefined
  const actionCount = metadata.action_count as number | undefined

  const [oldConfigId, setOldConfigId] = useState<string | undefined>(undefined)

  const { data: recentConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org!.id, appId: app!.id, limit: 10 }),
    enabled: isConfigStep && !!org?.id && !!app?.id && !!appConfigId,
  })

  const previousConfigs = (recentConfigs || []).filter((c: any) => c.id !== appConfigId)
  const selectedOldConfigId = oldConfigId ?? previousConfigs[0]?.id

  const { data: diffData } = useQuery({
    queryKey: ['app-config-diff', org?.id, app?.id, appConfigId, selectedOldConfigId],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: app!.id,
        configId: appConfigId!,
        oldConfigId: selectedOldConfigId,
      }),
    enabled: isConfigStep && !!org?.id && !!app?.id && !!appConfigId,
  })

  const githubRepoUrl = repo ? `https://github.com/${repo}` : undefined
  const githubCommitUrl = repo && commitSha ? `${githubRepoUrl}/commit/${commitSha}` : undefined

  const isInProgress = step.status?.status === 'in-progress'
  const duration = formatDuration(step.execution_time)

  const cardBorderClass = isInProgress
    ? 'border-blue-400/40 dark:border-blue-500/40'
    : 'border-cool-grey-200 dark:border-dark-grey-700'
  const cardShadow = isInProgress
    ? '0 0 0 3px rgba(63,116,224,0.08), 0 0 16px rgba(63,116,224,0.10)'
    : undefined

  const stepIndexStr = String(step.group_idx ?? '').padStart(2, '0') || '—'

  return (
    <div
      className={`rounded-xl border bg-white dark:bg-dark-grey-900 overflow-hidden transition-all ${cardBorderClass}`}
      style={cardShadow ? { boxShadow: cardShadow } : undefined}
    >
      {/* ── Header row ── */}
      <div className="flex items-center gap-3 px-5 py-4 border-b border-cool-grey-100 dark:border-dark-grey-800">
        <DetailStatusIcon status={step.status?.status} />

        <span className="font-mono text-[12px] text-cool-grey-400 dark:text-cool-grey-500 shrink-0">
          {stepIndexStr}
        </span>

        <h2 className="text-[18px] font-semibold text-cool-grey-900 dark:text-white leading-tight flex-none">
          {step.name || 'Step details'}
        </h2>

        {step.group_idx !== undefined && (
          <span className="text-[10.5px] uppercase tracking-[0.07em] font-semibold px-2 py-0.5 rounded-full border border-cool-grey-300 dark:border-dark-grey-600 text-cool-grey-500 dark:text-cool-grey-400 bg-cool-grey-50 dark:bg-dark-grey-800 shrink-0">
            Group {step.group_idx}
          </span>
        )}

        <Badge theme={statusTheme(step.status?.status)} size="sm" className="shrink-0">
          {isInProgress && (
            <svg className="animate-spin w-3 h-3 shrink-0" viewBox="0 0 12 12" fill="none">
              <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeOpacity="0.3" strokeWidth="1.5" />
              <path d="M6 1.5 A4.5 4.5 0 0 1 10.5 6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          )}
          {step.status?.status || 'pending'}
        </Badge>

        <div className="flex-1" />

        {duration && (
          <div className="flex items-center gap-1.5 text-cool-grey-400 dark:text-cool-grey-500 shrink-0">
            <Icon variant="ClockIcon" size={13} />
            <span className="font-mono text-[12px]">{duration}</span>
          </div>
        )}
      </div>

      {/* ── Sub-bar: metadata row ── */}
      <div className="flex items-start gap-6 px-5 py-3 bg-cool-grey-50 dark:bg-dark-grey-800 border-b border-cool-grey-100 dark:border-dark-grey-800 flex-wrap">
        <div className="flex flex-col gap-0.5">
          <span className="text-[10.5px] uppercase tracking-[0.06em] font-semibold text-cool-grey-400 dark:text-cool-grey-500">Step ID</span>
          <ID className="text-[12px]">{step.id}</ID>
        </div>
        {step.started_at && (
          <div className="flex flex-col gap-0.5">
            <span className="text-[10.5px] uppercase tracking-[0.06em] font-semibold text-cool-grey-400 dark:text-cool-grey-500">Started</span>
            <Time time={step.started_at} format="relative" variant="subtext" />
          </div>
        )}
        <div className="flex flex-col gap-0.5">
          <span className="text-[10.5px] uppercase tracking-[0.06em] font-semibold text-cool-grey-400 dark:text-cool-grey-500">Execution</span>
          <span className="text-[12px] text-cool-grey-700 dark:text-cool-grey-200">{step.execution_type || 'system'}</span>
        </div>
        {step.retryable !== undefined && (
          <div className="flex flex-col gap-0.5">
            <span className="text-[10.5px] uppercase tracking-[0.06em] font-semibold text-cool-grey-400 dark:text-cool-grey-500">Retryable</span>
            <Badge theme={step.retryable ? 'success' : 'neutral'} size="sm">
              {step.retryable ? 'Yes' : 'No'}
            </Badge>
          </div>
        )}
      </div>

      {/* ── Content area ── */}
      <div className="p-5 space-y-4">

        {/* ===== COMMIT STEP ===== */}
        {isCommitStep && commitSha && (
          <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg overflow-hidden">
            {repo && (
              <div className="flex items-center gap-2 px-4 py-2 bg-cool-grey-100 dark:bg-dark-grey-800 border-b border-cool-grey-200 dark:border-dark-grey-700">
                <Icon variant="GithubLogoIcon" size={16} className="text-cool-grey-600 dark:text-cool-grey-300" />
                <a
                  href={githubRepoUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs font-mono text-primary-600 dark:text-primary-400 hover:underline"
                >
                  {repo}
                </a>
                {branch && (
                  <>
                    <Icon variant="GitBranchIcon" size={14} className="text-cool-grey-400 ml-2" />
                    <Text variant="subtext" family="mono">{branch}</Text>
                  </>
                )}
              </div>
            )}
            <div className="p-4">
              <div className="flex items-start gap-3">
                <div className="w-8 h-8 rounded-full bg-cool-grey-200 dark:bg-dark-grey-700 flex items-center justify-center shrink-0">
                  <Icon variant="UserIcon" size={16} className="text-cool-grey-500 dark:text-cool-grey-400" />
                </div>
                <div className="min-w-0 flex-1">
                  <Text variant="base" weight="strong">
                    {commitMessage?.split('\n')[0] || 'No message'}
                  </Text>
                  {commitMessage?.includes('\n') && (
                    <Text variant="subtext" theme="neutral" className="mt-1 whitespace-pre-wrap">
                      {commitMessage.split('\n').slice(1).join('\n').trim()}
                    </Text>
                  )}
                  <div className="flex items-center gap-3 mt-2 flex-wrap">
                    {authorName && (
                      <Text variant="subtext" weight="strong">{authorName}</Text>
                    )}
                    {githubCommitUrl ? (
                      <a
                        href={githubCommitUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-xs font-mono px-1.5 py-0.5 rounded bg-cool-grey-100 dark:bg-dark-grey-800 text-primary-600 dark:text-primary-400 hover:underline"
                      >
                        {commitSha.substring(0, 8)}
                      </a>
                    ) : (
                      <Text variant="subtext" theme="neutral" family="mono">
                        {commitSha.substring(0, 8)}
                      </Text>
                    )}
                    {authorEmail && (
                      <Text variant="subtext" theme="neutral">
                        &lt;{authorEmail}&gt;
                      </Text>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {isCommitStep && !commitSha && (
          <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
            <Text variant="subtext" theme="neutral">
              {step.status?.status === 'in-progress' ? 'Fetching commit from VCS...' : 'Waiting to fetch commit...'}
            </Text>
          </div>
        )}

        {/* ===== CONFIG STEP ===== */}
        {isConfigStep && appConfigId && (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Text variant="label" theme="neutral">Config diff</Text>
                {componentCount !== undefined && (
                  <Badge theme="neutral" size="sm">{componentCount} components</Badge>
                )}
                {actionCount !== undefined && actionCount > 0 && (
                  <Badge theme="neutral" size="sm">{actionCount} actions</Badge>
                )}
              </div>
              {previousConfigs.length > 0 && (
                <select
                  className="text-xs bg-cool-grey-100 dark:bg-dark-grey-800 border border-cool-grey-200 dark:border-dark-grey-700 rounded px-2 py-1"
                  value={selectedOldConfigId || ''}
                  onChange={(e) => setOldConfigId(e.target.value || undefined)}
                >
                  <option value="">No previous config</option>
                  {previousConfigs.map((c: any) => (
                    <option key={c.id} value={c.id}>
                      {c.id?.substring(0, 12)} — {new Date(c.created_at).toLocaleDateString()}
                    </option>
                  ))}
                </select>
              )}
            </div>
            {diffData && (
              <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg overflow-hidden">
                <div className="flex items-center gap-3 px-4 py-2 bg-cool-grey-100 dark:bg-dark-grey-800 border-b border-cool-grey-200 dark:border-dark-grey-700">
                  {diffData.summary.has_changed ? (
                    <>
                      {diffData.summary.added > 0 && (
                        <Text variant="subtext" className="text-green-600 dark:text-green-400">+{diffData.summary.added} added</Text>
                      )}
                      {diffData.summary.removed > 0 && (
                        <Text variant="subtext" className="text-red-600 dark:text-red-400">-{diffData.summary.removed} removed</Text>
                      )}
                      {diffData.summary.changed > 0 && (
                        <Text variant="subtext" className="text-yellow-600 dark:text-yellow-400">~{diffData.summary.changed} changed</Text>
                      )}
                    </>
                  ) : (
                    <Text variant="subtext" theme="neutral">No changes</Text>
                  )}
                </div>
                {diffData.diff && (
                  <div className="p-3 max-h-96 overflow-y-auto bg-cool-grey-50 dark:bg-dark-grey-800">
                    {diffData.diff.children?.map((child, i) => (
                      <DiffNodeView key={`${child.key}-${i}`} node={child} />
                    ))}
                    {(!diffData.diff.children || diffData.diff.children.length === 0) && (
                      <Text variant="subtext" theme="neutral" family="mono">No changes detected</Text>
                    )}
                  </div>
                )}
              </div>
            )}
            {!diffData && step.status?.status !== 'pending' && (
              <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
                <Text variant="subtext" theme="neutral">Loading diff...</Text>
              </div>
            )}
          </div>
        )}

        {isConfigStep && !appConfigId && (
          <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
            <Text variant="subtext" theme="neutral">
              {step.status?.status === 'in-progress' ? 'Cloning repository and parsing configuration...' : 'Waiting to fetch app configuration...'}
            </Text>
          </div>
        )}

        {/* ===== BUILD STEP ===== */}
        {isBuildStep && builds.length > 0 && (
          <div>
            <div className="flex items-center justify-between mb-2">
              <Text variant="label" theme="neutral">Component builds</Text>
              <div className="flex items-center gap-2">
                {builds.filter((b: any) => b.status === 'success').length > 0 && (
                  <Badge theme="success" size="sm">{builds.filter((b: any) => b.status === 'success').length} passed</Badge>
                )}
                {builds.filter((b: any) => b.status === 'error').length > 0 && (
                  <Badge theme="error" size="sm">{builds.filter((b: any) => b.status === 'error').length} failed</Badge>
                )}
                {builds.filter((b: any) => b.status === 'in-progress').length > 0 && (
                  <Badge theme="info" size="sm">{builds.filter((b: any) => b.status === 'in-progress').length} running</Badge>
                )}
                {builds.filter((b: any) => b.status === 'skipped').length > 0 && (
                  <Badge theme="neutral" size="sm">{builds.filter((b: any) => b.status === 'skipped').length} skipped</Badge>
                )}
              </div>
            </div>
            <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg divide-y divide-cool-grey-200 dark:divide-dark-grey-700 overflow-hidden">
              {builds.map((build: any, i: number) => (
                <div key={build.component_id || i} className="flex items-center justify-between px-4 py-3">
                  <div className="flex items-center gap-3">
                    <Icon
                      variant={miniStatusIcon(build.status) as any}
                      size={16}
                      className={miniStatusColor(build.status)}
                    />
                    <Text variant="base">{build.component_name || build.component_id}</Text>
                  </div>
                  <Badge theme={statusTheme(build.status)} size="sm">
                    {build.status || 'pending'}
                  </Badge>
                </div>
              ))}
            </div>
          </div>
        )}

        {isBuildStep && builds.length === 0 && (
          <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
            <Text variant="subtext" theme="neutral">
              {step.status?.status === 'in-progress' ? 'Starting component builds...' : 'Waiting to start builds...'}
            </Text>
          </div>
        )}

        {/* ===== DEPLOY INSTALL GROUP STEP ===== */}
        {isDeployGroupStep && (() => {
          const groupName = step.name?.replace(/^deploy install group:\s*/i, '') || 'unknown'
          const installs = (metadata.installs as any[]) || []
          const totalInstalls = installs.length || (metadata.install_count as number) || 0
          const deployedCount = installs.filter((i: any) => i.status === 'success' || i.status === 'deployed').length
          const currentActivity = metadata.current_activity as string | undefined
          const showActivity = currentActivity || (step.status?.status === 'in-progress' && step.status?.status_human_description)
          const activityText = currentActivity || step.status?.status_human_description

          return (
            <div className="space-y-3">
              {/* Deploy head row */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Icon variant="PackageIcon" size={16} className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
                  <span className="text-[13px] text-cool-grey-600 dark:text-cool-grey-300">
                    install group:{' '}
                    <span className="font-semibold text-cool-grey-900 dark:text-white">{groupName}</span>
                  </span>
                  <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">
                    {totalInstalls} {totalInstalls === 1 ? 'install' : 'installs'}
                  </span>
                </div>
                {totalInstalls > 0 && (
                  <span className="font-mono text-[12px] text-cool-grey-500 dark:text-cool-grey-400">
                    {deployedCount} / {totalInstalls} deployed
                  </span>
                )}
              </div>

              {/* Activity bar */}
              {showActivity && activityText && (
                <div
                  className="flex items-center gap-3 px-4 py-3 rounded-[10px] border"
                  style={{
                    background: 'rgba(63,116,224,0.07)',
                    borderColor: 'rgba(63,116,224,0.32)',
                  }}
                >
                  <div className="w-[18px] h-[18px] rounded-full bg-blue-500 flex items-center justify-center shrink-0">
                    <svg className="animate-spin" width="12" height="12" viewBox="0 0 12 12" fill="none">
                      <circle cx="6" cy="6" r="4.5" stroke="rgba(255,255,255,0.3)" strokeWidth="1.5" />
                      <path d="M6 1.5 A4.5 4.5 0 0 1 10.5 6" stroke="white" strokeWidth="1.5" strokeLinecap="round" />
                    </svg>
                  </div>
                  <span className="font-mono text-[12.5px] text-blue-700 dark:text-blue-300 flex-1 truncate">
                    {activityText}
                  </span>
                  <div className="w-[120px] h-[6px] rounded-full bg-blue-100 dark:bg-blue-900/40 overflow-hidden shrink-0">
                    <div className="h-full bg-blue-500 rounded-full transition-all" style={{ width: '40%' }} />
                  </div>
                </div>
              )}

              {/* Install list */}
              {installs.length > 0 && (
                <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] divide-y divide-cool-grey-100 dark:divide-dark-grey-800 overflow-hidden">
                  {installs.map((inst: any, i: number) => {
                    const instStatus = inst.status || 'pending'
                    const isInstInProgress = instStatus === 'in-progress'
                    const isPending = instStatus === 'pending'

                    return (
                      <div
                        key={inst.install_id || i}
                        className={`px-4 py-3 transition-colors ${
                          isInstInProgress
                            ? 'bg-blue-50/60 dark:bg-[rgba(63,116,224,0.06)]'
                            : ''
                        } ${isPending ? 'opacity-[0.62]' : ''}`}
                      >
                        <div className="flex items-center justify-between gap-3">
                          <div className="flex items-center gap-2.5 min-w-0">
                            <InstallStatusIcon status={instStatus} />

                            <span className="text-[14px] font-semibold text-cool-grey-900 dark:text-white truncate">
                              {inst.install_name || inst.install_id}
                            </span>

                            {inst.region && (
                              <div className="flex items-center gap-1 shrink-0">
                                <Icon variant="GlobeIcon" size={12} className="text-cool-grey-400 dark:text-cool-grey-500" />
                                <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">{inst.region}</span>
                              </div>
                            )}

                            {inst.version && (
                              <span className="text-[11.5px] font-mono px-1.5 py-0.5 rounded-[6px] border border-cool-grey-200 dark:border-dark-grey-700 bg-cool-grey-50 dark:bg-dark-grey-800 text-cool-grey-500 dark:text-cool-grey-400 shrink-0">
                                {inst.version}
                              </span>
                            )}
                          </div>

                          <span className="font-mono text-[12.5px] text-cool-grey-400 dark:text-cool-grey-500 shrink-0">
                            {isPending ? '—' : (inst.duration || '')}
                          </span>
                        </div>

                        {isInstInProgress && (
                          <div className="flex items-center gap-3 mt-2 pl-[26px]">
                            <div className="w-[180px] h-[5px] rounded-full bg-cool-grey-200 dark:bg-dark-grey-700 overflow-hidden shrink-0">
                              <div className="h-full bg-blue-500 rounded-full transition-all" style={{ width: `${inst.progress || 30}%` }} />
                            </div>
                            {inst.activity && (
                              <span className="text-[11.5px] text-cool-grey-500 dark:text-cool-grey-400 truncate">
                                {inst.activity}
                              </span>
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}

              {installs.length === 0 && step.status?.status === 'in-progress' && !activityText && (
                <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
                  <Text variant="subtext" theme="neutral">Deploying to install group...</Text>
                </div>
              )}
            </div>
          )
        })()}

        {/* Generic fallback for other step types */}
        {!isCommitStep && !isBuildStep && !isConfigStep && !isDeployGroupStep && step.status?.status_human_description && (
          <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
            <Text variant="base">{step.status.status_human_description}</Text>
          </div>
        )}

        {/* Footer */}
        {step.install_workflow_id && (
          <div className="flex items-center gap-4 pt-3 border-t border-cool-grey-200 dark:border-dark-grey-700">
            <AdminDashboardLink path={`/workflows/${step.install_workflow_id}`} label="admin panel" />
          </div>
        )}
      </div>
    </div>
  )
}

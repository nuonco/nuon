import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toast } from '@/components/surfaces/Toast'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { useToast } from '@/hooks/use-toast'
import { getAppConfigs, getAppConfigDiff, approveWorkflowStep } from '@/lib'
import type { TDiffNode } from '@/lib/ctl-api/apps/get-app-config-diff'
import type { TInstallWorkflowStep, TAPIError } from '@/types'
import { useState } from 'react'

function statusTheme(status?: string) {
  if (status === 'success' || status === 'succeeded') return 'success'
  if (status === 'error') return 'error'
  if (status === 'in-progress') return 'info'
  return 'neutral'
}

const formatDuration = (ns?: number | null): string => {
  if (!ns) return ''
  const secs = ns / 1_000_000_000
  if (secs < 60) return `${secs.toFixed(1)}s`
  const mins = Math.floor(secs / 60)
  const rem = Math.round(secs % 60)
  return `${mins}m ${rem}s`
}

const DetailStatusIcon = ({ status }: { status?: string }) => {
  if (status === 'success' || status === 'succeeded') {
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

function getInitials(name?: string): string {
  if (!name) return '??'
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

function cacheBadgeTheme(cacheStatus?: string) {
  if (cacheStatus === 'cache hit') return 'success'
  if (cacheStatus === 'no cache') return 'warn'
  return 'neutral'
}

// ─── Diff entry markers ────────────────────────────────────────────
const DiffMarker = ({ op }: { op?: string }) => {
  if (op === 'add') {
    return (
      <div className="w-[20px] h-[20px] rounded-full bg-green-500/20 flex items-center justify-center shrink-0">
        <span className="text-[12px] font-bold text-green-600 dark:text-green-400 leading-none">+</span>
      </div>
    )
  }
  if (op === 'remove') {
    return (
      <div className="w-[20px] h-[20px] rounded-full bg-red-500/20 flex items-center justify-center shrink-0">
        <span className="text-[12px] font-bold text-red-600 dark:text-red-400 leading-none">−</span>
      </div>
    )
  }
  if (op === 'change') {
    return (
      <div className="w-[20px] h-[20px] rounded-full bg-yellow-500/20 flex items-center justify-center shrink-0">
        <span className="text-[12px] font-bold text-yellow-600 dark:text-yellow-400 leading-none">~</span>
      </div>
    )
  }
  return null
}

function diffRowBg(op?: string) {
  if (op === 'add') return 'bg-green-500/[0.06]'
  if (op === 'remove') return 'bg-red-500/[0.06]'
  return ''
}

// ─── Main Component ────────────────────────────────────────────────

export const WorkflowStepDetail = ({ step, onClose: _onClose }: IWorkflowStepDetail) => {
  const metadata = step.status?.metadata || {}

  const isCommitStep = step.name?.toLowerCase().includes('commit')
  const isBuildStep = step.name?.toLowerCase().includes('build')
  const isConfigStep = step.name?.toLowerCase().includes('config') && !step.name?.toLowerCase().includes('diff')
  const isPlanGroupStep = step.name?.toLowerCase().includes('plan install group')
  const isDeployGroupStep = step.name?.toLowerCase().includes('deploy install group')

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
        {isCommitStep && <CommitStepContent metadata={metadata} />}

        {/* ===== CONFIG STEP ===== */}
        {isConfigStep && <ConfigStepContent metadata={metadata} status={step.status?.status} />}

        {/* ===== BUILD STEP ===== */}
        {isBuildStep && <BuildStepContent metadata={metadata} status={step.status?.status} />}

        {/* ===== PLAN INSTALL GROUP STEP ===== */}
        {isPlanGroupStep && <PlanGroupStepContent step={step} metadata={metadata} />}

        {/* ===== DEPLOY INSTALL GROUP STEP ===== */}
        {isDeployGroupStep && <DeployGroupStepContent step={step} metadata={metadata} />}

        {/* Generic fallback */}
        {!isCommitStep && !isBuildStep && !isConfigStep && !isPlanGroupStep && !isDeployGroupStep && step.status?.status_human_description && (
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

// ─── COMMIT STEP ───────────────────────────────────────────────────

const CommitStepContent = ({ metadata }: { metadata: Record<string, any> }) => {
  const commitSha = metadata.commit_sha as string | undefined
  const commitMessage = metadata.commit_message as string | undefined
  const authorName = metadata.author_name as string | undefined
  const repo = metadata.repo as string | undefined
  const branch = metadata.branch as string | undefined
  const baseBranch = metadata.base_branch as string | undefined

  const prNumber = metadata.pr_number as number | undefined
  const prTitle = metadata.pr_title as string | undefined
  const prStatus = metadata.pr_status as string | undefined
  const prReviewerCount = metadata.pr_reviewer_count as number | undefined
  const prUrl = metadata.pr_url as string | undefined

  const filesChanged = metadata.files_changed as number | undefined
  const additions = metadata.additions as number | undefined
  const deletions = metadata.deletions as number | undefined
  const changedFiles = (metadata.changed_files as any[]) || []

  if (!commitSha) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
        <Text variant="subtext" theme="neutral">Fetching commit from VCS...</Text>
      </div>
    )
  }

  const messageLines = commitMessage?.split('\n') || []
  const title = messageLines[0] || 'No message'
  const body = messageLines.slice(1).join('\n').trim()

  return (
    <div className="space-y-4">
      {/* Commit message block */}
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3 min-w-0 flex-1">
          <div className="w-[32px] h-[32px] rounded-full bg-cool-grey-200 dark:bg-dark-grey-700 flex items-center justify-center shrink-0 mt-0.5">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" className="text-cool-grey-500 dark:text-cool-grey-400">
              <path d="M7 2v10M4 9l3 3 3-3" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[15px] font-semibold text-cool-grey-900 dark:text-white leading-snug">
              {title}
            </p>
            {body && (
              <p className="text-[13px] text-cool-grey-500 dark:text-cool-grey-400 mt-1 whitespace-pre-wrap leading-relaxed">
                {body}
              </p>
            )}
          </div>
        </div>
        <ID className="text-[12.5px] font-mono shrink-0">{commitSha?.substring(0, 7)}</ID>
      </div>

      {/* Author row */}
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <div className="w-[28px] h-[28px] rounded-full bg-primary-500 flex items-center justify-center shrink-0">
            <span className="text-[11px] font-semibold text-white leading-none">{getInitials(authorName)}</span>
          </div>
          <span className="text-[13px] text-cool-grey-700 dark:text-cool-grey-200">
            <span className="font-semibold text-cool-grey-900 dark:text-white">{authorName}</span>
            {' committed '}
          </span>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {branch && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full border border-cool-grey-300 dark:border-dark-grey-600 bg-cool-grey-50 dark:bg-dark-grey-800 font-mono text-[11.5px] text-cool-grey-600 dark:text-cool-grey-300">
              <svg width="12" height="12" viewBox="0 0 16 16" fill="none" className="text-cool-grey-400 dark:text-cool-grey-500">
                <path d="M5 3a2 2 0 1 0 0 4 2 2 0 0 0 0-4zm0 6a2 2 0 1 0 0 4 2 2 0 0 0 0-4zm6-6a2 2 0 1 0 0 4 2 2 0 0 0 0-4z" fill="currentColor" fillOpacity=".6" />
                <path d="M5 7v2M5 9a4 4 0 0 0 4 4h2" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
              </svg>
              {branch}
            </span>
          )}
          {baseBranch && (
            <>
              <span className="text-[12px] text-cool-grey-400">→</span>
              <span className="font-mono text-[11.5px] text-cool-grey-500 dark:text-cool-grey-400">{baseBranch}</span>
            </>
          )}
        </div>
      </div>

      {/* PR section */}
      {prNumber && prTitle && (
        <a
          href={prUrl || '#'}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-3 px-4 py-3 rounded-[10px] border border-cool-grey-200 dark:border-dark-grey-700 bg-cool-grey-50/50 dark:bg-dark-grey-800/50 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-800 transition-colors"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0">
            <path d="M10 3a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM6 9a2 2 0 1 0 0 4 2 2 0 0 0 0-4z" fill="currentColor" fillOpacity=".5" />
            <path d="M10 7v2a4 4 0 0 1-4 4M10 7V3" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
          </svg>
          <span className="text-[13px] font-semibold text-cool-grey-600 dark:text-cool-grey-300">#{prNumber}</span>
          <span className="text-[13px] text-cool-grey-700 dark:text-cool-grey-200 flex-1 truncate">{prTitle}</span>
          {prStatus && (
            <Badge theme={prStatus === 'open' ? 'success' : 'neutral'} size="sm">{prStatus}</Badge>
          )}
          {(prReviewerCount ?? 0) > 0 && (
            <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">{prReviewerCount} reviewers</span>
          )}
          <Icon variant="ArrowSquareOutIcon" size={14} className="text-cool-grey-400 shrink-0" />
        </a>
      )}

      {/* File diff summary */}
      {filesChanged !== undefined && (
        <div className="space-y-2">
          <div className="flex items-center gap-3 flex-wrap">
            <span className="text-[13px] text-cool-grey-700 dark:text-cool-grey-200">
              <span className="font-semibold">{filesChanged}</span> files changed
            </span>
            {(additions ?? 0) > 0 && (
              <span className="text-[13px] font-semibold text-green-600 dark:text-green-400">+{additions?.toLocaleString()}</span>
            )}
            {(deletions ?? 0) > 0 && (
              <span className="text-[13px] font-semibold text-red-600 dark:text-red-400">−{deletions?.toLocaleString()}</span>
            )}
            {/* Heat bar */}
            {(additions ?? 0) + (deletions ?? 0) > 0 && (
              <div className="flex gap-[2px] ml-1">
                {Array.from({ length: Math.min(Math.round(((additions ?? 0) / ((additions ?? 0) + (deletions ?? 0))) * 20), 20) }).map((_, i) => (
                  <div key={`a${i}`} className="w-[8px] h-[8px] rounded-[2px] bg-green-500" />
                ))}
                {Array.from({ length: Math.min(Math.round(((deletions ?? 0) / ((additions ?? 0) + (deletions ?? 0))) * 20), 20) }).map((_, i) => (
                  <div key={`d${i}`} className="w-[8px] h-[8px] rounded-[2px] bg-red-500" />
                ))}
              </div>
            )}
          </div>

          {/* Changed files list */}
          {changedFiles.length > 0 && (
            <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] divide-y divide-cool-grey-100 dark:divide-dark-grey-800 overflow-hidden">
              {changedFiles.map((file: any, i: number) => (
                <div key={file?.path || i} className="flex items-center justify-between px-4 py-2.5">
                  <div className="flex items-center gap-2 min-w-0">
                    <Icon variant="FileTextIcon" size={14} className="text-cool-grey-400 dark:text-cool-grey-500 shrink-0" />
                    <span className="font-mono text-[12.5px] text-cool-grey-700 dark:text-cool-grey-200 truncate">{file?.path}</span>
                  </div>
                  <div className="flex items-center gap-2 shrink-0 ml-3">
                    {(file?.additions ?? 0) > 0 && (
                      <span className="text-[12px] font-semibold text-green-600 dark:text-green-400">+{file?.additions}</span>
                    )}
                    {(file?.deletions ?? 0) > 0 && (
                      <span className="text-[12px] font-semibold text-red-600 dark:text-red-400">−{file?.deletions}</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ─── CONFIG STEP ───────────────────────────────────────────────────

// Sections we want to display from the diff tree, in order
const DIFF_SECTION_KEYS: Record<string, string> = {
  components: 'Components',
  actions: 'Actions',
  inputs: 'Install inputs',
  secrets: 'Secrets',
  sandbox: 'Sandbox',
  runner: 'Runner',
  permissions: 'Permissions',
  stack: 'Stack',
}

type DiffSectionData = {
  name: string
  additions: number
  removals: number
  changed: number
  entries: { op: string; name: string; description: string }[]
}

function extractSections(node?: TDiffNode): DiffSectionData[] {
  if (!node?.children) return []

  const sections: DiffSectionData[] = []
  for (const child of node.children) {
    const displayName = DIFF_SECTION_KEYS[child.key]
    if (!displayName) continue

    const section: DiffSectionData = { name: displayName, additions: 0, removals: 0, changed: 0, entries: [] }
    collectDiffEntries(child, '', section)
    if (section.entries.length > 0) {
      sections.push(section)
    }
  }
  return sections
}

function collectDiffEntries(node: TDiffNode, parentKey: string, section: DiffSectionData) {
  if (node.diff && node.diff.op !== 'noop' && node.diff.op !== '') {
    const entry = {
      op: node.diff.op,
      name: parentKey || node.key,
      description: parentKey ? node.diff.diff : node.diff.diff,
    }
    if (node.diff.op === 'add') section.additions++
    else if (node.diff.op === 'remove') section.removals++
    else if (node.diff.op === 'change') section.changed++
    section.entries.push(entry)
    return
  }

  if (node.children) {
    const hasLeaves = node.children.some((c) => c.diff && c.diff.op !== 'noop' && c.diff.op !== '')
    if (hasLeaves) {
      for (const c of node.children) {
        if (c.diff && c.diff.op !== 'noop' && c.diff.op !== '') {
          const entry = { op: c.diff.op, name: node.key, description: c.diff.diff }
          if (c.diff.op === 'add') section.additions++
          else if (c.diff.op === 'remove') section.removals++
          else if (c.diff.op === 'change') section.changed++
          section.entries.push(entry)
        }
      }
    } else {
      for (const c of node.children) {
        collectDiffEntries(c, node.key || parentKey, section)
      }
    }
  }
}

function computeSummary(sections: DiffSectionData[]) {
  let added = 0, removed = 0, changed = 0
  for (const s of sections) {
    added += s.additions
    removed += s.removals
    changed += s.changed
  }
  return { added, removed, changed }
}

const ConfigStepContent = ({ metadata, status }: { metadata: Record<string, any>; status?: string }) => {
  const { org } = useOrg()
  const { app } = useApp()
  const appConfigId = metadata.app_config_id as string | undefined

  // Fetch recent configs to find the previous one for diffing
  const { data: recentConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org!.id, appId: app!.id, limit: 10 }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const previousConfigs = (recentConfigs || []).filter((c: any) => c.id !== appConfigId)
  const oldConfigId = previousConfigs[0]?.id

  // Fetch the diff from the API
  const { data: diffData, isError: diffError } = useQuery({
    queryKey: ['app-config-diff', org?.id, app?.id, appConfigId, oldConfigId],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: app!.id,
        configId: appConfigId!,
        oldConfigId,
      }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
    retry: 1,
  })

  if (!appConfigId) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
        <Text variant="subtext" theme="neutral">
          {status === 'in-progress' ? 'Cloning repository and parsing configuration...' : 'Waiting to fetch app configuration...'}
        </Text>
      </div>
    )
  }

  // Extract sections from the API diff tree
  const sections = diffData?.diff ? extractSections(diffData.diff) : []
  const summary = sections.length > 0 ? computeSummary(sections) : (diffData?.summary || null)

  return (
    <div className="space-y-4">
      {/* Header: config file + summary counts */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="flex items-center gap-3">
          <span className="font-mono text-[12.5px] px-2.5 py-1 rounded-[6px] border border-cool-grey-300 dark:border-dark-grey-600 bg-cool-grey-50 dark:bg-dark-grey-800 text-cool-grey-700 dark:text-cool-grey-200">
            nuon.toml
          </span>
          {summary && (
            <>
              {(summary.added ?? 0) > 0 && (
                <span className="text-[13px] text-green-600 dark:text-green-400">
                  <span className="font-semibold">{summary.added}</span> additions
                </span>
              )}
              {(summary.removed ?? 0) > 0 && (
                <span className="text-[13px] text-red-600 dark:text-red-400">
                  <span className="font-semibold">{summary.removed}</span> removals
                </span>
              )}
              {(summary.changed ?? 0) > 0 && (
                <span className="text-[13px] text-yellow-600 dark:text-yellow-400">
                  <span className="font-semibold">{summary.changed}</span> changed
                </span>
              )}
            </>
          )}
        </div>
      </div>

      {/* Diff sections from API */}
      {sections.map((section, i) => (
        <ConfigDiffSectionView key={section.name || i} section={section} />
      ))}

      {/* Loading / error / no-changes fallback */}
      {sections.length === 0 && appConfigId && (
        <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
          <Text variant="subtext" theme="neutral">
            {diffError || diffData
              ? metadata.component_count !== undefined
                ? `Synced ${metadata.component_count} components${metadata.action_count ? `, ${metadata.action_count} actions` : ''}`
                : 'No changes detected'
              : 'Loading diff...'}
          </Text>
        </div>
      )}
    </div>
  )
}

const ConfigDiffSectionView = ({ section }: { section: DiffSectionData }) => {
  const [expanded, setExpanded] = useState(true)

  return (
    <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] overflow-hidden">
      {/* Section header */}
      <button
        className="flex items-center justify-between w-full px-4 py-2.5 bg-cool-grey-100/70 dark:bg-dark-grey-800 border-b border-cool-grey-200 dark:border-dark-grey-700 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-750 transition-colors text-left"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-2">
          <svg
            width="12" height="12" viewBox="0 0 12 12" fill="none"
            className={`text-cool-grey-400 shrink-0 transition-transform ${expanded ? 'rotate-90' : ''}`}
          >
            <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          <span className="text-[13px] font-semibold text-cool-grey-900 dark:text-white">{section.name}</span>
        </div>
        <div className="flex items-center gap-2">
          {section.additions > 0 && (
            <span className="text-[12px] font-semibold text-green-600 dark:text-green-400">+{section.additions}</span>
          )}
          {section.removals > 0 && (
            <span className="text-[12px] font-semibold text-red-600 dark:text-red-400">−{section.removals}</span>
          )}
          {section.changed > 0 && (
            <span className="text-[12px] font-semibold text-yellow-600 dark:text-yellow-400">~{section.changed}</span>
          )}
        </div>
      </button>
      {/* Entries */}
      {expanded && (
        <div className="divide-y divide-cool-grey-100 dark:divide-dark-grey-800">
          {section.entries.map((entry, j) => (
            <div key={`${entry.name}-${j}`} className={`flex items-center gap-3 px-4 py-2.5 ${diffRowBg(entry.op)}`}>
              <DiffMarker op={entry.op} />
              <span className="font-mono text-[12.5px] font-semibold text-cool-grey-900 dark:text-white">{entry.name}</span>
              <span className="font-mono text-[12px] text-cool-grey-500 dark:text-cool-grey-400 truncate">{entry.description}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// ─── BUILD STEP ────────────────────────────────────────────────────

const BuildStepContent = ({ metadata, status }: { metadata: Record<string, any>; status?: string }) => {
  const builds = (metadata.builds as any[]) || []
  const [expandedId, setExpandedId] = useState<string | null>(null)

  if (builds.length === 0) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
        <Text variant="subtext" theme="neutral">
          {status === 'in-progress' ? 'Starting component builds...' : 'Waiting to start builds...'}
        </Text>
      </div>
    )
  }

  const succeededCount = builds.filter((b: any) => b.status === 'success' || b.status === 'skipped').length
  const totalDuration = builds.reduce((acc: number, b: any) => acc + (b.duration || 0), 0)

  return (
    <div className="space-y-3">
      {/* Summary row */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-[13px] text-cool-grey-700 dark:text-cool-grey-200">
            <span className="font-semibold">{builds.length}</span> components built
          </span>
          <span className="text-[12px] text-cool-grey-400">·</span>
          <span className="text-[13px] font-semibold text-green-600 dark:text-green-400">
            {succeededCount} succeeded
          </span>
        </div>
        {totalDuration > 0 && (
          <span className="font-mono text-[12px] text-cool-grey-500 dark:text-cool-grey-400">
            {totalDuration.toFixed(1)}s total
          </span>
        )}
      </div>

      {/* Component build rows */}
      <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] divide-y divide-cool-grey-100 dark:divide-dark-grey-800 overflow-hidden">
        {builds.map((build: any, i: number) => {
          const buildStatus = build.status || 'pending'
          const isExpanded = expandedId === (build.component_id || i)

          return (
            <div key={build.component_id || i}>
              <button
                className="flex items-center gap-3 px-4 py-3 w-full text-left hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800 transition-colors"
                onClick={() => setExpandedId(isExpanded ? null : (build.component_id || i))}
              >
                <svg
                  width="12" height="12" viewBox="0 0 12 12" fill="none"
                  className={`text-cool-grey-400 shrink-0 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                >
                  <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                </svg>

                <DetailStatusIcon status={buildStatus === 'skipped' ? 'success' : buildStatus} />

                <span className="text-[13.5px] font-semibold text-cool-grey-900 dark:text-white">
                  {build.component_name || build.component_id}
                </span>

                {build.cache_status && (
                  <Badge theme={cacheBadgeTheme(build.cache_status)} size="sm">
                    {build.cache_status}
                  </Badge>
                )}

                <div className="flex-1" />

                {build.image_digest && (
                  <span className="font-mono text-[11.5px] text-cool-grey-400 dark:text-cool-grey-500 shrink-0">
                    {build.image_digest.length > 20 ? build.image_digest.substring(0, 20) : build.image_digest}
                  </span>
                )}

                {build.duration && (
                  <span className="font-mono text-[12.5px] text-cool-grey-500 dark:text-cool-grey-400 shrink-0 ml-2">
                    {Number(build.duration).toFixed(1)}s
                  </span>
                )}
              </button>
            </div>
          )
        })}
      </div>
    </div>
  )
}

// ─── PLAN GROUP STEP ──────────────────────────────────────────────

const PlanGroupStepContent = ({ step, metadata }: { step: TInstallWorkflowStep; metadata: Record<string, any> }) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const approvalId = step.approval?.id
  const hasApproval = step.execution_type === 'approval' && !!approvalId
  const hasResponse = !!step.approval?.response
  const isAwaiting = step.status?.status === 'approval-awaiting'

  const { data: plan } = useQuery({
    queryKey: ['approval-plan', orgId, step.id, approvalId],
    queryFn: async () => {
      const res = await fetch(
        `/api/orgs/${orgId}/workflows/${step.install_workflow_id}/steps/${step.id}/approvals/${approvalId}/contents`
      )
      if (!res.ok) throw new Error(`Failed to fetch approval contents: ${res.status}`)
      return res.json()
    },
    enabled: !!orgId && !!step.id && !!step.install_workflow_id && !!approvalId,
  })

  const { mutate: respond, isPending: isResponding } = useMutation({
    mutationFn: (responseType: 'approve' | 'deny' | 'deny-skip-current') =>
      approveWorkflowStep({
        orgId,
        workflowId: step.install_workflow_id,
        workflowStepId: step.id,
        approvalId: approvalId!,
        body: { response_type: responseType, note: '' },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Plan approved" theme="success">
          <Text>Approved install group plan.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['branch-run'] })
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Approval failed" theme="error">
          <Text>{err?.error || 'Unable to respond to approval.'}</Text>
        </Toast>
      )
    },
  })

  const installs = (plan?.installs || metadata.installs || []) as any[]
  const groupName = plan?.install_group || metadata.install_group_name || step.name?.replace(/^plan install group:\s*/i, '')

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon variant="ListChecksIcon" size={16} className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
          <span className="text-[13px] text-cool-grey-600 dark:text-cool-grey-300">
            install group:{' '}
            <span className="font-semibold text-cool-grey-900 dark:text-white">{groupName}</span>
          </span>
          <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">
            {installs.length} {installs.length === 1 ? 'install' : 'installs'}
          </span>
        </div>
      </div>

      {hasResponse && (
        <div className="flex items-center gap-2 px-4 py-3 rounded-[10px] border border-green-300 dark:border-green-700/40 bg-green-50 dark:bg-green-950/30">
          <Icon variant="CheckCircleIcon" size={18} className="text-green-600 dark:text-green-400 shrink-0" />
          <span className="text-[13px] text-green-700 dark:text-green-300">
            Plan {step.approval?.response?.response_type === 'approve' ? 'approved' : step.approval?.response?.response_type || 'responded'}
          </span>
        </div>
      )}

      {installs.length > 0 && (
        <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] divide-y divide-cool-grey-100 dark:divide-dark-grey-800 overflow-hidden">
          {installs.map((inst: any, i: number) => (
            <PlanInstallRow key={inst.install_id || i} install={inst} orgId={orgId} />
          ))}
        </div>
      )}

      {installs.length === 0 && (
        <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
          <Text variant="subtext" theme="neutral">
            {step.status?.status === 'in-progress' ? 'Computing install diffs...' : 'Waiting to compute plan...'}
          </Text>
        </div>
      )}

      {hasApproval && isAwaiting && !hasResponse && (
        <div className="flex items-center gap-3 px-4 py-3 rounded-[10px] border border-yellow-300 dark:border-yellow-700/40 bg-yellow-50 dark:bg-yellow-950/30">
          <Icon variant="WarningCircleIcon" size={18} className="text-yellow-600 dark:text-yellow-400 shrink-0" />
          <span className="text-[13px] text-yellow-700 dark:text-yellow-300 flex-1">
            Review the changes above and approve to proceed with deployment.
          </span>
          <div className="flex items-center gap-2 shrink-0">
            <Button
              variant="danger"
              size="sm"
              onClick={() => respond('deny-skip-current')}
              disabled={isResponding}
            >
              Skip
            </Button>
            <Button
              variant="primary"
              size="sm"
              onClick={() => respond('approve')}
              disabled={isResponding}
            >
              {isResponding ? 'Approving...' : 'Approve'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

const PlanInstallRow = ({ install, orgId }: { install: any; orgId: string }) => {
  const [expanded, setExpanded] = useState(false)
  const diff = install.diff

  const added = Array.isArray(diff?.added) ? diff.added : []
  const changed = Array.isArray(diff?.changed) ? diff.changed : []
  const removed = Array.isArray(diff?.removed) ? diff.removed : []

  const addedCount = added.length || (install.added ?? 0)
  const changedCount = changed.length || (install.changed ?? 0)
  const removedCount = removed.length || (install.removed ?? 0)
  const sandboxChanged = diff?.sandbox_changed || install.sandbox_changed
  const stackChanged = diff?.stack_changed || install.stack_changed
  const hasChanges = addedCount > 0 || changedCount > 0 || removedCount > 0 || sandboxChanged || stackChanged
  const hasDetailedDiff = added.length > 0 || changed.length > 0 || removed.length > 0

  const installLabels = install.install_labels as Record<string, string> | undefined
  const labelEntries = installLabels ? Object.entries(installLabels) : []

  return (
    <div>
      <button
        className="flex items-center gap-3 px-4 py-3 w-full text-left hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <svg
          width="12" height="12" viewBox="0 0 12 12" fill="none"
          className={`text-cool-grey-400 shrink-0 transition-transform ${expanded ? 'rotate-90' : ''}`}
        >
          <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>

        <InstallStatusIcon status={install.status} />

        <div className="flex items-center gap-2 min-w-0">
          <span className="text-[13.5px] font-semibold text-cool-grey-900 dark:text-white truncate">
            {install.install_name || install.install_id}
          </span>
          {install.install_id && (
            <a
              href={`/${orgId}/installs/${install.install_id}`}
              onClick={(e) => e.stopPropagation()}
              className="text-cool-grey-400 hover:text-primary-500 dark:text-cool-grey-500 dark:hover:text-primary-400 shrink-0"
            >
              <Icon variant="ArrowSquareOutIcon" size={14} />
            </a>
          )}
          {labelEntries.map(([k, v]) => (
            <span key={k} className="inline-flex items-center px-1.5 py-0.5 rounded border border-cool-grey-200 dark:border-dark-grey-600 bg-cool-grey-50 dark:bg-dark-grey-800 font-mono text-[10.5px] text-cool-grey-500 dark:text-cool-grey-400 shrink-0">
              {k}={v}
            </span>
          ))}
        </div>

        <div className="flex items-center gap-1.5 ml-auto shrink-0">
          {addedCount > 0 && (
            <span className="text-[12px] font-semibold text-green-600 dark:text-green-400">+{addedCount}</span>
          )}
          {changedCount > 0 && (
            <span className="text-[12px] font-semibold text-yellow-600 dark:text-yellow-400">~{changedCount}</span>
          )}
          {removedCount > 0 && (
            <span className="text-[12px] font-semibold text-red-600 dark:text-red-400">-{removedCount}</span>
          )}
          {sandboxChanged && (
            <Badge theme="warn" size="sm">sandbox</Badge>
          )}
          {stackChanged && (
            <Badge theme="warn" size="sm">stack</Badge>
          )}
          {!hasChanges && (
            <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">no changes</span>
          )}
        </div>
      </button>

      {expanded && hasChanges && hasDetailedDiff && (
        <div className="px-4 pb-3 pl-[52px] space-y-1">
          {added.map((c: any) => (
            <div key={c.component_id} className="flex items-center gap-2">
              <DiffMarker op="add" />
              <span className="font-mono text-[12.5px] text-cool-grey-700 dark:text-cool-grey-200">{c.component_name || c.component_id}</span>
              {c.component_type && (
                <span className="font-mono text-[11px] text-cool-grey-400 dark:text-cool-grey-500">{c.component_type}</span>
              )}
            </div>
          ))}
          {changed.map((c: any) => (
            <div key={c.component_id} className="flex items-center gap-2">
              <DiffMarker op="change" />
              <span className="font-mono text-[12.5px] text-cool-grey-700 dark:text-cool-grey-200">{c.component_name || c.component_id}</span>
              {c.component_type && (
                <span className="font-mono text-[11px] text-cool-grey-400 dark:text-cool-grey-500">{c.component_type}</span>
              )}
            </div>
          ))}
          {removed.map((c: any) => (
            <div key={c.component_id} className="flex items-center gap-2">
              <DiffMarker op="remove" />
              <span className="font-mono text-[12.5px] text-cool-grey-700 dark:text-cool-grey-200">{c.component_name || c.component_id}</span>
              {c.component_type && (
                <span className="font-mono text-[11px] text-cool-grey-400 dark:text-cool-grey-500">{c.component_type}</span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// ─── DEPLOY GROUP STEP ─────────────────────────────────────────────

const DeployGroupStepContent = ({ step, metadata }: { step: TInstallWorkflowStep; metadata: Record<string, any> }) => {
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
                className={`px-4 py-3 transition-colors ${isInstInProgress ? 'bg-blue-50/60 dark:bg-[rgba(63,116,224,0.06)]' : ''
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
}

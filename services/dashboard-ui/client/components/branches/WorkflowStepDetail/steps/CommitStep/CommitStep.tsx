import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { getInitials } from '../../shared/format'

interface ICommitStep {
  metadata: Record<string, any>
}

export const CommitStep = ({ metadata }: ICommitStep) => {
  const commitSha = metadata.commit_sha as string | undefined
  const commitMessage = metadata.commit_message as string | undefined
  const authorName = metadata.author_name as string | undefined
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

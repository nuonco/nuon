import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { buildCommitUrl } from '@/utils/vcs-urls'
import { getInitials } from '../../shared/format'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'

interface ICommitStep {
  metadata: Record<string, any>
}

export const CommitStep = ({ metadata }: ICommitStep) => {
  const commitSha = metadata.commit_sha as string | undefined
  const commitMessage = metadata.commit_message as string | undefined
  const authorName = metadata.author_name as string | undefined
  const branch = metadata.branch as string | undefined
  const repo = metadata.repo as string | undefined
  const commitUrl = buildCommitUrl(repo, commitSha)

  const prNumber = metadata.pr_number as number | undefined
  const prUrl = metadata.pr_url as string | undefined

  const filesChanged = metadata.files_changed as number | undefined
  const additions = metadata.additions as number | undefined
  const deletions = metadata.deletions as number | undefined
  const changedFiles = (metadata.changed_files as any[]) || []

  if (!commitSha) {
    return <StepStatePlaceholder>Fetching commit from VCS</StepStatePlaceholder>
  }

  const messageLines = commitMessage?.split('\n') || []
  const title = messageLines[0] || 'No message'
  const body = messageLines.slice(1).join('\n').trim()

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <Text as="p" variant="base" weight="strong" className="leading-snug">
            {title}
            {prNumber && (
              <>
                {' ('}
                {prUrl ? (
                  <Link href={prUrl} isExternal className="font-semibold">#{prNumber}</Link>
                ) : (
                  <span className="text-cool-grey-500 dark:text-cool-grey-400">#{prNumber}</span>
                )}
                {')'}
              </>
            )}
          </Text>
          {body && (
            <Text as="p" variant="body" theme="neutral" className="mt-1 whitespace-pre-wrap leading-relaxed">
              {body}
            </Text>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {branch && (
            <Badge size="sm" variant="code">
              {branch}
            </Badge>
          )}
          <ID className="text-[12.5px] font-mono">{commitSha?.substring(0, 7)}</ID>
          {commitUrl && (
            <Link
              href={commitUrl}
              isExternal
              aria-label="View commit"
              className="text-cool-grey-400 hover:text-cool-grey-600 dark:text-cool-grey-500 dark:hover:text-cool-grey-300"
            />
          )}
        </div>
      </div>

      <div className="flex items-center gap-2">
        <div className="w-[24px] h-[24px] rounded-full bg-primary-500 flex items-center justify-center shrink-0">
          <span className="text-[10px] font-semibold text-white leading-none">{getInitials(authorName)}</span>
        </div>
        <Text variant="subtext" theme="neutral">
          <span className="font-semibold text-cool-grey-900 dark:text-white">{authorName}</span> committed
        </Text>
      </div>

      {filesChanged !== undefined && (
        <div className="flex flex-col gap-2 pt-3 border-t">
          <div className="flex items-center gap-3 flex-wrap">
            <Text variant="subtext" theme="neutral">
              Showing{' '}
              <span className="font-semibold text-cool-grey-700 dark:text-cool-grey-200">{filesChanged}</span>{' '}
              changed {filesChanged === 1 ? 'file' : 'files'}
            </Text>
            {(additions ?? 0) > 0 && (
              <Text variant="body" weight="strong" theme="success">+{additions?.toLocaleString()}</Text>
            )}
            {(deletions ?? 0) > 0 && (
              <Text variant="body" weight="strong" theme="error">−{deletions?.toLocaleString()}</Text>
            )}
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

          {changedFiles.length > 0 && (
            <div className="border rounded-[10px] divide-y overflow-hidden">
              {changedFiles.map((file: any, i: number) => (
                <div key={file?.path || i} className="flex items-center justify-between px-4 py-2.5">
                  <div className="flex items-center gap-2 min-w-0">
                    <Icon variant="FileTextIcon" size={14} className="text-cool-grey-400 dark:text-cool-grey-500 shrink-0" />
                    <Text variant="subtext" family="mono" className="truncate">{file?.path}</Text>
                  </div>
                  <div className="flex items-center gap-2 shrink-0 ml-3">
                    {(file?.additions ?? 0) > 0 && (
                      <Text variant="subtext" weight="strong" theme="success">+{file?.additions}</Text>
                    )}
                    {(file?.deletions ?? 0) > 0 && (
                      <Text variant="subtext" weight="strong" theme="error">−{file?.deletions}</Text>
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

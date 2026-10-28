import { useMemo, useState, type CSSProperties } from 'react'
import { Badge } from '@/components/common/Badge'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { buildCommitUrl } from '@/utils/vcs-urls'
import '@/components/common/LabelBadge.css'
import type { TPreviewSourcePR, TPreviewSourcePRLabel } from '@/types'

export interface IPullRequestList {
  pullRequests: TPreviewSourcePR[]
  basePath: string
  isLoading?: boolean
}

const LOADING_ROWS = [0, 1, 2, 3]

export const openPullRequestLabel = (shown: number, total: number) =>
  shown === total ? `${total} open` : `${shown} of ${total} open`

export const pullRequestMatches = (
  pullRequest: TPreviewSourcePR,
  query: string,
  branches: Set<string>,
  authors: Set<string>
) => {
  if (branches.size > 0 && !branches.has(pullRequest.head_ref)) return false
  if (authors.size > 0 && !authors.has(pullRequest.author ?? '')) return false
  const needle = query.trim().toLowerCase().replace(/^#/, '')
  if (!needle) return true
  return [
    pullRequest.title,
    String(pullRequest.pr_number),
    pullRequest.head_ref,
    pullRequest.head_sha,
    pullRequest.author,
    ...(pullRequest.labels ?? []).map((label) => label.name),
  ]
    .filter(Boolean)
    .some((value) => value.toLowerCase().includes(needle))
}

const facetOptions = (values: Array<string | undefined>) =>
  [...new Set(values.filter((value): value is string => !!value))]
    .sort((a, b) => a.localeCompare(b))
    .map((value) => ({ label: value, value }))

const githubLabelColor = (color?: string) =>
  color && /^[0-9a-fA-F]{6}$/.test(color) ? `#${color}` : undefined

const repoFromPullRequestUrl = (url?: string) => {
  if (!url) return undefined
  const match = url.match(
    /^(https?:\/\/(?:[^/]+\.)?(?:github\.com|gitlab\.com|bitbucket\.org)\/[^/]+\/[^/]+)/i
  )
  return match?.[1]
}

const PullRequestLabel = ({ label }: { label: TPreviewSourcePRLabel }) => {
  const color = githubLabelColor(label.color)
  return (
    <Badge
      size="sm"
      theme={color ? 'none' : 'default'}
      className={color ? 'label-badge-value' : undefined}
      style={
        color
          ? ({ '--label-color': color } as CSSProperties)
          : undefined
      }
    >
      {label.name}
    </Badge>
  )
}

export const PullRequestList = ({
  pullRequests,
  basePath,
  isLoading = false,
}: IPullRequestList) => {
  const [search, setSearch] = useState('')
  const [branches, setBranches] = useState<Set<string>>(new Set())
  const [authors, setAuthors] = useState<Set<string>>(new Set())
  const branchChoices = useMemo(
    () => facetOptions(pullRequests.map((pullRequest) => pullRequest.head_ref)),
    [pullRequests]
  )
  const authorChoices = useMemo(
    () => facetOptions(pullRequests.map((pullRequest) => pullRequest.author)),
    [pullRequests]
  )
  const visible = useMemo(
    () =>
      pullRequests.filter((pullRequest) =>
        pullRequestMatches(pullRequest, search, branches, authors)
      ),
    [pullRequests, search, branches, authors]
  )
  const filtering =
    search.trim().length > 0 || branches.size > 0 || authors.size > 0

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row flex-wrap items-center justify-between gap-4">
        <SearchInput
          className="w-full md:w-fit"
          labelClassName="w-full md:w-fit"
          placeholder="Search pull requests"
          value={search}
          onChange={setSearch}
          onClear={() => setSearch('')}
        />
        {authorChoices.length > 1 || branchChoices.length > 1 ? (
          <div className="flex flex-wrap items-center gap-3">
            {authorChoices.length > 1 ? (
              <CheckboxFilterDropdown
                id="pull-request-author"
                label="Author"
                options={authorChoices}
                selected={authors}
                onChange={setAuthors}
              />
            ) : null}
            {branchChoices.length > 1 ? (
              <CheckboxFilterDropdown
                id="pull-request-branch"
                label="Branch"
                options={branchChoices}
                selected={branches}
                onChange={setBranches}
              />
            ) : null}
          </div>
        ) : null}
      </div>

      <div className="overflow-hidden rounded-lg border">
        <div className="flex items-center gap-2 bg-cool-grey-100 px-4 py-2 dark:bg-dark-grey-700">
          <Icon variant="GitPullRequestIcon" theme="success" size={16} />
          {isLoading ? (
            <Text variant="subtext" loading loadingWidth={8} />
          ) : (
            <Text variant="subtext" weight="strong">
              {openPullRequestLabel(visible.length, pullRequests.length)}
            </Text>
          )}
        </div>

        {isLoading ? (
          <ul>
            {LOADING_ROWS.map((row) => (
              <li
                key={row}
                className="flex items-start gap-3 border-t px-4 py-3"
              >
                <Icon variant="GitPullRequestIcon" theme="neutral" size={16} />
                <span className="flex min-w-0 flex-1 flex-col gap-1">
                  <Text variant="body" loading loadingWidth={36} />
                  <span className="flex flex-wrap items-center gap-2">
                    <Text variant="subtext" loading loadingWidth={4} />
                    <Badge size="sm" loading loadingWidth={16} />
                    <Text variant="subtext" loading loadingWidth={7} />
                  </span>
                </span>
              </li>
            ))}
          </ul>
        ) : visible.length === 0 ? (
          <div className="border-t">
            <EmptyState
              variant="table"
              emptyTitle={
                filtering ? 'No matching pull requests' : 'No open pull requests'
              }
              emptyMessage={
                filtering
                  ? 'Try a different title, number, author, or branch.'
                  : 'Pull requests targeting this branch show up here once GitHub reports them.'
              }
            />
          </div>
        ) : (
          <ul>
            {visible.map((pullRequest) => (
              <li
                key={pullRequest.pr_number}
                className="flex items-start gap-3 border-t px-4 py-3 hover:bg-black/5 focus-within:bg-black/5 dark:hover:bg-white/5 dark:focus-within:bg-white/5"
              >
                <Icon
                  className="mt-1 shrink-0"
                  variant="GitPullRequestIcon"
                  theme="success"
                  size={16}
                />
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                  <Text
                    as="div"
                    variant="body"
                    weight="strong"
                    className="min-w-0"
                  >
                    <span className="flex min-w-0 flex-wrap items-center gap-2">
                      <Link
                        href={`${basePath}/pull-requests/${pullRequest.pr_number}`}
                        className="!w-auto min-w-0 shrink truncate"
                        variant="inline"
                      >
                        {pullRequest.title ||
                          `Pull request #${pullRequest.pr_number}`}
                      </Link>
                      {pullRequest.is_draft ? (
                        <Badge size="sm" className="shrink-0">
                          Draft
                        </Badge>
                      ) : null}
                      {(pullRequest.labels ?? []).map((label) => (
                        <PullRequestLabel key={label.name} label={label} />
                      ))}
                    </span>
                  </Text>
                  <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
                    <Text variant="subtext" family="mono" theme="neutral">
                      #{pullRequest.pr_number}
                    </Text>
                    {pullRequest.author || pullRequest.created_at ? (
                      <span className="inline-flex flex-wrap items-center gap-1">
                        {pullRequest.author ? (
                          <Text variant="subtext" theme="neutral">
                            {pullRequest.author}
                          </Text>
                        ) : null}
                        {pullRequest.created_at ? (
                          <>
                            <Text variant="subtext" theme="neutral">
                              opened
                            </Text>
                            <Time
                              time={pullRequest.created_at}
                              format="relative"
                              variant="subtext"
                              theme="neutral"
                            />
                          </>
                        ) : null}
                      </span>
                    ) : null}
                    {pullRequest.head_ref ? (
                      <Text
                        variant="subtext"
                        family="mono"
                        theme="neutral"
                        flex
                      >
                        <Icon variant="GitBranchIcon" />
                        {pullRequest.head_ref}
                      </Text>
                    ) : null}
                    {pullRequest.head_sha ? (
                      <CommitLink
                        sha={pullRequest.head_sha}
                        href={
                          buildCommitUrl(
                            repoFromPullRequestUrl(pullRequest.url),
                            pullRequest.head_sha
                          ) ?? undefined
                        }
                        textVariant="subtext"
                      />
                    ) : null}
                  </span>
                </div>
                {pullRequest.url ? (
                  <Link
                    href={pullRequest.url}
                    isExternal
                    className="shrink-0 self-center"
                    textVariant="subtext"
                  >
                    GitHub
                  </Link>
                ) : null}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

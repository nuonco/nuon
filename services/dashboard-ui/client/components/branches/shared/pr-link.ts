import { externalGitUrl, gitReferenceUrl } from '@/utils/vcs-urls'

const GITHUB_PR_URL = /github\.com\/[^/]+\/[^/]+\/pull\/(\d+)/i
const PR_NUMBER_PAREN = /\(#(\d+)\)\s*$/
const PR_NUMBER_HASH = /(?:^|\s)#(\d+)(?:\s|$)/

export type TPrLink = {
  number: number
  url: string
}

export function resolvePrLink({
  repoSlug,
  prNumber,
  commitMessage,
}: {
  repoSlug?: string
  prNumber?: number
  commitMessage?: string
}): TPrLink | null {
  if (prNumber != null && repoSlug) {
    const url = gitReferenceUrl(repoSlug, {
      type: 'pull-request',
      number: prNumber,
    })
    if (url) return { number: prNumber, url }
  }

  if (!commitMessage) {
    return null
  }

  const urlMatch = commitMessage.match(GITHUB_PR_URL)
  if (urlMatch) {
    const number = Number.parseInt(urlMatch[1], 10)
    const url = externalGitUrl(urlMatch[0])
    if (!Number.isNaN(number) && url) {
      return { number, url }
    }
  }

  const parenMatch = commitMessage.match(PR_NUMBER_PAREN)
  if (parenMatch && repoSlug) {
    const number = Number.parseInt(parenMatch[1], 10)
    const url = gitReferenceUrl(repoSlug, {
      type: 'pull-request',
      number,
    })
    if (!Number.isNaN(number) && url) return { number, url }
  }

  const hashMatch = commitMessage.match(PR_NUMBER_HASH)
  if (hashMatch && repoSlug) {
    const number = Number.parseInt(hashMatch[1], 10)
    const url = gitReferenceUrl(repoSlug, {
      type: 'pull-request',
      number,
    })
    if (!Number.isNaN(number) && url) return { number, url }
  }

  return null
}

export function githubCommitUrl(
  repoSlug: string | undefined,
  sha: string | undefined
): string | undefined {
  if (!sha) return undefined
  return gitReferenceUrl(repoSlug, { type: 'commit', sha })
}

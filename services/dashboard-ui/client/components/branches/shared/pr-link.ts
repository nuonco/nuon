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
    return {
      number: prNumber,
      url: `https://github.com/${repoSlug}/pull/${prNumber}`,
    }
  }

  if (!commitMessage) {
    return null
  }

  const urlMatch = commitMessage.match(GITHUB_PR_URL)
  if (urlMatch) {
    const number = Number.parseInt(urlMatch[1], 10)
    if (!Number.isNaN(number)) {
      return { number, url: urlMatch[0] }
    }
  }

  const parenMatch = commitMessage.match(PR_NUMBER_PAREN)
  if (parenMatch && repoSlug) {
    const number = Number.parseInt(parenMatch[1], 10)
    if (!Number.isNaN(number)) {
      return {
        number,
        url: `https://github.com/${repoSlug}/pull/${number}`,
      }
    }
  }

  const hashMatch = commitMessage.match(PR_NUMBER_HASH)
  if (hashMatch && repoSlug) {
    const number = Number.parseInt(hashMatch[1], 10)
    if (!Number.isNaN(number)) {
      return {
        number,
        url: `https://github.com/${repoSlug}/pull/${number}`,
      }
    }
  }

  return null
}

export function githubCommitUrl(repoSlug: string | undefined, sha: string | undefined): string | undefined {
  if (!repoSlug || !sha) {
    return undefined
  }
  return `https://github.com/${repoSlug}/commit/${sha}`
}

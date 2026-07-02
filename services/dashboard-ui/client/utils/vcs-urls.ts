export const buildCommitUrl = (
  repo: string | undefined,
  sha: string | undefined
): string | null => {
  if (!repo || !sha) return null

  const cleanRepo = repo.replace(/\.git$/, '')

  if (cleanRepo.includes('github.com')) {
    return `${cleanRepo}/commit/${sha}`
  } else if (cleanRepo.includes('gitlab.com')) {
    return `${cleanRepo}/-/commit/${sha}`
  } else if (cleanRepo.includes('bitbucket.org')) {
    return `${cleanRepo}/commits/${sha}`
  }

  return `https://github.com/${cleanRepo}/commit/${sha}`
}

import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getVCSConnectionRepos, getConnectionBranches } from '@/lib'
import type { TVCSConnectionRepo, TVCSBranch } from '@/types'

export function useVcsRepoBrowser({
  orgId,
  vcsConnectionId,
  enabled = true,
  initialRepo,
  initialBranch = 'main',
}: {
  orgId: string
  vcsConnectionId: string
  enabled?: boolean
  initialRepo?: string
  initialBranch?: string
}) {
  const [selectedRepo, setSelectedRepo] = useState<TVCSConnectionRepo | null>(null)
  const [selectedBranch, setSelectedBranch] = useState(initialBranch)

  const {
    data: reposData,
    isLoading: loadingRepos,
    error: reposQueryError,
  } = useQuery({
    queryKey: ['vcs-repos', orgId, vcsConnectionId],
    queryFn: () => getVCSConnectionRepos({ orgId, connectionId: vcsConnectionId }),
    enabled: enabled && !!orgId && !!vcsConnectionId,
    retry: 2,
  })

  const repos = reposData?.repositories
    ? [...reposData.repositories].sort((a, b) => a.full_name.localeCompare(b.full_name))
    : []

  useEffect(() => {
    if (repos.length === 0) {
      if (initialRepo && !selectedRepo) {
        // Config has a repo that isn't in the VCS connection list (e.g. public repo).
        // Create a synthetic entry so the selector shows the correct value.
        setSelectedRepo({ full_name: initialRepo, name: initialRepo.split('/')[1] || initialRepo, private: false } as TVCSConnectionRepo)
      } else {
        setSelectedRepo(null)
      }
      return
    }
    if (initialRepo) {
      const match = repos.find((r) => r.full_name === initialRepo)
      if (match) {
        setSelectedRepo(match)
      } else {
        // Repo not in this connection's list — keep the saved value
        setSelectedRepo({ full_name: initialRepo, name: initialRepo.split('/')[1] || initialRepo, private: false } as TVCSConnectionRepo)
      }
    } else if (!selectedRepo) {
      setSelectedRepo(repos[0])
    }
  }, [repos.length, initialRepo])

  const [owner, repoName] = selectedRepo?.full_name?.split('/') ?? []

  const {
    data: branches = [],
    isLoading: loadingBranches,
    error: branchesQueryError,
  } = useQuery({
    queryKey: ['vcs-branches', orgId, vcsConnectionId, owner, repoName],
    queryFn: () => getConnectionBranches(orgId, vcsConnectionId, owner!, repoName!),
    enabled: enabled && !!orgId && !!vcsConnectionId && !!owner && !!repoName,
    retry: 2,
  })

  useEffect(() => {
    if (branches.length === 0) return
    const target = initialBranch || 'main'
    const match = branches.find((b) => b.name === target)
    if (match) {
      setSelectedBranch(match.name)
    } else {
      setSelectedBranch(branches[0].name)
    }
  }, [branches.length, initialBranch])

  return {
    repos,
    loadingRepos,
    reposError: reposQueryError ? 'Failed to load repositories. Please check your VCS connection.' : null,
    branches,
    loadingBranches,
    branchesError: branchesQueryError ? 'Failed to load branches. Please try again.' : null,
    selectedRepo,
    setSelectedRepo,
    selectedBranch,
    setSelectedBranch,
  }
}

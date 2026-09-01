import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { useVcsRepoBrowser } from '@/hooks/use-vcs-repo-browser'
import { createAppBranch, createBranchConfig, getAppBranches } from '@/lib'
import type {
  TAppBranchConfig,
  TCreateAppBranchRequest,
  TVCSConnectionRepo,
} from '@/types'
import { BranchFormModal } from '@/components/branches/BranchForm'
import type { BranchFormOutput } from '@/components/branches/BranchForm/schema'

type TCreateBranchBody = TCreateAppBranchRequest & {
  vcs_connection_id?: string
  connected_github_vcs_config?: {
    repo: string
    branch: string
    directory: string
  }
  public_git_vcs_config?: {
    repo: string
    branch: string
    directory: string
  }
}

const buildCreateBranchBody = (output: BranchFormOutput): TCreateBranchBody => {
  const body: TCreateBranchBody = { name: output.name }
  if (output.useVcs && output.selectedRepo) {
    const config = {
      repo: output.selectedRepo.full_name,
      branch: output.selectedBranch,
      directory: output.directory,
    }
    if (output.selectedRepo.private) {
      body.vcs_connection_id = output.selectedVcsConnectionId
      body.connected_github_vcs_config = config
    } else {
      body.public_git_vcs_config = config
    }
  }
  return body
}

function extractRepoFromConfig(config: TAppBranchConfig): {
  repo: TVCSConnectionRepo
  branch: string
  directory: string
  vcsConnectionId?: string
} | null {
  if (config.connected_github_vcs_config?.repo) {
    const fullName = config.connected_github_vcs_config.repo
    const parts = fullName.split('/')
    return {
      repo: {
        id: 0,
        name: parts.length > 1 ? parts.slice(1).join('/') : fullName,
        full_name: fullName,
        private: true,
        fork: false,
        html_url: `https://github.com/${fullName}`,
        default_branch: config.connected_github_vcs_config.branch || 'main',
        updated_at: '',
      },
      branch: config.connected_github_vcs_config.branch || '',
      directory: config.connected_github_vcs_config.directory || '.',
      vcsConnectionId: config.connected_github_vcs_config.vcs_connection_id,
    }
  }
  if (config.public_git_vcs_config?.repo) {
    const fullName = config.public_git_vcs_config.repo
    const parts = fullName.split('/')
    return {
      repo: {
        id: 0,
        name: parts.length > 1 ? parts.slice(1).join('/') : fullName,
        full_name: fullName,
        private: false,
        fork: false,
        html_url: `https://github.com/${fullName}`,
        default_branch: config.public_git_vcs_config.branch || 'main',
        updated_at: '',
      },
      branch: config.public_git_vcs_config.branch || '',
      directory: config.public_git_vcs_config.directory || '.',
    }
  }
  return null
}

type ICreateBranchModalContainer = IModal

export const CreateBranchModalContainer = ({
  onSubmit: _onSubmit,
  ...props
}: ICreateBranchModalContainer) => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const vcsConnections = org?.vcs_connections || []
  const [vcsConnectionId, setVcsConnectionId] = useState(
    vcsConnections[0]?.id || ''
  )

  const { data: existingBranches } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branches', org?.id, app?.id, 0],
    queryFn: () =>
      getAppBranches({ appId: app!.id, orgId: org!.id, limit: 20, offset: 0 }),
    enabled: !!org?.id && !!app?.id,
  })

  const existingRepoInfo = useMemo(() => {
    for (const branch of existingBranches?.data ?? []) {
      for (const config of branch.configs ?? []) {
        const info = extractRepoFromConfig(config)
        if (info) return info
      }
    }
    return null
  }, [existingBranches])

  const vcsBrowser = useVcsRepoBrowser({
    orgId: org.id,
    vcsConnectionId,
    enabled: !!vcsConnectionId,
  })

  const [didAutofill, setDidAutofill] = useState(false)

  useEffect(() => {
    if (didAutofill || !existingRepoInfo) return

    if (existingRepoInfo.vcsConnectionId) {
      setVcsConnectionId(existingRepoInfo.vcsConnectionId)
    } else {
      const repoOwner = existingRepoInfo.repo.full_name.split('/')[0]
      const match = vcsConnections.find(
        (c) => c.github_account_name?.toLowerCase() === repoOwner.toLowerCase()
      )
      if (match) {
        setVcsConnectionId(match.id)
      }
    }

    vcsBrowser.setSelectedRepo(existingRepoInfo.repo)
    if (existingRepoInfo.branch) {
      vcsBrowser.setSelectedBranch(existingRepoInfo.branch)
    }
    setDidAutofill(true)
  }, [
    existingRepoInfo,
    didAutofill,
    vcsConnections,
    vcsBrowser.setSelectedRepo,
    vcsBrowser.setSelectedBranch,
  ])

  const repos = useMemo(() => {
    const list = vcsBrowser.repos ?? []
    if (
      vcsBrowser.selectedRepo &&
      !list.some((r) => r.full_name === vcsBrowser.selectedRepo!.full_name)
    ) {
      return [vcsBrowser.selectedRepo, ...list]
    }
    return list
  }, [vcsBrowser.repos, vcsBrowser.selectedRepo])

  const {
    mutate,
    isPending: isLoading,
    error: submitError,
  } = useMutation({
    mutationFn: async (body: TCreateBranchBody) => {
      const branch = await createAppBranch({
        appId: app.id,
        body: { name: body.name },
        orgId: org.id,
      })

      if (body.connected_github_vcs_config) {
        await createBranchConfig({
          appId: app.id,
          branchId: branch.id,
          orgId: org.id,
          request: {
            connected_github_vcs_config: {
              vcs_connection_id: body.vcs_connection_id || '',
              ...body.connected_github_vcs_config,
            },
          },
        })
      } else if (body.public_git_vcs_config) {
        await createBranchConfig({
          appId: app.id,
          branchId: branch.id,
          orgId: org.id,
          request: {
            public_git_vcs_config: body.public_git_vcs_config,
          },
        })
      }

      return branch
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['app-branches'] })
      addToast(
        <Toast heading="Branch created" theme="success">
          <Text>
            Created app branch{' '}
            <Badge variant="code" size="md">
              {data.name}
            </Badge>
            .
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
      navigate(`/${org.id}/apps/${app.id}/branches/${data.id}`)
    },
  })

  return (
    <BranchFormModal
      mode="create"
      vcsConnections={vcsConnections}
      repos={repos}
      branches={vcsBrowser.branches}
      loadingRepos={vcsBrowser.loadingRepos && !didAutofill}
      loadingBranches={vcsBrowser.loadingBranches}
      reposError={vcsBrowser.reposError}
      branchesError={vcsBrowser.branchesError}
      selectedVcsConnectionId={vcsConnectionId}
      onVcsConnectionChange={setVcsConnectionId}
      selectedRepo={vcsBrowser.selectedRepo}
      onRepoChange={vcsBrowser.setSelectedRepo}
      selectedBranch={vcsBrowser.selectedBranch}
      onBranchChange={vcsBrowser.setSelectedBranch}
      defaultDirectory={existingRepoInfo?.directory ?? '.'}
      isSubmitting={isLoading}
      submitError={submitError}
      onSubmit={(output) => mutate(buildCreateBranchBody(output))}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const CreateBranchButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateBranchModalContainer />
  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      {props?.isMenuButton ? null : <Icon variant="PlusIcon" size={16} />}
      Create branch
      {props?.isMenuButton ? <Icon variant="PlusIcon" size={16} /> : null}
    </Button>
  )
}

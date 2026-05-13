import { useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
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
import { createAppBranch } from '@/lib'
import type { TAPIError, TCreateAppBranchRequest } from '@/types'
import { CreateBranchModal } from './CreateBranchModal'

type ICreateBranchModalContainer = IModal

export const CreateBranchModalContainer = ({
  onSubmit: _onSubmit,
  ...props
}: ICreateBranchModalContainer) => {
  const navigate = useNavigate()
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const vcsConnections = org?.vcs_connections || []
  const [vcsConnectionId, setVcsConnectionId] = useState(vcsConnections[0]?.id || '')

  const vcsBrowser = useVcsRepoBrowser({
    orgId: org.id,
    vcsConnectionId,
    enabled: !!vcsConnectionId,
  })

  const { mutate, isPending: isLoading } = useMutation({
    mutationFn: async (
      body: TCreateAppBranchRequest & {
        vcs_connection_id?: string
        connected_github_vcs_config?: {
          repo: string
          branch: string
          directory: string
          path_filter?: string
        }
        public_git_vcs_config?: {
          repo: string
          branch: string
          directory: string
          path_filter?: string
        }
      }
    ) => {
      return createAppBranch({ appId: app.id, body, orgId: org.id })
    },
    onSuccess: (data) => {
      addToast(
        <Toast heading="Branch created successfully" theme="success">
          <Text>Created app branch: {data.name}</Text>
        </Toast>
      )
      removeModal(props.modalId)
      navigate(`/${org.id}/apps/${app.id}/branches/${data.id}`)
    },
    onError: (error: TAPIError) => {
      addToast(
        <Toast heading="Branch creation failed" theme="error">
          <Text>{error.error || 'Failed to create app branch.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateBranchModal
      vcsConnections={vcsConnections}
      repos={vcsBrowser.repos}
      branches={vcsBrowser.branches}
      loadingRepos={vcsBrowser.loadingRepos}
      loadingBranches={vcsBrowser.loadingBranches}
      reposError={vcsBrowser.reposError}
      branchesError={vcsBrowser.branchesError}
      selectedVcsConnectionId={vcsConnectionId}
      onVcsConnectionChange={setVcsConnectionId}
      selectedRepo={vcsBrowser.selectedRepo}
      onRepoChange={vcsBrowser.setSelectedRepo}
      selectedBranch={vcsBrowser.selectedBranch}
      onBranchChange={vcsBrowser.setSelectedBranch}
      isSubmitting={isLoading}
      onSubmit={(body) => mutate(body)}
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
      <Icon variant="PlusIcon" size={16} />
      Create branch
    </Button>
  )
}

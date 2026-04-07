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
import { createAppBranch } from '@/lib'
import type { TCreateAppBranchRequest } from '@/types'
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

  const { mutate, isPending: isLoading } = useMutation({
    mutationFn: async (
      body: TCreateAppBranchRequest & {
        vcs_connection_id?: string
        connected_github_vcs_config?: any
        public_git_vcs_config?: any
      }
    ) => {
      return createAppBranch({ appId: app.id, body, orgId: org.id })
    },
    onSuccess: (data) => {
      const name = (data as any).name || ''
      addToast(
        <Toast heading="Branch created successfully" theme="success">
          <Text>Created app branch: {name}</Text>
        </Toast>
      )
      removeModal(props.modalId)
      navigate(`/${org.id}/apps/${app.id}/branches/${data.id}`)
    },
    onError: (error: Error) => {
      addToast(
        <Toast heading="Branch creation failed" theme="error">
          <Text>Failed to create app branch.</Text>
          <Text>{error.message || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateBranchModal
      orgId={org.id}
      vcsConnections={org?.vcs_connections || []}
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
      <Icon variant="Plus" size={16} />
      Create branch
    </Button>
  )
}

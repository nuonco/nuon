import { useForm, useStore } from '@tanstack/react-form'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { z } from 'zod'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { updateBranch } from '@/lib'
import type { TAPIError, TAppBranch } from '@/types'

const editBranchNameSchema = z.object({
  name: z.string().trim().min(1, 'Branch name is required'),
})

export const EditBranchNameModal = ({
  branchName,
  isPending,
  error,
  onSubmit,
  onCancel,
  ...props
}: {
  branchName: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: (name: string) => void
  onCancel: () => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: { name: branchName },
    validators: {
      onMount: editBranchNameSchema,
      onChange: editBranchNameSchema,
    },
    onSubmit: ({ value }) => onSubmit(value.name.trim()),
  })
  const name = useStore(form.store, (state) => state.values.name)
  const canSubmit = useStore(form.store, (state) => state.canSubmit)

  return (
    <Modal
      heading="Edit name"
      primaryActionTrigger={{
        children: isPending ? 'Saving...' : 'Save changes',
        disabled: !canSubmit || name.trim() === branchName || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        disabled: isPending,
        onClick: onCancel,
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner
          error={error}
          fallback="Unable to update branch name"
        />
        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="branch-name"
              labelProps={{ labelText: 'Branch name' }}
              placeholder="production"
              disabled={isPending}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}

const EditBranchNameOnlyModalContainer = ({
  branch,
  onSuccess,
  onSubmit: _onSubmit,
  ...props
}: {
  branch: TAppBranch
  onSuccess?: () => void
} & IModal) => {
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()

  const {
    mutate: save,
    isPending,
    error,
  } = useMutation<TAppBranch, TAPIError, string>({
    mutationFn: (name) =>
      updateBranch({
        appId: app?.id ?? '',
        branchId: branch.id ?? '',
        orgId: org?.id ?? '',
        request: { name },
      }),
    onSuccess: (updatedBranch) => {
      queryClient.invalidateQueries({
        queryKey: ['app-branch', org?.id, app?.id, branch.id],
      })
      queryClient.invalidateQueries({
        queryKey: ['app-branches', org?.id, app?.id],
      })
      addToast(
        <Toast heading="Branch name updated" theme="success">
          <Text>Renamed this branch to {updatedBranch.name}.</Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
  })

  return (
    <EditBranchNameModal
      branchName={branch.name ?? ''}
      isPending={isPending}
      error={error ?? null}
      onSubmit={(name) => save(name)}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const EditBranchNameButton = ({
  branch,
  onSuccess,
  ...props
}: {
  branch: TAppBranch
  onSuccess?: () => void
} & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = (
    <EditBranchNameOnlyModalContainer branch={branch} onSuccess={onSuccess} />
  )

  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PencilSimpleLineIcon" size={16} />
      Edit name
    </Button>
  )
}

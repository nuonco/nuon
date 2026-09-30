import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { updateOrg } from '@/lib'
import { orgNameSubmitError } from './org-name-error'
import { UpdateOrgNameForm } from './UpdateOrgName'

export const UpdateOrgName = () => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { addToast } = useToast()
  const currentName = org.name ?? ''

  const { mutate, isPending, error } = useMutation({
    mutationFn: (name: string) => updateOrg({ orgId: org.id, body: { name } }),
    onSuccess: (updated, name) => {
      queryClient.setQueryData(['org', org.id], updated)
      queryClient.invalidateQueries({ queryKey: ['org', org.id] })
      queryClient.invalidateQueries({ queryKey: ['orgs'] })
      addToast(
        <Toast heading="Organization renamed" theme="success">
          <Text>
            Renamed {currentName} to {name}.
          </Text>
        </Toast>
      )
    },
  })

  return (
    <UpdateOrgNameForm
      key={currentName}
      currentName={currentName}
      isPending={isPending}
      error={orgNameSubmitError(error)}
      onSubmit={mutate}
    />
  )
}

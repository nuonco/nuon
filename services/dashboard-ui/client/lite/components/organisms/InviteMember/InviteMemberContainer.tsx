import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { inviteUser, listRoles } from '@/lib'
import { useSurfaces } from '../../../hooks/use-surfaces'
import { useToast } from '../../../hooks/use-toast'
import { useOrg } from '../../../providers/org-provider'
import { InviteMember, type TInviteMemberValues } from './InviteMember'

export const InviteMemberContainer = () => {
  const { orgId } = useOrg()
  const queryClient = useQueryClient()
  const { closeTopSurface } = useSurfaces()
  const { addToast } = useToast()

  const { data: roles = [], isLoading: rolesLoading } = useQuery({
    queryKey: ['roles', orgId, 'team'],
    queryFn: () => listRoles({ orgId: orgId!, context: 'team' }),
    enabled: !!orgId,
    staleTime: 60_000,
  })

  const mutation = useMutation({
    mutationFn: (values: TInviteMemberValues) =>
      inviteUser({
        orgId: orgId ?? '',
        body: {
          email: values.email.trim(),
          role_type: values.roleType,
        },
      }),
    onSuccess: (invite) => {
      void queryClient.invalidateQueries({
        queryKey: ['org-members', orgId],
      })
      closeTopSurface()
      addToast({
        heading: 'Invite sent',
        description: `Sent an invitation to ${invite?.email ?? 'the team member'}.`,
        theme: 'success',
      })
    },
  })

  return (
    <InviteMember
      roles={roles}
      rolesLoading={rolesLoading}
      onSubmit={(values) => mutation.mutate(values)}
      pending={mutation.isPending}
      error={mutation.error}
    />
  )
}

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { getAccount } from '@/lib/ctl-api/accounts'
import { updateOrgTelemetry } from '@/lib/ctl-api/orgs'
import type { TOrg } from '@/types'
import { OrgTelemetryModal } from './OrgTelemetry'

export const OrgTelemetryModalContainer = ({
  org,
  ...props
}: { org: TOrg } & Omit<IModal, 'onSubmit'>) => {
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const { mutate, isPending, error } = useMutation({
    mutationFn: (enabled: boolean) =>
      updateOrgTelemetry({ orgId: org.id, enabled }),
    onSuccess: (updatedOrg) => {
      queryClient.setQueryData(['org', org.id], updatedOrg)
      queryClient.invalidateQueries({ queryKey: ['org', org.id] })
      queryClient.invalidateQueries({ queryKey: ['orgs'] })
      queryClient.invalidateQueries({ queryKey: ['install-telemetry', org.id] })
      queryClient.invalidateQueries({ queryKey: ['runner', org.id] })
      addToast(
        <Toast heading="Telemetry settings updated" theme="success">
          <Text>
            Installs using the org default will apply it on their next runner
            refresh.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <OrgTelemetryModal
      orgName={org.name || org.id}
      enabled={org.telemetry?.enabled ?? false}
      isPending={isPending}
      error={error}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const OrgTelemetryButton = (
  props: Omit<IButtonAsButton, 'children'>
) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const { data: account } = useQuery({
    queryKey: ['account'],
    queryFn: getAccount,
    staleTime: 60_000,
    enabled: !!org?.id,
  })
  const isOrgAdmin = account?.roles?.some(
    (role) => role.org_id === org?.id && role.role_type === 'org_admin'
  )

  if (!org || !isOrgAdmin) return null

  return (
    <Button
      onClick={() => addModal(<OrgTelemetryModalContainer org={org} />)}
      {...props}
    >
      Manage telemetry <Icon variant="SlidersHorizontalIcon" />
    </Button>
  )
}

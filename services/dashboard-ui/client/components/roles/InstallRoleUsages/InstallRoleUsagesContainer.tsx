import { useQuery } from '@tanstack/react-query'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallRoleUsages } from '@/lib'
import type { TInstallRoleUsage } from '@/types'
import {
  InstallRoleUsagesModal as InstallRoleUsagesModalComponent,
  InstallRoleUsagesTrigger as InstallRoleUsagesTriggerComponent,
} from './InstallRoleUsages'

interface IInstallRoleUsages {
  installId: string
  roleName: string
  roleDisplayName?: string
}

export const InstallRoleUsagesModal = ({
  installId,
  roleName,
  roleDisplayName,
  ...props
}: IInstallRoleUsages & IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()

  const {
    data: usages,
    error,
    isLoading,
  } = useQuery<TInstallRoleUsage[]>({
    queryKey: ['install-role-usages', org?.id, installId, roleName],
    queryFn: () =>
      getInstallRoleUsages({ installId, orgId: org.id, roleName }),
    enabled: !!org?.id && !!installId && !!roleName,
  })

  return (
    <InstallRoleUsagesModalComponent
      orgId={org?.id ?? ''}
      installId={installId}
      usages={usages}
      isLoading={isLoading}
      error={error}
      roleDisplayName={roleDisplayName}
      onNavigate={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const InstallRoleUsagesTrigger = ({
  installId,
  roleName,
  roleDisplayName,
}: IInstallRoleUsages) => {
  const { addModal } = useSurfaces()
  const modal = (
    <InstallRoleUsagesModal
      installId={installId}
      roleName={roleName}
      roleDisplayName={roleDisplayName}
    />
  )
  return (
    <InstallRoleUsagesTriggerComponent onOpenModal={() => addModal(modal)} />
  )
}

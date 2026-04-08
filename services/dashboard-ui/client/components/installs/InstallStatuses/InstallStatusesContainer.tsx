import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import { InstallStatuses, type IInstallStatuses } from './InstallStatuses'

export const InstallStatusesContainer = (
  props: Omit<IInstallStatuses, 'install' | 'stack'>
) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { data: stack } = useQuery({
    queryKey: ['install-stack', org.id, install.id],
    queryFn: () => getInstallStack({ installId: install.id, orgId: org.id }),
    enabled: !!install.id,
  })
  return <InstallStatuses install={install} stack={stack} {...props} />
}

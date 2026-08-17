import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import {
  InstallStacksTable,
  parseInstallStackSummaryToTableData,
} from './InstallStacksTable'

const LIMIT = 10

export const InstallStacksTableContainer = ({
  pollInterval = 20000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: stack, isLoading } = useQuery({
    queryKey: ['install-stack', org?.id, install?.id],
    queryFn: () => getInstallStack({ orgId: org.id, installId: install.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const pagination = { hasNext: false, offset: 0, limit: LIMIT }

  if (isLoading) {
    return <InstallStacksTable data={[]} isLoading pagination={pagination} />
  }
  if (!stack) return null

  return (
    <InstallStacksTable
      data={parseInstallStackSummaryToTableData(stack, org.id, install.app_id)}
      isLoading={false}
      pagination={pagination}
    />
  )
}

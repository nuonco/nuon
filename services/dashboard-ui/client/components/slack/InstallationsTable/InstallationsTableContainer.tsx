import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getSlackInstallations, getSlackOrgLinks } from '@/lib'
import { InstallationsTable } from './InstallationsTable'

export const InstallationsTableContainer = ({
  pollInterval = 30000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const { org } = useOrg()

  const installationsQuery = useQuery({
    queryKey: ['slack-installations', org.id],
    queryFn: () => getSlackInstallations({ orgId: org.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const linksQuery = useQuery({
    queryKey: ['slack-org-links', org.id],
    queryFn: () => getSlackOrgLinks({ orgId: org.id }),
  })

  return (
    <InstallationsTable
      data={installationsQuery.data ?? []}
      links={linksQuery.data ?? []}
      isLoading={installationsQuery.isLoading || linksQuery.isLoading}
    />
  )
}

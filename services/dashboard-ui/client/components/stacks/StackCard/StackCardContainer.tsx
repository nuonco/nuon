import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { getInstallStack } from '@/lib'
import { StackCard } from './StackCard'

export const StackCardContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: stack, isLoading, error } = useQuery({
    queryKey: ['install-stack', org.id, install.id],
    queryFn: () =>
      getInstallStack({
        installId: install.id!,
        orgId: org.id!,
      }),
    enabled: !!install.id && !!org.id,
  })

  if (error) {
    return <StackCard error="Failed to load stack" />
  }

  if (!isLoading && !stack?.versions?.length) {
    return <StackCard error="No stack found" />
  }

  const latestVersion = stack?.versions?.[0]
  const runCount = stack?.versions?.reduce(
    (sum, v) => sum + (v.runs?.length || 0),
    0
  )
  const createdAt = stack?.versions?.[stack.versions.length - 1]?.created_at

  const href = `/${org.id}/installs/${install.id}/stacks`

  return (
    <StackCard
      status={latestVersion?.composite_status?.status}
      runCount={runCount}
      createdAt={createdAt}
      href={href}
      isLoading={isLoading}
    />
  )
}

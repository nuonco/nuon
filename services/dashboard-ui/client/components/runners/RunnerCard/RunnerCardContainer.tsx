import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { RunnerCard } from './RunnerCard'

export const RunnerCardContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  if (!install.runner_id) {
    return <RunnerCard error="No runner found" />
  }

  const href = `/${org.id}/installs/${install.id}/runner`

  return (
    <RunnerCard
      status={install.runner_status}
      href={href}
    />
  )
}

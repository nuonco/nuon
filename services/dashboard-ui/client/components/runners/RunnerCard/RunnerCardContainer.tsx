import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { RunnerCard } from './RunnerCard'

export const RunnerCardContainer = () => {
  const { install } = useInstall()
  const installLink = useInstallLink()

  if (!install.runner_id) {
    return <RunnerCard error="No runner found" />
  }

  const href = installLink({ installId: install.id, appId: install.app_id, suffix: `/runner` })

  return (
    <RunnerCard
      status={install.runner_status}
      href={href}
    />
  )
}

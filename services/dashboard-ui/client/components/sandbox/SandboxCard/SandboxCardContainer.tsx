import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { SandboxCard } from './SandboxCard'

export const SandboxCardContainer = () => {
  const { install } = useInstall()
  const installLink = useInstallLink()

  if (!install.sandbox) {
    return <SandboxCard error="No sandbox found" />
  }

  const href = installLink({ installId: install.id, appId: install.app_id, suffix: `/sandbox` })

  return (
    <SandboxCard
      status={install.sandbox_status}
      href={href}
    />
  )
}

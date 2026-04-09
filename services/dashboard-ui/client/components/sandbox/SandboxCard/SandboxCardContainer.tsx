import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { SandboxCard } from './SandboxCard'

export const SandboxCardContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  if (!install.sandbox) {
    return <SandboxCard error="No sandbox found" />
  }

  const href = `/${org.id}/installs/${install.id}/sandbox`

  return (
    <SandboxCard
      status={install.sandbox_status}
      href={href}
    />
  )
}

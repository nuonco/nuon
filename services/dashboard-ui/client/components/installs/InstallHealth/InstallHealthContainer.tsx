import { HealthTimeline } from '@/components/install-health/HealthTimeline'
import { InstallResourcesTable } from '@/components/install-resources/InstallResourcesTable'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { InstallHealth } from './InstallHealth'

export const InstallHealthContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const resourcesComponentsPath = `/${org?.id}/installs/${install?.id}/resources/components`

  return (
    <InstallHealth
      timeline={
        <HealthTimeline
          days={30}
          shouldPoll
          getComponentHref={({ component_id, component_name }) => {
            const q = encodeURIComponent(component_name || component_id)
            return `${resourcesComponentsPath}?q=${q}`
          }}
        />
      }
      resources={<InstallResourcesTable shouldPoll />}
    />
  )
}

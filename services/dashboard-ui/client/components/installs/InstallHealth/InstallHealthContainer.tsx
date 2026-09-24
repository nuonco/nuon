import {
  HealthTimeline,
  isImageComponentType,
} from '@/components/install-health/HealthTimeline'
import { InstallResourcesTable } from '@/components/install-resources/InstallResourcesTable'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { InstallHealth } from './InstallHealth'

export const InstallHealthContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const resourcesPath = `/${org?.id}/installs/${install?.id}/resources`

  return (
    <InstallHealth
      timeline={
        <HealthTimeline
          days={30}
          shouldPoll
          groupByKind
          getComponentHref={({
            component_id,
            component_name,
            component_type,
          }) => {
            const q = encodeURIComponent(component_name || component_id)
            const page = isImageComponentType(component_type)
              ? 'images'
              : 'components'
            return `${resourcesPath}/${page}?q=${q}`
          }}
        />
      }
      resources={<InstallResourcesTable shouldPoll />}
    />
  )
}

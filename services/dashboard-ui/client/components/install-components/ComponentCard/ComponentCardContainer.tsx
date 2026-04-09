import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { getInstallComponent } from '@/lib'
import { ComponentCard } from './ComponentCard'

interface IComponentCardContainer {
  id?: string
  name?: string
}

export const ComponentCardContainer = ({ id, name }: IComponentCardContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const resolvedComponent = name
    ? install.install_components?.find((ic) => ic.component?.name === name)
    : id
      ? install.install_components?.find((ic) => ic.component_id === id)
      : undefined

  const componentId = id || resolvedComponent?.component_id

  const { data: installComponent, isLoading, error } = useQuery({
    queryKey: ['install-component', componentId],
    queryFn: () =>
      getInstallComponent({
        installId: install.id!,
        componentId: componentId!,
        orgId: org.id!,
      }),
    enabled: !!componentId && !!install.id && !!org.id,
  })

  if (!id && !name) {
    return <ComponentCard error="Missing id or name attribute" />
  }

  if (name && !resolvedComponent && !isLoading) {
    return <ComponentCard error={`Component "${name}" not found`} />
  }

  const component = installComponent?.component
  const href = componentId
    ? `/${org.id}/installs/${install.id}/components/${componentId}`
    : undefined

  return (
    <ComponentCard
      name={component?.name || resolvedComponent?.component?.name || name}
      type={component?.type || resolvedComponent?.component?.type}
      status={installComponent?.status}
      href={href}
      isLoading={isLoading}
      error={error ? 'Failed to load component' : undefined}
    />
  )
}

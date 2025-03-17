import React, { type FC } from 'react'
import { Link } from '@/components/Link'
import { Text } from '@/components/Typography'
import type { TComponent, TInstallComponent } from '@/types'

export interface IComponentDependencies {
  appComponents?: Array<TComponent>
  appId?: string
  dependentIds: Array<string>
  installComponents?: Array<TInstallComponent>
  installId?: string
  orgId: string
}

// TODO(nnnnat): rename to ComponentDependencies
export const DependentComponents: FC<IComponentDependencies> = ({
  appComponents,
  appId,
  dependentIds,
  installComponents,
  installId,
  orgId,
}) => {
  const path = appId
    ? `/${orgId}/apps/${appId}/components`
    : `/${orgId}/installs/${installId}/components`

  return (
    <div className="flex flex-wrap items-center justify-start gap-3">
      {appComponents &&
        appComponents
          .filter((comp) => dependentIds.some((depId) => comp.id === depId))
          .map((dep, i) => (
            <Text
              key={`${dep.id}-${i}`}
              className="bg-gray-500/10 p-2 rounded-lg border w-fit"
            >
              <Link href={`${path}/${dep.id}`}>{dep.name}</Link>
            </Text>
          ))}

      {installComponents &&
        installComponents
          .filter((comp) =>
            dependentIds.some((depId) => comp.component_id === depId)
          )
          .map((dep, i) => (
            <Text
              key={`${dep.id}-${i}`}
              className="bg-gray-500/10 p-2 rounded-lg border w-fit"
            >
              <Link href={`${path}/${dep.component_id}`}>{dep.component?.name}</Link>
            </Text>
          ))}
    </div>
  )
}

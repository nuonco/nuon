import React, { type FC } from 'react'
import { ComponentDependencies } from '@/components/old/Components'
import { Link } from '@/components/old/Link'
import { Text } from '@/components/old/Typography'
import type { TComponent, TInstallComponent } from '@/types'

export interface IComponentDependencies {
  appComponents?: Array<TComponent>
  appId?: string
  dependentIds: Array<string>
  installComponents?: Array<TInstallComponent>
  installId?: string
  name: string
  orgId: string
}

// TODO(nnnnat): rename to ComponentDependencies
export const DependentComponents: FC<IComponentDependencies> = ({
  appComponents,
  appId,
  dependentIds,
  installComponents,
  installId,
  name,
  orgId,
}) => {
  const path = appId
    ? `/${orgId}/apps/${appId}/components`
    : `/${orgId}/installs/${installId}/components`

  return (
    <div className="flex flex-wrap items-center justify-start gap-3">
      {appComponents?.length ? (
        <ComponentDependencies
          deps={appComponents.filter((comp) =>
            dependentIds.some((depId) => comp.id === depId)
          )}
          name={name}
        />
      ) : null}

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
              <Link href={`${path}/${dep.component_id}`}>
                {dep.component?.name}
              </Link>
            </Text>
          ))}
    </div>
  )
}

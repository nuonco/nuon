import React, { Suspense, type FC } from 'react'
import { Card, Heading, Link, Text } from '@/components'
import { getComponent, type IGetComponent } from '@/lib'
import type { TComponent, TInstallComponent } from '@/types'

export const ComponentDependencies: FC<IGetComponent> = async (props) => {
  let component: TComponent
  try {
    component = await getComponent(props)
  } catch (error) {
    return (
      <Text variant="label">Error: Can not find component dependencies</Text>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {component.dependencies?.length ? (
        component.dependencies.map((d) => (
          <Text variant="overline" key={d}>
            {d}
          </Text>
        ))
      ) : (
        <Text variant="overline">No dependencies to show</Text>
      )}
    </div>
  )
}

export const ComponentDependenciesCard: FC<
  IGetComponent & { heading?: string }
> = ({ heading = 'Dependencies', ...props }) => (
  <Card className="flex-1">
    <Heading>{heading}</Heading>
    <Suspense fallback="Loading component dependencies...">
      <ComponentDependencies {...props} />
    </Suspense>
  </Card>
)

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
              className="bg-gray-500/10 leading-3  p-2 rounded-lg border w-fit"
              variant="caption"
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
              className="bg-gray-500/10 leading-3  p-2 rounded-lg border w-fit"
              variant="caption"
            >
              <Link href={`${path}/${dep.id}`}>{dep.component?.name}</Link>
            </Text>
          ))}
    </div>
  )
}

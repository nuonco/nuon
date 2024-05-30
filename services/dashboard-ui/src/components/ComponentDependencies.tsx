import React, { Suspense, type FC } from 'react'
import { Card, Heading, Text } from '@/components'
import { getComponent, type IGetComponent } from '@/lib'
import type { TComponent } from '@/types'

export const ComponentDependencies: FC<IGetComponent> = async (props) => {
  let component: TComponent
  try {
    component = await getComponent(props)
  } catch (error) {
    return <Text variant="label">Error: Can not find component dependencies</Text>
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

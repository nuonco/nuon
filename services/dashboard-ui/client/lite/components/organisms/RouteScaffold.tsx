import type { ReactNode } from 'react'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'

export interface IRouteScaffold {
  title: string
  description: string
  children?: ReactNode
}

export const RouteScaffold = ({
  title,
  description,
  children,
}: IRouteScaffold) => (
  <section className="flex w-full flex-col gap-6">
    <div className="flex flex-col gap-1">
      <Text as="h1" variant="title">
        {title}
      </Text>
      <Text as="p" variant="caption" color="secondary">
        {description}
      </Text>
    </div>
    {children ?? (
      <Card className="min-h-40">
        <Text variant="caption" color="tertiary">
          Page content will be added in a follow-up.
        </Text>
      </Card>
    )}
  </section>
)

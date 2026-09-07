import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'

export interface IRouteScaffold {
  title: string
  description: string
}

export const RouteScaffold = ({ title, description }: IRouteScaffold) => (
  <section className="mx-auto flex w-full max-w-6xl flex-col gap-6">
    <div className="flex flex-col gap-1">
      <Text as="h1" variant="title">
        {title}
      </Text>
      <Text as="p" variant="caption" color="secondary">
        {description}
      </Text>
    </div>
    <Card className="min-h-40">
      <Text variant="caption" color="tertiary">
        Page content will be added in a follow-up.
      </Text>
    </Card>
  </section>
)

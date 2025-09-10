import { ComponentDependencies, Text } from '@/components'
import type { TComponent } from '@/types'
import { api } from '@/lib/api'

export const Dependencies = async ({
  component,
  orgId,
}: {
  component: TComponent
  orgId: string
}) => {
  const { data, error } = await api<TComponent[]>({
    orgId,
    path: `components/${component?.id}/dependencies`,
  })

  return (
    <div className="flex items-center gap-4">
      {error ? (
        <Text>{error?.error}</Text>
      ) : (
        <ComponentDependencies deps={data} name={component?.name} />
      )}
    </div>
  )
}

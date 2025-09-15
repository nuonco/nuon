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
    path: `apps/${component?.app_id}/components`,
  })

  const deps = data
    ? data?.filter((comp) => component?.dependencies?.includes(comp?.id))
    : []

  return (
    <div className="flex items-center gap-4">
      {deps && !error ? (
        <ComponentDependencies deps={deps} name={component?.name} />
      ) : (
        <Text>{error?.error}</Text>
      )}
    </div>
  )
}

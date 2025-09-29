import { ComponentDependencies as Deps, Text } from '@/components'
import { api } from '@/lib/api'
import type { TComponent } from '@/types'

export const ComponentDependencies = async ({
  component,
  installId,
  orgId,
}: {
  component: TComponent
  installId: string
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
        <Deps deps={data} installId={installId} name={component?.name} />
      )}
    </div>
  )
}

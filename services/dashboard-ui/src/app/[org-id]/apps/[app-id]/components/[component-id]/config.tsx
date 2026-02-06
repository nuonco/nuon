import { ComponentConfigCard, ComponentConfigCardSkeleton } from "@/components/components/ComponentConfigCard"
import { EmptyState } from '@/components/common/EmptyState'
import type { TComponentConfig } from '@/types'
import { api } from '@/lib/api'

// TODO(nnnat): get the component config form the app config
export const Config = async ({
  componentId,
  orgId,
}: {
  componentId: string
  orgId: string
}) => {
  const { data: componentConfig, error } = await api<TComponentConfig>({
    orgId,
    path: `components/${componentId}/configs/latest`,
  })

  return error ? (
    <ConfigError />
  ) : (
    <ComponentConfigCard config={componentConfig} />
  )
}

export const ConfigError = () => (
  <EmptyState
    variant="table"
    emptyTitle="Unable to load component config"
    emptyMessage="There was an error loading the component configuration."
  />
)

export const ConfigSkeleton = ComponentConfigCardSkeleton

import { api } from '@/lib/api'

export type TAppLabelKey = {
  key: string
  color: string
  default_color: string
  is_override: boolean
  values: string[]
  entity_types: string[]
  usage_count: number
}

export type TAppLabelsResponse = {
  labels: TAppLabelKey[]
  label_colors: Record<string, string>
  default_colors: string[]
}

export const getAppLabels = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<TAppLabelsResponse>({
    path: `apps/${appId}/labels`,
    orgId,
  })

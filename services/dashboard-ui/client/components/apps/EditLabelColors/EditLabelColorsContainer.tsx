import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppLabels, updateApp } from '@/lib'
import type { TAPIError } from '@/types/dashboard.types'
import { EditLabelColorsModal } from './EditLabelColors'
import type { IModal } from '@/components/surfaces/Modal'

export const EditLabelColorsContainer = (props: IModal) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()

  const { data: labelsData } = useQuery({
    queryKey: ['app-labels', org?.id, app?.id],
    queryFn: () => getAppLabels({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const defaults: Record<string, string> = {}
  for (const lk of labelsData?.labels ?? []) {
    if (!lk.is_override) {
      defaults[lk.key] = lk.color
    }
  }

  const { mutate, isPending, error } = useMutation({
    mutationFn: (labelColors: Record<string, string>) =>
      updateApp({
        appId: app.id,
        orgId: org.id,
        body: { label_colors: labelColors },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app', org?.id, app?.id] })
      queryClient.invalidateQueries({ queryKey: ['app-labels', org?.id, app?.id] })
      removeModal(props.modalId)
    },
  })

  const initialColors: Record<string, string> = { ...defaults }
  for (const [k, v] of Object.entries(app?.label_colors ?? {})) {
    initialColors[k] = v
  }

  const allLabelKeys = (labelsData?.labels ?? []).map((lk) => lk.key)

  return (
    <EditLabelColorsModal
      labelColors={initialColors}
      defaultColors={labelsData?.default_colors ?? []}
      availableKeys={allLabelKeys}
      isPending={isPending}
      error={(error as TAPIError) ?? null}
      onSubmit={(labelColors) => mutate(labelColors)}
      {...props}
    />
  )
}

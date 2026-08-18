import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { getAppLabels, updateApp } from '@/lib'
import type { TAPIError } from '@/types/dashboard.types'
import { AppLabels } from './AppLabels'
import { ResetLabelColorsModal } from './ResetLabelColorsModal'

export const AppLabelsContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const { addModal } = useSurfaces()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-labels', org?.id, app?.id],
    queryFn: () => getAppLabels({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const { mutate: saveLabelColors, isPending } = useMutation({
    mutationFn: (labelColors: Record<string, string>) =>
      updateApp({ orgId: org.id, appId: app.id, body: { label_colors: labelColors } }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-labels', org?.id, app?.id] })
      queryClient.invalidateQueries({ queryKey: ['app', org?.id, app?.id] })
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Label update failed" theme="error">
          <Text>{err?.error || 'Unable to update label colors.'}</Text>
        </Toast>
      )
    },
  })

  const overrides = data?.label_colors ?? {}
  const labels = data?.labels ?? []
  const hasOverrides = labels.some((l) => l.is_override)

  const handleOverride = (key: string, color: string) => {
    saveLabelColors({ ...overrides, [key]: color })
  }

  const handleRemoveOverride = (key: string) => {
    const next = { ...overrides }
    delete next[key]
    saveLabelColors(next)
  }

  const handleResetAll = () => {
    addModal(<ResetLabelColorsModal orgId={org.id} appId={app.id} appName={app?.name} />)
  }

  return (
    <AppLabels
      labels={labels}
      defaultLabels={app?.default_labels ?? {}}
      isLoading={isLoading}
      isPending={isPending}
      resetAction={
        hasOverrides ? (
          <Button variant="ghost" onClick={handleResetAll} disabled={isPending}>
            <Icon variant="ArrowCounterClockwiseIcon" size={16} />
            Reset all to defaults
          </Button>
        ) : null
      }
      onOverride={handleOverride}
      onRemoveOverride={handleRemoveOverride}
    />
  )
}

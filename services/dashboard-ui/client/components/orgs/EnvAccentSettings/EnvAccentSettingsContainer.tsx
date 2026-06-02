import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { updateEnvAccentConfig } from '@/lib'
import type { TAPIError } from '@/types'
import { EnvAccentSettings } from './EnvAccentSettings'

export const EnvAccentSettingsContainer = () => {
  const { org, refresh } = useOrg()
  const queryClient = useQueryClient()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: updateEnvAccentConfig,
    onSuccess: () => {
      addToast(<Toast heading="Environment accents updated" theme="success" />)
      refresh()
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      queryClient.invalidateQueries({ queryKey: ['app-installs'] })
      queryClient.invalidateQueries({ queryKey: ['install', org.id] })
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to update env accents" theme="error">
          <Text>{err.error}</Text>
        </Toast>,
      )
    },
  })

  return (
    <EnvAccentSettings
      config={org.env_accent_config ?? { label_key: 'env', values: {} }}
      isPending={isPending}
      error={error}
      onSubmit={(body) => mutate({ orgId: org.id, body })}
    />
  )
}

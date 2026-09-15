import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useConfig } from '@/hooks/use-config'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getInstallStack } from '@/lib/ctl-api/installs/get-install-stack'
import { getInstallTelemetrySettings } from '@/lib/ctl-api/installs/get-install-telemetry-settings'
import { updateInstallTelemetrySettings } from '@/lib/ctl-api/installs/update-install-telemetry-settings'
import { InstallTelemetry } from './InstallTelemetry'

export const InstallTelemetryContainer = () => {
  const { isByoc } = useConfig()
  const { install } = useInstall()
  const { org } = useOrg()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const orgId = org?.id
  const installId = install?.id
  const canQuery = !!isByoc && !!orgId && !!installId
  const telemetryKey = ['install-telemetry', orgId, installId]

  const stack = useQuery({
    queryKey: ['install-stack', orgId, installId],
    queryFn: () => getInstallStack({ orgId: orgId!, installId: installId! }),
    enabled: canQuery,
  })
  const settings = useQuery({
    queryKey: telemetryKey,
    queryFn: () =>
      getInstallTelemetrySettings({ orgId: orgId!, installId: installId! }),
    enabled: canQuery,
    retry: false,
  })
  const endpoint =
    stack.data?.install_stack_outputs?.data_contents?.telemetry_endpoint
  const hasSetup =
    !!install?.runner_id && typeof endpoint === 'string' && !!endpoint.trim()

  const mutation = useMutation({
    mutationFn: (variables: {
      orgId: string
      installId: string
      runnerId?: string
      enabled: boolean
    }) => updateInstallTelemetrySettings(variables),
    onSuccess: (data, variables) => {
      const updatedKey = [
        'install-telemetry',
        variables.orgId,
        variables.installId,
      ]
      queryClient.setQueryData(updatedKey, data)
      queryClient.invalidateQueries({ queryKey: updatedKey })
      queryClient.invalidateQueries({
        queryKey: ['runner', variables.orgId, variables.runnerId],
      })
      addToast(
        <Toast
          heading={data.enabled ? 'Telemetry enabled' : 'Telemetry disabled'}
          theme="success"
        >
          <Text>The runner will apply this setting on its next refresh.</Text>
        </Toast>
      )
    },
    onError: (error) => {
      addToast(
        <Toast heading="Telemetry update failed" theme="error">
          <Text>
            {error.description ||
              error.error ||
              'Unable to update telemetry settings. Try again.'}
          </Text>
        </Toast>
      )
    },
  })

  if (!canQuery) return null

  const needsStack = !settings.data?.enabled && !!install?.runner_id
  const isLoading = settings.isPending || (needsStack && stack.isPending)
  const error = settings.error || (needsStack ? stack.error : null)

  return (
    <InstallTelemetry
      enabled={!!settings.data?.enabled}
      hasSetup={hasSetup}
      isLoading={isLoading}
      error={error}
      isPending={mutation.isPending}
      onToggle={(enabled) => {
        if (
          !mutation.isPending &&
          !isLoading &&
          settings.data &&
          !error &&
          (!enabled || hasSetup)
        ) {
          mutation.mutate({
            enabled,
            orgId: orgId!,
            installId: installId!,
            runnerId: install?.runner_id,
          })
        }
      }}
      onRetry={() => {
        void settings.refetch()
        if (needsStack) void stack.refetch()
      }}
    />
  )
}

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getInstallStack } from '@/lib/ctl-api/installs/get-install-stack'
import { getInstallTelemetrySettings } from '@/lib/ctl-api/installs/get-install-telemetry-settings'
import { updateInstallTelemetrySettings } from '@/lib/ctl-api/installs/update-install-telemetry-settings'
import { InstallTelemetry } from './InstallTelemetry'

export const InstallTelemetryContainer = () => {
  const { install } = useInstall()
  const { org } = useOrg()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const orgId = org?.id
  const installId = install?.id
  const canQuery = !!orgId && !!installId
  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'
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
  const hasEndpoint = typeof endpoint === 'string' && !!endpoint.trim()
  const isRunnerActive =
    !!install?.runner_id && install.runner_status === 'active'

  const lastStackError = useRef<string>()
  useEffect(() => {
    if (!canQuery || !settings.data?.enabled || !stack.isError) return
    const errorKey = `${orgId}:${installId}:${stack.errorUpdatedAt}`
    if (lastStackError.current === errorKey) return
    lastStackError.current = errorKey
    addToast(
      <Toast heading="Telemetry endpoint check failed" theme="warn">
        <Text>
          {stack.error.description ||
            stack.error.error ||
            'Unable to load the install stack. Try again.'}
        </Text>
        <Button
          variant="secondary"
          className="w-fit"
          onClick={() => {
            void queryClient.invalidateQueries({
              queryKey: ['install-stack', orgId, installId],
              exact: true,
            })
          }}
        >
          Retry settings
        </Button>
      </Toast>
    )
  }, [
    canQuery,
    settings.data?.enabled,
    stack.isError,
    stack.error,
    stack.errorUpdatedAt,
    orgId,
    installId,
    addToast,
    queryClient,
  ])

  const mutation = useMutation({
    mutationFn: (variables: {
      orgId: string
      installId: string
      runnerId?: string
      enabled: boolean | null
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
      queryClient.invalidateQueries({
        queryKey: ['install', variables.orgId, variables.installId],
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

  const needsStack = !settings.data?.enabled
  const isLoading = settings.isPending || (needsStack && stack.isPending)
  const error = settings.error || (needsStack ? stack.error : null)
  const canUseOrgDefault = !settings.data?.org_default || hasEndpoint

  return (
    <InstallTelemetry
      enabled={!!settings.data?.enabled}
      hasEndpoint={stack.data ? hasEndpoint : undefined}
      isRunnerActive={isRunnerActive}
      isInherited={settings.data ? settings.data.override === null : undefined}
      isManagedByConfig={isManagedByConfig}
      canUseOrgDefault={canUseOrgDefault}
      isLoading={isLoading}
      error={error}
      isPending={mutation.isPending}
      onToggle={(enabled) => {
        if (
          !isManagedByConfig &&
          !mutation.isPending &&
          !isLoading &&
          settings.data &&
          !error &&
          (!enabled || hasEndpoint)
        ) {
          mutation.mutate({
            enabled,
            orgId: orgId!,
            installId: installId!,
            runnerId: install?.runner_id,
          })
        }
      }}
      onUseOrgDefault={() => {
        if (
          !isManagedByConfig &&
          !mutation.isPending &&
          !isLoading &&
          settings.data &&
          !error &&
          canUseOrgDefault
        ) {
          mutation.mutate({
            enabled: null,
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

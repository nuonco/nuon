import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getInstallCurrentInputs } from '@/lib'
import type { TAppInput } from '@/types'
import { normalizeAppInputGroups } from '@/utils/app-utils'

export type TConfigurationInputGroup = {
  id?: string
  name?: string
  display_name?: string
  description?: string
  app_inputs?: TAppInput[]
}

export const useConfigurationInputs = (): {
  groups: TConfigurationInputGroup[]
  isLoading: boolean
  values: Record<string, string>
} => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { appConfig, isLoading: configLoading } = useInstallAppConfig()

  const { data: inputs, isLoading: inputsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const groups = appConfig
    ? normalizeAppInputGroups(
        appConfig.input?.input_groups ?? [],
        appConfig.input?.inputs ?? []
      )
    : []

  return {
    groups,
    isLoading: configLoading || inputsLoading,
    values: inputs?.redacted_values ?? {},
  }
}

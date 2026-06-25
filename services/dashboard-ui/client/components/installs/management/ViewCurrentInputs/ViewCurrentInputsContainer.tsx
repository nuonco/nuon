import { useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  getAppConfig,
  getInstallComponents,
  getInstallCurrentInputs,
} from '@/lib'
import { normalizeAppInputGroups } from '@/utils/app-utils'
import { EditInputsButton } from '../EditInputs'
import { ViewCurrentInputsModal } from './ViewCurrentInputs'

export const ViewCurrentInputsModalContainer = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const canRenameInstall = !!org?.features?.['install-rename']

  const { data: inputs, isLoading: inputsLoading } = useQuery({
    queryKey: ['install-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const { data: config, isLoading: configLoading } = useQuery({
    queryKey: ['app-config', org?.id, install?.app_id, install?.app_config_id],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  const { data: componentsResult } = useQuery({
    queryKey: ['install-components', org?.id, install?.id],
    queryFn: () =>
      getInstallComponents({ orgId: org.id, installId: install.id, limit: 100 }),
    enabled: !!org?.id && !!install?.id,
  })

  const isLoading = inputsLoading || configLoading
  const redactedValues = inputs?.redacted_values ?? {}
  const inputGroups = config
    ? normalizeAppInputGroups(
        config.input?.input_groups ?? [],
        config.input?.inputs ?? []
      )
    : []

  const configComponents = config?.component_config_connections ?? []
  const effectiveEnabledByName: Record<string, boolean | undefined> = {}
  for (const c of componentsResult?.data ?? []) {
    if (c.component?.name) effectiveEnabledByName[c.component.name] = c.enabled
  }

  return (
    <ViewCurrentInputsModal
      isLoading={isLoading}
      redactedValues={redactedValues}
      inputGroups={inputGroups as any}
      configComponents={configComponents}
      effectiveEnabledByName={effectiveEnabledByName}
      footerActions={
        <EditInputsButton variant="primary" showNameField={canRenameInstall} />
      }
      {...props}
    />
  )
}

export const ViewCurrentInputsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      variant="ghost"
      onClick={() => {
        const modal = <ViewCurrentInputsModalContainer />
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="ListChecksIcon" />}
      Current inputs
      {props?.isMenuButton ? <Icon variant="ListChecksIcon" /> : null}
    </Button>
  )
}

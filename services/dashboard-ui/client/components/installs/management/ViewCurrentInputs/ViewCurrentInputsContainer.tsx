import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallCurrentInputs } from '@/lib'
import { normalizeAppInputGroups } from '@/utils/app-utils'
import { EditInputsButton } from '../EditInputs'
import { ViewCurrentInputsModal } from './ViewCurrentInputs'

export const ViewCurrentInputsModalContainer = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const canRenameInstall = !!org?.features?.['install-rename']

  const { data: inputs, isLoading: inputsLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const { appConfig: config, isLoading: configLoading } = useInstallAppConfig()

  const isLoading = inputsLoading || configLoading
  const redactedValues = inputs?.redacted_values ?? {}
  const inputGroups = config
    ? normalizeAppInputGroups(
        config.input?.input_groups ?? [],
        config.input?.inputs ?? []
      )
    : []

  return (
    <ViewCurrentInputsModal
      isLoading={isLoading}
      redactedValues={redactedValues}
      inputGroups={inputGroups as any}
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

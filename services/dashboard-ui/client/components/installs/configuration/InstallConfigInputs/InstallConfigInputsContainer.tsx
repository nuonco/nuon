import { EditInputsButton } from '@/components/installs/management/EditInputs'
import { useConfigurationInputs } from '@/components/installs/configuration/shared/use-configuration-inputs'
import { COMPONENT_OVERRIDE_INPUT_GROUP } from '@/utils/install-utils'
import { InstallConfigInputs } from './InstallConfigInputs'

export const InstallConfigInputsContainer = () => {
  const { groups, isLoading, values } = useConfigurationInputs()

  return (
    <InstallConfigInputs
      action={<EditInputsButton variant="secondary" />}
      groups={groups.filter(
        (group) => group.name !== COMPONENT_OVERRIDE_INPUT_GROUP
      )}
      isLoading={isLoading}
      values={values}
    />
  )
}

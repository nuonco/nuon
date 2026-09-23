import { EditInputsButton } from '@/components/installs/management/EditInputs'
import { useConfigurationInputs } from '@/components/installs/configuration/shared/use-configuration-inputs'
import { useInstall } from '@/hooks/use-install'
import type { TComponentType } from '@/types'
import { COMPONENT_OVERRIDE_INPUT_GROUP } from '@/utils/install-utils'
import { InstallConfigOverrides } from './InstallConfigOverrides'

export const InstallConfigOverridesContainer = () => {
  const { install } = useInstall()
  const { groups, isLoading, values } = useConfigurationInputs()
  const overrideGroup = groups.find(
    (group) => group.name === COMPONENT_OVERRIDE_INPUT_GROUP
  )

  // Override input names encode only the override kind, so enabled-only
  // components have no type until it is looked up here.
  const componentTypes = Object.fromEntries(
    (install?.install_components ?? []).flatMap((installComponent) => {
      const { name, type } = installComponent.component ?? {}
      return name && type ? [[name, type as TComponentType]] : []
    })
  )

  return (
    <InstallConfigOverrides
      action={<EditInputsButton variant="secondary" />}
      componentTypes={componentTypes}
      inputs={overrideGroup?.app_inputs}
      isLoading={isLoading}
      values={values}
    />
  )
}

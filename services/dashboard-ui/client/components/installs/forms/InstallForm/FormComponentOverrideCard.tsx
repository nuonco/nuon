import { useStore } from '@tanstack/react-form'
import { Badge } from '@/components/common/Badge'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import { FormCodeInput } from '@/components/common/form/FormCodeInput'
import { FormToggle } from '@/components/common/form/FormToggle'
import {
  type TComponentOverrideCard,
  type TComponentType,
} from '@/utils/install-utils'
import type { TAppInput } from '@/types'
import type { InstallInputFieldName } from './schema'
import type { InstallFormApi } from './useInstallForm'

const COMPONENT_TYPE_LABELS: Record<TComponentType, string> = {
  terraform_module: 'Terraform module',
  helm_chart: 'Helm chart',
}

const CONFIG_KIND: Record<
  'tf_vars' | 'helm_values',
  { label: string; language: 'hcl' | 'yaml'; placeholder: string }
> = {
  tf_vars: {
    label: 'Terraform variables',
    language: 'hcl',
    placeholder: 'Enter HCL (.tfvars) configuration...',
  },
  helm_values: {
    label: 'Helm values',
    language: 'yaml',
    placeholder: 'Enter YAML configuration...',
  },
}

interface IFormComponentOverrideCard {
  form: InstallFormApi
  card: TComponentOverrideCard<TAppInput>
  disabled?: boolean
}

export const FormComponentOverrideCard = ({
  form,
  card,
  disabled,
}: IFormComponentOverrideCard) => {
  const { enabledInput, configInput, configKind } = card
  const config = configKind ? CONFIG_KIND[configKind] : undefined

  const enabledFieldName: InstallInputFieldName | undefined = enabledInput?.name
    ? `inputs.${enabledInput.name}`
    : undefined
  const enabled = useStore(form.store, (s) =>
    enabledFieldName
      ? ((s.values.inputs?.[enabledInput!.name!] as boolean | undefined) ??
        true)
      : true
  )

  return (
    <div className="flex flex-col gap-3 border rounded-md p-4">
      <div className="flex items-center justify-between gap-4">
        <span className="flex items-center gap-2">
          <Text variant="body" weight="strong">
            {card.component}
          </Text>
          {card.componentType && (
            <Badge size="sm" theme="neutral">
              {COMPONENT_TYPE_LABELS[card.componentType]}
            </Badge>
          )}
        </span>

        {enabledInput ? (
          <form.Field name={enabledFieldName!}>
            {(field) => (
              <FormToggle
                field={field}
                label={field.state.value ? 'Enabled' : 'Disabled'}
                disabled={disabled}
              />
            )}
          </form.Field>
        ) : (
          <Text variant="subtext" theme="neutral">
            Always deployed
          </Text>
        )}
      </div>

      {enabledInput && !enabled && (
        <Text variant="subtext" theme="neutral">
          Disabled — won't be deployed on this install.
        </Text>
      )}

      {configInput?.name && config && (
        <Expand
          id={`override-${card.component}-config`}
          heading={
            <Text variant="subtext" weight="strong">
              {config.label}{' '}
              <Text variant="subtext" theme="neutral">
                (optional)
              </Text>
            </Text>
          }
          headerClassName="!px-4 bg-code"
          className="border rounded-md"
        >
          <div className="p-4 border-t">
            <form.Field
              name={`inputs.${configInput.name}` as InstallInputFieldName}
            >
              {(field) => (
                <FormCodeInput
                  field={field}
                  language={config.language}
                  placeholder={config.placeholder}
                  minHeight={120}
                  disabled={disabled || !enabled}
                />
              )}
            </form.Field>
          </div>
        </Expand>
      )}
    </div>
  )
}

import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { CodeInput } from '@/components/common/form/CodeInput'
import { Toggle } from '@/components/common/form/Toggle'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import {
  groupComponentOverrideInputs,
  type TComponentOverrideCard,
  type TComponentType,
} from '@/utils/install-utils'
import type { TAppInput, TAppInputConfig, TInstall } from '@/types'

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

const ComponentOverrideCard = ({
  card,
  mergedValues,
}: {
  card: TComponentOverrideCard<TAppInput>
  mergedValues?: Record<string, string>
}) => {
  const { enabledInput, configInput, configKind } = card
  const toggleable = !!enabledInput

  const initialEnabled = enabledInput
    ? (mergedValues?.[enabledInput.name || ''] ?? enabledInput.default) ===
      'true'
    : true
  const [enabled, setEnabled] = useState(initialEnabled)

  const config = configKind ? CONFIG_KIND[configKind] : undefined

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

        {toggleable ? (
          <>
            <input
              type="hidden"
              name={`inputs:${enabledInput!.name}`}
              value={enabled ? 'true' : 'false'}
            />
            <Toggle
              checked={enabled}
              onChange={setEnabled}
              label={enabled ? 'Enabled' : 'Disabled'}
            />
          </>
        ) : (
          <Text variant="subtext" theme="neutral">
            Always deployed
          </Text>
        )}
      </div>

      {toggleable && !enabled && (
        <Text variant="subtext" theme="neutral">
          Disabled — won't be deployed on this install.
        </Text>
      )}

      {configInput && config && (
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
            <CodeInput
              language={config.language}
              name={`inputs:${configInput.name}`}
              defaultValue={
                mergedValues?.[configInput.name || ''] ?? configInput.default
              }
              placeholder={config.placeholder}
              minHeight={120}
              disabled={!enabled}
            />
          </div>
        </Expand>
      )}
    </div>
  )
}

export const ComponentOverridesSection = ({
  group,
  install,
  draftValues,
}: {
  group: NonNullable<TAppInputConfig['input_groups']>[number]
  install?: TInstall
  draftValues?: Record<string, string> | null
}) => {
  const installInputs = install ? install?.install_inputs?.at(0)?.values : {}

  const hasDraftValues = draftValues && Object.keys(draftValues).length > 0
  const normalizedDraftValues: Record<string, string> = {}
  if (hasDraftValues) {
    Object.entries(draftValues!).forEach(([key, value]) => {
      if (key.startsWith('inputs:')) {
        normalizedDraftValues[key.replace('inputs:', '')] = value
      }
    })
  }
  const mergedValues = hasDraftValues
    ? { ...installInputs, ...normalizedDraftValues }
    : installInputs

  const cards = groupComponentOverrideInputs(group?.app_inputs || [])
  if (cards.length === 0) {
    return null
  }

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="flex flex-col gap-0 pr-6">
        <span className="text-lg font-semibold">Components</span>
        <span className="text-sm font-normal">
          Choose which components deploy on this install and customize each one.
        </span>
      </legend>

      <div className="flex flex-col gap-4">
        {cards.map((card) => (
          <ComponentOverrideCard
            key={card.component}
            card={card}
            mergedValues={mergedValues}
          />
        ))}
      </div>
    </fieldset>
  )
}

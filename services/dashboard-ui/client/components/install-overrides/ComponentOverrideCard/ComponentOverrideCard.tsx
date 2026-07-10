import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { CodeInput } from '@/components/common/form/CodeInput'
import { Toggle } from '@/components/common/form/Toggle'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import {
  type TComponentOverrideCard,
  type TComponentType,
} from '@/utils/install-utils'
import type { TAppInput } from '@/types'

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

export interface IComponentOverrideCard {
  card: TComponentOverrideCard<TAppInput>
  values?: Record<string, string>
  readOnly?: boolean
  showEnabled?: boolean
}

export const ComponentOverrideCard = ({
  card,
  values,
  readOnly = false,
  showEnabled = true,
}: IComponentOverrideCard) => {
  const { enabledInput, configInput, configKind } = card
  const toggleable = !!enabledInput

  const initialEnabled = enabledInput
    ? (values?.[enabledInput.name || ''] ?? enabledInput.default) === 'true'
    : true
  const [enabled, setEnabled] = useState(initialEnabled)

  const config = configKind ? CONFIG_KIND[configKind] : undefined
  const configValue =
    configInput?.name != null
      ? (values?.[configInput.name] ?? configInput?.default)
      : undefined

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

        {showEnabled &&
          (toggleable ? (
            readOnly ? (
              <Badge size="sm" theme={initialEnabled ? 'success' : 'neutral'}>
                {initialEnabled ? 'Enabled' : 'Disabled'}
              </Badge>
            ) : (
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
            )
          ) : (
            <Text variant="subtext" theme="neutral">
              Always deployed
            </Text>
          ))}
      </div>

      {showEnabled && toggleable && !readOnly && !enabled && (
        <Text variant="subtext" theme="neutral">
          Disabled — won't be deployed on this install.
        </Text>
      )}

      {configInput &&
        config &&
        (readOnly ? (
          configValue ? (
            <Expand
              id={`override-${card.component}-config`}
              isOpen
              heading={
                <Text variant="subtext" weight="strong">
                  {config.label}
                </Text>
              }
              headerClassName="!px-4 bg-code"
              className="border rounded-md"
            >
              <CodeBlock
                className="!text-xs w-full !max-h-64 border-t"
                language={config.language}
                showCopy
              >
                {String(configValue).replace(/\n+$/, '')}
              </CodeBlock>
            </Expand>
          ) : (
            <Text variant="subtext" theme="neutral">
              No {config.label.toLowerCase()} override set.
            </Text>
          )
        ) : (
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
                defaultValue={configValue}
                placeholder={config.placeholder}
                minHeight={120}
                disabled={!enabled}
              />
            </div>
          </Expand>
        ))}
    </div>
  )
}

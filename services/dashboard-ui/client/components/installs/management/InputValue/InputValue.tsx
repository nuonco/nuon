import { CodeBlock } from '@/components/common/CodeBlock'
import { Text } from '@/components/common/Text'
import {
  getComponentOverrideKind,
  type TComponentOverrideKind,
} from '@/utils/install-utils'

const KIND_LANGUAGE: Record<TComponentOverrideKind, 'yaml' | 'hcl'> = {
  helm_values: 'yaml',
  tf_vars: 'hcl',
}

interface IInputValue {
  name?: string
  value?: string | null
}

// InputValue renders a single install input value. Component-override inputs
// (Helm values / Terraform vars) carry multi-line YAML/HCL, so they render in a
// syntax-highlighted code block instead of a cramped inline string.
export const InputValue = ({ name, value }: IInputValue) => {
  if (value == null) {
    return (
      <Text variant="subtext" family="mono" theme="neutral">
        —
      </Text>
    )
  }

  const str = String(value)
  if (str === '') {
    return (
      <Text variant="subtext" family="mono" theme="neutral">
        &quot;&quot;
      </Text>
    )
  }

  const kind = name ? getComponentOverrideKind(name) : null
  if (kind) {
    return (
      <CodeBlock
        className="!text-xs w-full"
        language={KIND_LANGUAGE[kind]}
        showCopy
      >
        {str.replace(/\n+$/, '')}
      </CodeBlock>
    )
  }

  return (
    <Text variant="subtext" family="mono" weight="strong">
      {str}
    </Text>
  )
}

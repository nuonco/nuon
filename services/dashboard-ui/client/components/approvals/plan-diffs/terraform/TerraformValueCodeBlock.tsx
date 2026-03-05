import { useMemo } from 'react'
import { CodeBlock } from '@/components/common/CodeBlock'
import { detectValueFormat } from '@/utils/terraform-utils'

export const TerraformValueCodeBlock = ({ value }: { value: string }) => {
  const { displayValue, language, showLineNumbers } = useMemo(
    () => detectValueFormat(value),
    [value]
  )

  return (
    <CodeBlock language={language} showLineNumbers={showLineNumbers}>
      {displayValue}
    </CodeBlock>
  )
}

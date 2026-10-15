import { CodeBlock } from '@/components/common/CodeBlock'
import { Text } from '@/components/common/Text'
import type { IAdhocActionDetails } from './types'

export const AdhocActionDetails = ({ actionRun }: IAdhocActionDetails) => {
  const adhocConfig = actionRun?.steps?.at(0)?.adhoc_config

  return (
    <>
      {adhocConfig?.command ? (
        <div className="flex flex-col gap-2">
          <Text weight="strong">Command</Text>
          <CodeBlock language="bash">{adhocConfig.command}</CodeBlock>
        </div>
      ) : null}

      {adhocConfig?.inline_contents ? (
        <div className="flex flex-col gap-2">
          <Text weight="strong">Inline contents</Text>
          <CodeBlock language="bash">{adhocConfig.inline_contents}</CodeBlock>
        </div>
      ) : null}
    </>
  )
}

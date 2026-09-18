import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

export const StepErrorLog = ({
  text,
  logsHref,
}: {
  text: string
  logsHref?: string
}) => (
  <Banner theme="error">
    <div className="flex w-full min-w-0 flex-col gap-2">
      <div className="flex items-center justify-between gap-4">
        <Text weight="strong">Last error log</Text>
        {logsHref ? <Link href={logsHref}>View logs</Link> : null}
      </div>
      <CodeBlock
        language="text"
        showCopy
        className="!whitespace-pre-wrap !break-all !pr-12 [&_code]:!whitespace-pre-wrap [&_code]:!break-all"
      >
        {text}
      </CodeBlock>
    </div>
  </Banner>
)

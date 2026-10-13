import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

interface IConfigParseFailure {
  lines: string[]
  className?: string
  headerAction?: ReactNode
}

export const ConfigParseFailure = ({
  lines,
  className,
  headerAction,
}: IConfigParseFailure) => (
  <Banner theme="error" className={cn('basis-full', className)}>
    <div className="flex min-w-0 flex-col gap-2">
      <div className="flex items-start justify-between gap-3">
        <Text weight="strong">This app config could not be parsed.</Text>
        {headerAction}
      </div>
      {lines.length ? (
        <CodeBlock
          language="text"
          showCopy
          className="!whitespace-pre-wrap !break-all !pr-12 [&_code]:!whitespace-pre-wrap [&_code]:!break-all"
        >
          {lines.join('\n')}
        </CodeBlock>
      ) : null}
    </div>
  </Banner>
)

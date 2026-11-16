import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

export const StepErrorLog = ({
  text,
  logsHref,
  heading = 'Last error log',
  linkLabel = 'View logs',
  onLinkClick,
  children,
}: {
  text: string
  logsHref?: string
  heading?: string
  linkLabel?: string
  onLinkClick?: () => void
  children?: ReactNode
}) => (
  <Banner theme="error">
    <div className="flex w-full min-w-0 flex-col gap-2">
      <div className="flex items-center justify-between gap-4">
        <Text weight="strong">{heading}</Text>
        {logsHref ? (
          <Link href={logsHref} onClick={onLinkClick}>
            {linkLabel}
          </Link>
        ) : null}
      </div>
      {children ? <Text>{children}</Text> : null}
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

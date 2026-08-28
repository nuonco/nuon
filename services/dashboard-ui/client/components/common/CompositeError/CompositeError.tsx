import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import type {
  TCompositeError,
  TCompositeErrorSection,
  TCompositeErrorSeverity,
  TTheme,
} from '@/types'

interface ICompositeError {
  error: TCompositeError
}

const SEVERITY_THEME: Record<TCompositeErrorSeverity, TTheme> = {
  fatal: 'error',
  error: 'error',
  warning: 'warn',
  info: 'info',
}

const SectionBody = ({ section }: { section: TCompositeErrorSection }) => {
  const body = section?.body ?? ''
  if (!body.trim()) {
    return null
  }

  switch (section?.kind) {
    case 'code':
      return (
        <CodeBlock
          language="text"
          showCopy
          className="!whitespace-pre-wrap !break-all !pr-12 [&_code]:!whitespace-pre-wrap [&_code]:!break-all"
        >
          {body}
        </CodeBlock>
      )
    case 'text':
      return (
        <Text variant="subtext" className="whitespace-pre-wrap break-words">
          {body}
        </Text>
      )
    default:
      return <Markdown content={body} variant="compact" />
  }
}

export const CompositeError = ({ error }: ICompositeError) => {
  const theme = SEVERITY_THEME[error?.severity] ?? 'error'
  const sections = Array.isArray(error?.sections) ? error.sections : []

  const hasContent =
    Boolean(error?.message?.trim()) ||
    Boolean(error?.type?.trim()) ||
    sections.some(
      (section) =>
        Boolean(section?.heading?.trim()) || Boolean(section?.body?.trim())
    )

  if (!hasContent) {
    return null
  }

  return (
    <Banner theme={theme}>
      <div className="flex w-full min-w-0 flex-col gap-2">
        <div className="flex flex-wrap items-center gap-2 mb-2">
          <Text variant="base" weight="strong">
            {error?.message}
          </Text>
          {error?.type ? (
            <Badge variant="code" size="sm" theme="neutral">
              {error.type}
            </Badge>
          ) : null}
        </div>

        {sections.map((section, i) => (
          <div key={i} className="flex min-w-0 flex-col gap-1">
            {section?.heading ? (
              <Text weight="strong">{section.heading}</Text>
            ) : null}
            <SectionBody section={section} />
          </div>
        ))}
      </div>
    </Banner>
  )
}

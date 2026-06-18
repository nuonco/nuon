import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import type { TCompositeError, TCompositeErrorSeverity, TTheme } from '@/types'

interface ICompositeError {
  error: TCompositeError
}

const SEVERITY_THEME: Record<TCompositeErrorSeverity, TTheme> = {
  fatal: 'error',
  error: 'error',
  warning: 'warn',
  info: 'info',
}

export const CompositeError = ({ error }: ICompositeError) => {
  const theme = SEVERITY_THEME[error?.severity] ?? 'error'
  const sections = Array.isArray(error?.sections) ? error.sections : []

  return (
    <Banner theme={theme}>
      <div className="flex w-full min-w-0 flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <Text weight="strong">{error?.message}</Text>
          {error?.type ? (
            <Badge variant="code" size="sm" theme={theme}>
              {error.type}
            </Badge>
          ) : null}
        </div>

        {sections.map((section, i) => (
          <div key={i} className="flex min-w-0 flex-col gap-1">
            {section?.heading ? (
              <Text variant="subtext" weight="strong">
                {section.heading}
              </Text>
            ) : null}
            <Markdown content={section?.body} />
          </div>
        ))}
      </div>
    </Banner>
  )
}

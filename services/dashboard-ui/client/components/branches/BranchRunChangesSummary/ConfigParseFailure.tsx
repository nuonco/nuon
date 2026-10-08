import type { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

interface IConfigParseFailure {
  title: string
  lines: string[]
  className?: string
  headerAction?: ReactNode
}

export const ConfigParseFailure = ({
  title,
  lines,
  className,
  headerAction,
}: IConfigParseFailure) => (
  <section
    className={cn(
      'border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden min-w-0',
      className
    )}
  >
    <header className="flex items-center justify-between gap-3 px-5 py-4">
      <Text variant="h3" weight="strong">
        {title}
      </Text>
      {headerAction}
    </header>
    <div className="flex flex-col gap-3 border-t px-5 py-4">
      <Text variant="subtext" weight="strong">
        This app config could not be parsed.
      </Text>
      {lines.length ? (
        <ul className="flex flex-col gap-2">
          {lines.map((line, index) => (
            <li key={`${index}-${line}`}>
              <Text
                as="p"
                variant="subtext"
                family="mono"
                className="whitespace-pre-wrap break-words"
              >
                {line}
              </Text>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  </section>
)

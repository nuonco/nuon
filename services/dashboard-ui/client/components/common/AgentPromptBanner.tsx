import { useId, useState } from 'react'
import { AnimatedHeight } from '@/components/common/AnimatedHeight'
import { Button } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface IAgentPromptBanner {
  buttonText?: string
  heading: string
  description: string
  prompt: string
  defaultExpanded?: boolean
  className?: string
}

export const AgentPromptBanner = ({
  buttonText = 'Ask your agent',
  heading,
  description,
  prompt,
  defaultExpanded = false,
  className,
}: IAgentPromptBanner) => {
  const [expanded, setExpanded] = useState(defaultExpanded)
  const contentId = useId()

  return (
    <div
      className={cn(
        'flex w-full flex-col rounded-lg border p-3',
        'bg-cool-grey-50 text-cool-grey-800 !border-cool-grey-300',
        'dark:bg-dark-grey-600 dark:!border-cool-grey-600/40 dark:text-cool-grey-400',
        className,
      )}
    >
      <div className="flex min-w-0 items-center gap-3">
        <Button
          type="button"
          variant="secondary"
          className="shrink-0"
          aria-expanded={expanded}
          aria-controls={contentId}
          onClick={() => setExpanded((open) => !open)}
        >
          <Icon variant="SparkleIcon" />
          {buttonText}
        </Button>
        <Text as="p" weight="strong" className="min-w-0">
          {heading}
        </Text>
      </div>
      <AnimatedHeight>
        <div className={expanded ? undefined : 'h-0'}>
          <div
            id={contentId}
            aria-hidden={!expanded}
            inert={!expanded}
            className="flex flex-col gap-2 pt-3"
          >
            <Text as="p" variant="subtext">
              {description}
            </Text>
            <CodeBlock language="text" showCopy wrapLongLines>
              {prompt}
            </CodeBlock>
          </div>
        </div>
      </AnimatedHeight>
    </div>
  )
}

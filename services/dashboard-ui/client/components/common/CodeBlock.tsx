import { lazy, Suspense } from 'react'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { cn } from '@/utils/classnames'

const PrismCodeBlock = lazy(() =>
  import('./PrismCodeBlock').then((m) => ({ default: m.PrismCodeBlock }))
)

interface ICodeBlock
  extends Omit<React.HTMLAttributes<HTMLPreElement>, 'children'> {
  children: string
  language:
    | 'json'
    | 'yaml'
    | 'yml'
    | 'hcl'
    | 'sh'
    | 'bash'
    | 'toml'
    | 'markdown'
    | 'md'
    | string
  isDiff?: boolean
  showCopy?: boolean
  showLineNumbers?: boolean
}

function CodeBlockFallback({ children, className }: { children: string; className?: string }) {
  return (
    <pre
      className={cn(
        '!m-0 !p-4 !text-sm !rounded-md !shadow-sm min-h-[3rem] max-h-[40rem] overflow-auto bg-code',
        className,
      )}
    >
      <code className="bg-code font-mono w-full">{children}</code>
    </pre>
  )
}

export function CodeBlock({
  className,
  children,
  language,
  isDiff = false,
  showCopy = false,
  showLineNumbers = false,
}: ICodeBlock) {
  const prism = (
    <Suspense fallback={<CodeBlockFallback className={className}>{children}</CodeBlockFallback>}>
      <PrismCodeBlock
        className={className}
        language={language}
        isDiff={isDiff}
        showLineNumbers={showLineNumbers}
      >
        {children}
      </PrismCodeBlock>
    </Suspense>
  )

  if (!showCopy) return prism

  return (
    <div className="relative">
      <div className="absolute top-2 right-2 z-10">
        <ClickToCopyButton textToCopy={children} />
      </div>
      {prism}
    </div>
  )
}

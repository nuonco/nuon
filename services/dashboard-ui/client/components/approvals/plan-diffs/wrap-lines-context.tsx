import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'

type TWrapLinesContext = {
  wrapLines: boolean
  toggleWrapLines: () => void
}

const WrapLinesContext = createContext<TWrapLinesContext | null>(null)

export function useWrapLines(): boolean {
  return useContext(WrapLinesContext)?.wrapLines ?? false
}

export function WrapLinesProvider({ children }: { children: ReactNode }) {
  const [wrapLines, setWrapLines] = useState(false)
  const value = useMemo(
    () => ({ wrapLines, toggleWrapLines: () => setWrapLines((v) => !v) }),
    [wrapLines]
  )
  return (
    <WrapLinesContext.Provider value={value}>
      {children}
    </WrapLinesContext.Provider>
  )
}

export function WrapLinesToggle() {
  const ctx = useContext(WrapLinesContext)
  if (!ctx) return null
  return (
    <Button
      className="!p-1 flex items-center gap-1.5"
      variant="ghost"
      size="sm"
      aria-pressed={ctx.wrapLines}
      onClick={ctx.toggleWrapLines}
    >
      {ctx.wrapLines ? 'Unwrap lines' : 'Wrap lines'}
      <Icon variant="ArrowElbowDownLeftIcon" size="14" />
    </Button>
  )
}

export function DiffCodeBlock(props: React.ComponentProps<typeof CodeBlock>) {
  const wrapLines = useWrapLines()
  return <CodeBlock {...props} wrapLongLines={wrapLines} />
}

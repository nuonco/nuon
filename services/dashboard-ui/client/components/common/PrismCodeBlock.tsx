import { cloneElement, isValidElement } from 'react'
import { Prism } from 'react-syntax-highlighter'
import {
  oneDark,
  oneLight,
} from 'react-syntax-highlighter/dist/esm/styles/prism'
import createElement from 'react-syntax-highlighter/dist/esm/create-element'
import { useSystemTheme } from '@/hooks/use-system-theme'
import { cn } from '@/utils/classnames'
import { Expand } from './Expand'

const COLLAPSE_THRESHOLD = 15
const CONTEXT_LINES = 3

type DiffOp = 'add' | 'remove' | 'change'

const MARKER_TEXT: Record<DiffOp, string> = {
  add: 'text-green-800 dark:text-green-500',
  remove: 'text-red-800 dark:text-red-500',
  change: 'text-orange-800 dark:text-orange-400',
}

function lineOp(line: string): DiffOp | null {
  if (line.startsWith('+')) return 'add'
  if (line.startsWith('-')) return 'remove'
  if (line.startsWith('~')) return 'change'
  return null
}

function colorFirstChar(
  node: any,
  colorClass: string,
  state: { done: boolean }
): any {
  if (state.done) return node
  if (typeof node === 'string') {
    if (node.length === 0) return node
    state.done = true
    return [
      <span key="diff-marker" className={cn('font-bold', colorClass)}>
        {node[0]}
      </span>,
      node.slice(1),
    ]
  }
  if (Array.isArray(node)) {
    return node.map((n) => colorFirstChar(n, colorClass, state))
  }
  if (isValidElement(node)) {
    const el = node as any
    return cloneElement(
      el,
      { ...el.props },
      colorFirstChar(el.props.children, colorClass, state)
    )
  }
  return node
}

const DIFF_CLASSES = {
  added:
    'bg-[#F4FBF7] text-green-800 !border-green-400 dark:bg-[#0C1B14] dark:!border-green-500/40 dark:text-green-500 block w-full',
  removed:
    'bg-[#FEF2F2] text-red-800 !border-red-300 dark:bg-[#290C0D] dark:!border-red-500/40 dark:text-red-500 block w-full',
  changed:
    'bg-[#FFF8F0] text-orange-800 !border-orange-300 dark:bg-[#1A1408] dark:!border-orange-500/40 dark:text-orange-400 block w-full',
  afterApply: '!italic opacity-70',
}

function renderChangedLine(line: string) {
  const arrowIdx = line.indexOf(' -> ')
  if (arrowIdx === -1) return line

  const beforeArrow = line.substring(0, arrowIdx)
  const newVal = line.substring(arrowIdx + 4)

  const colonIdx = beforeArrow.indexOf(':')
  if (colonIdx === -1) return line

  const key = beforeArrow.substring(0, colonIdx + 1)
  const oldVal = beforeArrow.substring(colonIdx + 1).trimStart()

  return (
    <>
      <span className={cn('font-bold', MARKER_TEXT.change)}>{key[0]}</span>
      {key.slice(1)}{' '}
      <span className="line-through opacity-70 text-red-800 dark:text-red-400">
        {oldVal}
      </span>
      <span className="opacity-50">{' -> '}</span>
      {newVal}
    </>
  )
}

function renderDiffRow(
  row: any,
  i: number,
  lines: string[],
  stylesheet: any,
  useInlineStyles: boolean
) {
  const line = lines[i] || ''
  const defaultEl = createElement({
    node: row,
    stylesheet,
    useInlineStyles,
    key: `line-${i}`,
  }) as any

  const op = lineOp(line)
  if (!op) return defaultEl

  const children = Array.isArray(defaultEl.props.children)
    ? defaultEl.props.children
    : [defaultEl.props.children]

  const isLineNumber = (child: any) =>
    child?.props?.className?.includes('linenumber')

  const lineNumberChild = children.find(isLineNumber)
  const contentChildren = children.filter((c: any) => !isLineNumber(c))

  const content =
    op === 'change' && line.includes(' -> ')
      ? renderChangedLine(line)
      : colorFirstChar(contentChildren, MARKER_TEXT[op], { done: false })

  const newChildren = lineNumberChild ? [lineNumberChild, content] : [content]

  return {
    ...defaultEl,
    props: { ...defaultEl.props, children: newChildren },
    key: `line-${i}`,
  }
}

function DiffCollapsedLines({
  id,
  count,
  wrapLongLines,
  children,
}: {
  id: string
  count: number
  wrapLongLines?: boolean
  children: React.ReactNode
}) {
  return (
    <Expand
      id={id}
      isIconBeforeHeading
      hasNoHoverStyle
      className="border rounded-md my-1.5 overflow-hidden"
      headerClassName="!py-1 !px-2 font-sans text-xs bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 transition-colors"
      heading={
        <span className="opacity-70">
          {count} unmodified {count === 1 ? 'line' : 'lines'}
        </span>
      }
    >
      <div
        className={
          wrapLongLines ? 'whitespace-pre-wrap break-all' : 'whitespace-pre'
        }
      >
        {children}
      </div>
    </Expand>
  )
}

function collapseUnchangedRuns(
  renderedRows: any[],
  lines: string[],
  wrapLongLines?: boolean
) {
  const nodes: React.ReactNode[] = []
  const n = lines.length
  let i = 0

  while (i < n) {
    if (lineOp(lines[i]) !== null) {
      nodes.push(renderedRows[i])
      i++
      continue
    }

    let end = i
    while (end < n && lineOp(lines[end]) === null) end++

    const runLen = end - i
    const top = i > 0 ? CONTEXT_LINES : 0
    const bottom = end < n ? CONTEXT_LINES : 0
    const hiddenStart = i + top
    const hiddenEnd = end - bottom
    const hiddenCount = hiddenEnd - hiddenStart

    if (runLen >= COLLAPSE_THRESHOLD && hiddenCount > 0) {
      for (let k = i; k < hiddenStart; k++) nodes.push(renderedRows[k])
      nodes.push(
        <DiffCollapsedLines
          key={`collapse-${i}`}
          id={`collapse-${i}`}
          count={hiddenCount}
          wrapLongLines={wrapLongLines}
        >
          {renderedRows.slice(hiddenStart, hiddenEnd)}
        </DiffCollapsedLines>
      )
      for (let k = hiddenEnd; k < end; k++) nodes.push(renderedRows[k])
    } else {
      for (let k = i; k < end; k++) nodes.push(renderedRows[k])
    }

    i = end
  }

  return nodes
}

interface IPrismCodeBlock
  extends Omit<React.HTMLAttributes<HTMLPreElement>, 'children'> {
  children: string
  language: string
  isDiff?: boolean
  showLineNumbers?: boolean
  collapseUnchanged?: boolean
  wrapLongLines?: boolean
}

export function PrismCodeBlock({
  className,
  children,
  language,
  isDiff = false,
  showLineNumbers = false,
  collapseUnchanged = true,
  wrapLongLines = false,
}: IPrismCodeBlock) {
  const colorScheme = useSystemTheme()
  const bgCode =
    colorScheme === 'dark'
      ? 'var(--color-dark-grey-800)'
      : 'var(--color-cool-grey-100)'
  const baseTheme = colorScheme === 'dark' ? oneDark : oneLight
  const theme = {
    ...baseTheme,
    'pre[class*="language-"]': {
      ...baseTheme['pre[class*="language-"]'],
      background: bgCode,
    },
    'code[class*="language-"]': {
      ...baseTheme['code[class*="language-"]'],
      background: bgCode,
    },
  }
  const lines = isDiff ? children.split('\n') : []

  return (
    <Prism
      className={cn(
        '!m-0 !p-4 !text-sm !rounded-md !shadow-sm min-h-[3rem] max-h-[40rem] overflow-auto',
        className
      )}
      language={language}
      style={theme}
      wrapLines
      showLineNumbers={showLineNumbers || isDiff}
      lineProps={(lineNumber: number) => {
        if (typeof lineNumber !== 'number') return {}
        const line = isDiff ? lines[lineNumber - 1] || '' : ''
        let className = ''

        if (isDiff) {
          if (line.startsWith('+')) {
            className = DIFF_CLASSES.added
          } else if (line.startsWith('-')) {
            className = DIFF_CLASSES.removed
          } else if (line.startsWith('~')) {
            className = DIFF_CLASSES.changed
          }
        }

        if (line.includes('"Known after apply"')) {
          className = className
            ? `${className} ${DIFF_CLASSES.afterApply}`
            : DIFF_CLASSES.afterApply
        }

        return className ? { className } : {}
      }}
      renderer={
        isDiff
          ? ({ rows, stylesheet, useInlineStyles }) => {
              const renderedRows = rows.map((row, i) =>
                renderDiffRow(row, i, lines, stylesheet, useInlineStyles)
              )

              if (!collapseUnchanged || rows.length !== lines.length) {
                return renderedRows
              }

              return collapseUnchangedRuns(renderedRows, lines, wrapLongLines)
            }
          : undefined
      }
      codeTagProps={{
        className: cn(
          'bg-code font-mono w-full',
          isDiff && 'block',
          isDiff && !wrapLongLines && 'min-w-fit',
          wrapLongLines && '!whitespace-pre-wrap break-all !pr-4'
        ),
      }}
      customStyle={{
        background: 'var(--bg-code)',
        fontFamily: 'var(--font-hack)',
      }}
    >
      {children}
    </Prism>
  )
}

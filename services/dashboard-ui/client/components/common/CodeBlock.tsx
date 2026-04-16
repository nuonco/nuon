import { Prism } from 'react-syntax-highlighter'
import {
  oneDark,
  oneLight,
} from 'react-syntax-highlighter/dist/cjs/styles/prism'
import createElement from 'react-syntax-highlighter/dist/cjs/create-element'
import { useSystemTheme } from '@/hooks/use-system-theme'
import { cn } from '@/utils/classnames'

const DIFF_CLASSES = {
  added:
    'bg-green-500/5 dark:bg-green-500/5 text-green-800 dark:text-green-500 block w-full',
  removed:
    'bg-red-500/5 dark:bg-red-500/5 text-red-800 dark:text-red-500 block w-full',
  changed:
    'bg-orange-500/5 dark:bg-orange-500/5 text-orange-800 dark:text-orange-400 block w-full',
  afterApply: '!italic opacity-70',
}

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
  showLineNumbers?: boolean
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
      {key}{' '}
      <span className="line-through opacity-70 text-red-800 dark:text-red-400">
        {oldVal}
      </span>
      <span className="opacity-50">{' -> '}</span>
      {newVal}
    </>
  )
}

export function CodeBlock({
  className,
  children,
  language,
  isDiff = false,
  showLineNumbers = false,
}: ICodeBlock) {
  const colorScheme = useSystemTheme()
  const theme = colorScheme === 'dark' ? oneDark : oneLight
  const lines = isDiff ? children.split('\n') : []

  return (
    <Prism
      className={cn(
        '!m-0 !p-2 !text-sm !rounded-md min-h-[3rem] max-h-[40rem] overflow-auto',
        className
      )}
      language={language}
      style={theme}
      wrapLines
      showLineNumbers={showLineNumbers || isDiff}
      lineProps={(lineNumber: number) => {
        if (typeof lineNumber !== 'number') return {}
        const line = isDiff ? (lines[lineNumber - 1] || '') : ''
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
              return rows.map((row, i) => {
                const line = lines[i] || ''
                const defaultEl = createElement({
                  node: row,
                  stylesheet,
                  useInlineStyles,
                  key: `line-${i}`,
                }) as any

                if (!line.startsWith('~') || !line.includes(' -> ')) {
                  return defaultEl
                }

                const children = Array.isArray(defaultEl.props.children)
                  ? defaultEl.props.children
                  : [defaultEl.props.children]

                const isLineNumber = (child: any) =>
                  child?.props?.className?.includes('linenumber')

                const lineNumberChild = children.find(isLineNumber)
                const newChildren = lineNumberChild
                  ? [lineNumberChild, renderChangedLine(line)]
                  : [renderChangedLine(line)]

                return {
                  ...defaultEl,
                  props: { ...defaultEl.props, children: newChildren },
                  key: `line-${i}`,
                }
              })
            }
          : undefined
      }
      codeTagProps={{
        className: 'bg-code font-mono w-full',
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

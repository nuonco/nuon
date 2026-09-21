import { useEffect, useId, useMemo, useRef, useState } from 'react'
import {
  CodeView,
  File,
  type CodeViewHandle,
  type CodeViewItem,
  type FileContents,
} from '@pierre/diffs/react'
import { cn } from '@/utils/classnames'
import { Button } from '@/components/common/Button'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import {
  SYNTAX_THEME,
  registerSyntax,
  resolveLanguage,
  type TSyntaxLanguage,
} from '@/lib/syntax'
import {
  MATCH_NAV_TOOLTIP,
  lineMatches,
  matchNavKeyDown,
} from './code-search'

registerSyntax()

const VIRTUALIZE_ABOVE_LINES = 300

const EXTENSIONS: Partial<Record<TSyntaxLanguage, string>> = {
  shellscript: 'sh',
  json: 'json',
  yaml: 'yaml',
  hcl: 'hcl',
  terraform: 'tf',
  toml: 'toml',
  markdown: 'md',
  docker: 'Dockerfile',
  mermaid: 'mmd',
  rego: 'rego',
  text: 'txt',
}

export interface ICodeBlock {
  value: string
  language?: string
  filename?: string
  defaultWrap?: boolean
  copy?: boolean
  lineNumbers?: boolean
  maxHeight?: number
  className?: string
}

export const CodeBlock = ({
  value,
  language,
  filename,
  defaultWrap = false,
  copy = false,
  lineNumbers,
  maxHeight = 480,
  className,
}: ICodeBlock) => {
  const generatedId = useId()
  const viewer = useRef<CodeViewHandle<undefined>>(null)
  const [query, setQuery] = useState('')
  const [matchIndex, setMatchIndex] = useState(0)
  const [wrap, setWrap] = useState(defaultWrap)
  const [scrolled, setScrolled] = useState(false)

  const lang = resolveLanguage(language)
  const lineCount = useMemo(() => value.split('\n').length, [value])
  const virtualized = lineCount > VIRTUALIZE_ABOVE_LINES
  const showLineNumbers = lineNumbers ?? lineCount > 1

  const file = useMemo(
    () => ({
      name: filename ?? `block.${EXTENSIONS[lang] ?? 'txt'}`,
      contents: value,
      lang: lang as FileContents['lang'],
    }),
    [filename, lang, value]
  )

  const matches = useMemo(() => lineMatches(value, query), [query, value])

  useEffect(() => {
    setMatchIndex(0)
    const first = matches[0]
    if (!first) return
    viewer.current?.scrollTo({
      type: 'line',
      id: generatedId,
      lineNumber: first,
      align: 'center',
      behavior: 'smooth-auto',
    })
  }, [generatedId, matches])

  const highlightCSS = useMemo(() => {
    if (!matches.length) return undefined
    const all = matches
      .map((line) => `[data-line-index="${line - 1}"]`)
      .join(',')
    const current = `[data-line-index="${(matches[matchIndex] ?? matches[0]) - 1}"]`
    return (
      `${all}{background-color:var(--code-match);}` +
      `${current}{background-color:var(--code-match-current);}`
    )
  }, [matches, matchIndex])

  const options = useMemo(
    () => ({
      theme: SYNTAX_THEME,
      disableFileHeader: !filename,
      disableLineNumbers: !showLineNumbers,
      overflow: (wrap ? 'wrap' : 'scroll') as 'wrap' | 'scroll',
      unsafeCSS: highlightCSS,
    }),
    [filename, showLineNumbers, wrap, highlightCSS]
  )

  const items = useMemo<CodeViewItem<undefined>[]>(
    () => [{ id: generatedId, type: 'file', file }],
    [generatedId, file]
  )

  const goTo = (index: number) => {
    if (!matches.length) return
    const next = (index + matches.length) % matches.length
    setMatchIndex(next)
    viewer.current?.scrollTo({
      type: 'line',
      id: generatedId,
      lineNumber: matches[next],
      align: 'center',
      behavior: 'smooth-auto',
    })
  }

  const copyButton = copy ? (
    <ClickToCopyButton textToCopy={value} title="Copy code" />
  ) : null

  const search = virtualized ? (
    <div className="flex items-center gap-2 border-b px-2 py-1.5">
      <SearchInput
        value={query}
        placeholder="Find in block"
        aria-label="Find in block"
        onChange={setQuery}
        onKeyDown={matchNavKeyDown(matchIndex, goTo)}
        labelClassName="w-full max-w-xl flex-1"
        className="!h-8 md:min-w-0 w-full"
      />
      <Text
        variant="subtext"
        theme="neutral"
        className="w-20 shrink-0 text-right tabular-nums"
      >
        {query
          ? `${matches.length ? matchIndex + 1 : 0} of ${matches.length}`
          : `${lineCount} lines`}
      </Text>
      <Button
        size="sm"
        variant="icon"
        aria-label="Previous match"
        tooltipProps={{ tipContent: MATCH_NAV_TOOLTIP.previous }}
        disabled={!matches.length}
        onClick={() => goTo(matchIndex - 1)}
      >
        <Icon variant="CaretUpIcon" size={14} />
      </Button>
      <Button
        size="sm"
        variant="icon"
        aria-label="Next match"
        tooltipProps={{ tipContent: MATCH_NAV_TOOLTIP.next }}
        disabled={!matches.length}
        onClick={() => goTo(matchIndex + 1)}
      >
        <Icon variant="CaretDownIcon" size={14} />
      </Button>

      <span aria-hidden className="mx-0.5 h-4 border-l" />

      <Button
        size="sm"
        variant="icon"
        aria-pressed={wrap}
        aria-label={wrap ? 'Stop wrapping lines' : 'Wrap lines'}
        tooltipProps={{
          tipContent: wrap ? 'Stop wrapping lines' : 'Wrap lines',
        }}
        onClick={() => setWrap((current) => !current)}
      >
        <Icon
          variant={wrap ? 'ArrowElbowDownLeftIcon' : 'ArrowsHorizontalIcon'}
          size={14}
        />
      </Button>
      <Button
        size="sm"
        variant="icon"
        aria-label="Back to top"
        tooltipProps={{ tipContent: 'Back to top' }}
        disabled={!scrolled}
        onClick={() => {
          viewer.current?.scrollTo({
            type: 'position',
            position: 0,
            behavior: 'smooth',
          })
          setScrolled(false)
        }}
      >
        <Icon variant="ArrowUpIcon" size={14} />
      </Button>
      {copyButton}
    </div>
  ) : null

  return (
    <div
      data-virtualized={virtualized || undefined}
      className={cn(
        'relative overflow-hidden rounded-lg border bg-code-bg',
        className
      )}
    >
      {search}
      {copyButton && !virtualized ? (
        <div className="absolute top-1.5 right-1.5 z-10">{copyButton}</div>
      ) : null}
      {virtualized ? (
        <CodeView
          ref={viewer}
          items={items}
          options={options}
          className="overflow-auto"
          style={{ maxHeight: `${maxHeight}px` }}
          onScroll={(scrollTop) => setScrolled(scrollTop > 200)}
        />
      ) : (
        <File file={file} options={options} />
      )}
    </div>
  )
}

import { useId, useMemo, useRef, useState } from 'react'
import {
  CodeView,
  File,
  type CodeViewHandle,
  type CodeViewItem,
  type FileContents,
} from '@pierre/diffs/react'
import { cn } from '@/utils/classnames'
import {
  LITE_SYNTAX_THEME,
  registerSyntax,
  resolveLanguage,
  type TSyntaxLanguage,
} from '../../lib/syntax'
import { Button } from '../atoms/Button'
import { CopyButton } from '../atoms/CopyButton'
import { Icon } from '../atoms/Icon'
import { Input } from '../atoms/Input'
import { Text } from '../atoms/Text'

registerSyntax()

// Below this, a plain File is cheaper than a scroll container plus a
// virtualizer. Above it, the DOM cost of every line is what kills the page.
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
  /** Starting state of the wrap toggle. The block owns it after that. */
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
      // rego is registered by us, so it is outside the library's bundled union.
      lang: lang as FileContents['lang'],
    }),
    [filename, lang, value]
  )

  const matches = useMemo(() => {
    if (!query) return []
    const needle = query.toLowerCase()
    return value
      .split('\n')
      .map((line, index) => (line.toLowerCase().includes(needle) ? index + 1 : 0))
      .filter(Boolean)
  }, [query, value])

  // Lines carry data-line-index in the renderer's shadow root, and unsafeCSS is
  // injected into it, so matches are painted rather than annotated. That index
  // is zero-based, while scrollTo takes a one-based line number.
  const highlightCSS = useMemo(() => {
    if (!matches.length) return undefined
    const all = matches.map((line) => `[data-line-index="${line - 1}"]`).join(',')
    const current = `[data-line-index="${(matches[matchIndex] ?? matches[0]) - 1}"]`
    return (
      `${all}{background-color:var(--code-match);}` +
      `${current}{background-color:var(--code-match-current);}`
    )
  }, [matches, matchIndex])

  const options = useMemo(
    () => ({
      theme: LITE_SYNTAX_THEME,
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

  const copyButton = copy ? <CopyButton value={value} label="Copy code" /> : null

  const search = virtualized ? (
    <div className="flex items-center gap-2 border-b border-divider px-2 py-1.5">
      <label className="relative min-w-0 flex-1">
        <span className="sr-only">Find in block</span>
        <Icon
          variant="MagnifyingGlassIcon"
          size={14}
          aria-hidden
          className="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-tertiary"
        />
        <Input
          type="search"
          size="sm"
          value={query}
          placeholder="Find in block"
          onChange={(event) => {
            setQuery(event.target.value)
            setMatchIndex(0)
          }}
          onKeyDown={(event) => {
            if (event.key !== 'Enter') return
            event.preventDefault()
            goTo(event.shiftKey ? matchIndex - 1 : matchIndex + 1)
          }}
          className="pl-8"
        />
      </label>
      <Text variant="caption" color="tertiary">
        {query ? `${matches.length ? matchIndex + 1 : 0} of ${matches.length}` : `${lineCount} lines`}
      </Text>
      <Button
        size="sm"
        variant="ghost"
        iconOnly
        aria-label="Previous match"
        disabled={!matches.length}
        onClick={() => goTo(matchIndex - 1)}
      >
        <Icon variant="CaretUpIcon" size={14} />
      </Button>
      <Button
        size="sm"
        variant="ghost"
        iconOnly
        aria-label="Next match"
        disabled={!matches.length}
        onClick={() => goTo(matchIndex + 1)}
      >
        <Icon variant="CaretDownIcon" size={14} />
      </Button>

      <span aria-hidden className="mx-0.5 h-4 w-px bg-divider" />

      <Button
        size="sm"
        variant="ghost"
        iconOnly
        aria-pressed={wrap}
        aria-label={wrap ? 'Stop wrapping lines' : 'Wrap lines'}
        tooltip={wrap ? 'Stop wrapping lines' : 'Wrap lines'}
        className={wrap ? 'text-primary' : undefined}
        onClick={() => setWrap((current) => !current)}
      >
        <Icon
          variant={wrap ? 'ArrowElbowDownLeftIcon' : 'ArrowsHorizontalIcon'}
          size={14}
        />
      </Button>
      <Button
        size="sm"
        variant="ghost"
        iconOnly
        aria-label="Back to top"
        tooltip="Back to top"
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
        'relative overflow-hidden rounded-lg border border-divider bg-code-bg',
        className
      )}
    >
      {search}
      {/* With a search row the copy button lives in it; floating it would sit
          on top of the next/previous controls. */}
      {copyButton && !virtualized ? (
        <div className="absolute top-1.5 right-1.5 z-10">{copyButton}</div>
      ) : null}
      {virtualized ? (
        // CodeView owns its own scroll root and per-line virtualization, so it
        // must not be wrapped in a Virtualizer — that is the mixed-content path.
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

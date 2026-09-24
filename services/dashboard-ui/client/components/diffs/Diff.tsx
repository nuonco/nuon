import { parseDiffFromFile } from '@pierre/diffs'
import {
  CodeView,
  type CodeViewHandle,
  type CodeViewItem,
  type FileContents,
} from '@pierre/diffs/react'
import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { cn } from '@/utils/classnames'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { useDashboardPreferences } from '@/hooks/use-dashboard-preferences'
import { SYNTAX_THEME, registerSyntax, resolveLanguage } from '@/lib/syntax'
import { endWithNewline } from '@/lib/diffs'
import { MATCH_NAV_TOOLTIP, diffMatches, matchNavKeyDown } from './code-search'

registerSyntax()

const DIFF_CSS = `
:host {
  --diffs-light-bg: var(--code-bg);
  --diffs-dark-bg: var(--code-bg);
  --diffs-light-addition-color: var(--diff-add-gutter);
  --diffs-dark-addition-color: var(--diff-add-gutter);
  --diffs-light-deletion-color: var(--diff-remove-gutter);
  --diffs-dark-deletion-color: var(--diff-remove-gutter);
  --diffs-bg-addition-emphasis-override: var(--diff-add-emphasis);
  --diffs-bg-deletion-emphasis-override: var(--diff-remove-emphasis);
}
[data-line-type="change-addition"],
[data-line-type="change-addition"][data-selected-line] {
  --diffs-computed-diff-line-bg: var(--diff-add-row);
}
[data-line-type="change-deletion"],
[data-line-type="change-deletion"][data-selected-line] {
  --diffs-computed-diff-line-bg: var(--diff-remove-row);
}
[data-line][data-selected-line],
[data-column-number][data-selected-line] {
  --diffs-computed-selected-line-bg: var(--code-match-current);
}
`

export type TDiffView = 'unified' | 'split'

export interface IDiff {
  before: string
  after: string
  language?: string
  filename?: string
  view?: TDiffView
  defaultWrap?: boolean
  lineNumbers?: boolean
  search?: boolean
  maxHeight?: number
  className?: string
}

export const Diff = ({
  before,
  after,
  language,
  filename,
  view = 'unified',
  defaultWrap,
  lineNumbers = true,
  search = true,
  maxHeight = 640,
  className,
}: IDiff) => {
  const { diffWrap } = useDashboardPreferences()
  const id = useId()
  const viewer = useRef<CodeViewHandle<undefined>>(null)
  const [query, setQuery] = useState('')
  const [matchIndex, setMatchIndex] = useState(0)
  const [wrap, setWrap] = useState(defaultWrap ?? diffWrap === 'wrap')
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    if (defaultWrap === undefined) setWrap(diffWrap === 'wrap')
  }, [defaultWrap, diffWrap])

  const lang = resolveLanguage(language)
  const name = filename ?? `change.${lang === 'terraform' ? 'tf' : 'txt'}`

  const beforeText = endWithNewline(before)
  const afterText = endWithNewline(after)

  const fileDiff = useMemo(() => {
    const file = (contents: string): FileContents => ({
      name,
      contents,
      lang: lang as FileContents['lang'],
    })
    return parseDiffFromFile(file(beforeText), file(afterText))
  }, [afterText, beforeText, lang, name])

  const lineCount = useMemo(
    () =>
      Math.max(
        before ? before.split('\n').length : 0,
        after ? after.split('\n').length : 0
      ),
    [after, before]
  )

  const matches = useMemo(() => diffMatches(fileDiff, query), [fileDiff, query])

  useEffect(() => {
    setMatchIndex(0)
    const first = matches[0]
    if (!first) return
    viewer.current?.scrollTo({
      type: 'line',
      id,
      lineNumber: first.lineNumber,
      side: first.side,
      align: 'center',
      behavior: 'smooth-auto',
    })
  }, [id, matches])

  const currentMatch = matches[matchIndex]
  const selectedLines = currentMatch
    ? {
        id,
        range: {
          start: currentMatch.lineNumber,
          end: currentMatch.lineNumber,
          side: currentMatch.side,
          endSide: currentMatch.side,
        },
      }
    : null

  const options = useMemo(
    () => ({
      theme: SYNTAX_THEME,
      disableFileHeader: true,
      disableLineNumbers: !lineNumbers,
      overflow: (wrap ? 'wrap' : 'scroll') as 'wrap' | 'scroll',
      diffStyle: view,
      diffIndicators: 'classic' as const,
      hunkSeparators: 'line-info' as const,
      collapsedContextThreshold: 15,
      expansionLineCount: 3,
      lineDiffType: 'word' as const,
      unsafeCSS: DIFF_CSS,
    }),
    [lineNumbers, view, wrap]
  )

  const items = useMemo<CodeViewItem<undefined>[]>(
    () => [{ id, type: 'diff', fileDiff }],
    [fileDiff, id]
  )

  const goTo = (index: number) => {
    if (!matches.length) return
    const next = (index + matches.length) % matches.length
    const match = matches[next]
    setMatchIndex(next)
    viewer.current?.scrollTo({
      type: 'line',
      id,
      lineNumber: match.lineNumber,
      side: match.side,
      align: 'center',
      behavior: 'smooth-auto',
    })
  }

  return (
    <div
      data-diff-view={view}
      className={cn('overflow-hidden rounded-lg border bg-code-bg', className)}
    >
      {search || filename ? (
        <div className="flex items-center gap-2 border-b px-2 py-1.5">
          {filename ? (
            <span className="flex min-w-0 items-center gap-1.5">
              <Icon
                variant="FileTextIcon"
                size={14}
                aria-hidden
                theme="neutral"
              />
              <Text
                variant="subtext"
                family="mono"
                theme="neutral"
                className="truncate"
              >
                {filename}
              </Text>
            </span>
          ) : null}
          {filename && search ? (
            <span aria-hidden className="mx-0.5 h-4 border-l" />
          ) : null}
          {search ? (
            <>
              <SearchInput
                value={query}
                placeholder="Find in diff"
                aria-label="Find in diff"
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
            </>
          ) : (
            <span className="flex-1" />
          )}

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
        </div>
      ) : null}
      <CodeView
        ref={viewer}
        items={items}
        options={options}
        selectedLines={selectedLines}
        className="overflow-auto"
        style={{ maxHeight: `${maxHeight}px` }}
        onScroll={(scrollTop) => setScrolled(scrollTop > 200)}
      />
    </div>
  )
}

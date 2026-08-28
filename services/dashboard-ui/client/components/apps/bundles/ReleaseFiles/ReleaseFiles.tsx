import { useEffect, useMemo, useRef, useState } from 'react'
import { File as DiffFile, MultiFileDiff } from '@pierre/diffs/react'
import { FileTree, useFileTree } from '@pierre/trees/react'
import type { GitStatusEntry } from '@pierre/trees'
import { Badge } from '@/components/common/Badge'
import { Hash } from '@/components/common/Hash'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Select } from '@/components/common/form/Select'
import type { TReleaseFileContent } from '@/types'
import { formatBytes } from '@/utils/string-utils'

export type TReleaseFileEntry = {
  category: 'source' | 'artifact' | 'runtime' | 'document'
  change: 'added' | 'modified' | 'removed' | 'unchanged'
  current?: TReleaseFileVersion
  path: string
  previous?: TReleaseFileVersion
}

export type TReleaseFileVersion = {
  digest?: string
  mediaType?: string
  metadata?: Record<string, unknown>
  size?: number
}

const changeTheme = {
  added: 'success',
  modified: 'warn',
  removed: 'error',
  unchanged: 'neutral',
} as const

const diffOptions = {
  diffStyle: 'unified',
  expandUnchanged: false,
  lineDiffType: 'word',
  overflow: 'scroll',
  themeType: 'system',
} as const

const fileOptions = {
  overflow: 'scroll',
  themeType: 'system',
} as const

const previewFile = (
  entry: TReleaseFileEntry,
  side: 'current' | 'previous',
  content?: TReleaseFileContent
) => {
  const version = entry[side]
  const isSource = entry.category === 'source'
  if (!version) return null
  if (
    isSource &&
    content?.path === entry.path &&
    (!version.digest || content.digest === version.digest)
  ) {
    return {
      cacheKey: version.digest,
      contents: content.content,
      name: entry.path,
    }
  }
  if (version.metadata) {
    return {
      cacheKey: version.digest,
      contents: JSON.stringify(version.metadata, null, 2),
      lang: 'json' as const,
      name: entry.path,
    }
  }
  return null
}

const FilePreview = ({
  currentContent,
  entry,
  isContentLoading,
  previousContent,
}: {
  currentContent?: TReleaseFileContent
  entry: TReleaseFileEntry
  isContentLoading?: boolean
  previousContent?: TReleaseFileContent
}) => {
  const version = entry.current ?? entry.previous
  const currentFile = useMemo(
    () => previewFile(entry, 'current', currentContent),
    [currentContent, entry]
  )
  const previousFile = useMemo(
    () => previewFile(entry, 'previous', previousContent),
    [entry, previousContent]
  )
  const preview =
    entry.change === 'unchanged' ? (currentFile ?? previousFile) : null

  return (
    <div className="flex flex-col gap-4 min-w-0">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex flex-col gap-1 min-w-0">
          <Text family="mono" weight="strong" className="break-all">
            {entry.path}
          </Text>
          <div className="flex items-center gap-2">
            <Badge theme={changeTheme[entry.change]}>{entry.change}</Badge>
            <Badge variant="code">{entry.category}</Badge>
            {version?.size ? (
              <Text variant="subtext" theme="neutral">
                {formatBytes(version.size)}
              </Text>
            ) : null}
          </div>
        </div>
        {version?.digest ? <Hash hash={version.digest} /> : null}
      </div>

      {isContentLoading ? (
        <div className="flex justify-center py-12">
          <Loading />
        </div>
      ) : preview ? (
        <DiffFile
          key={`${entry.path}:${version?.digest ?? ''}`}
          file={preview}
          options={fileOptions}
        />
      ) : currentFile || previousFile ? (
        <MultiFileDiff
          key={`${entry.path}:${entry.previous?.digest ?? ''}:${entry.current?.digest ?? ''}`}
          newFile={currentFile}
          oldFile={previousFile}
          options={diffOptions}
        />
      ) : (
        <Text theme="neutral">
          This package entry has no text preview. Its digest and metadata are
          still available for verification.
        </Text>
      )}
    </div>
  )
}

const ReleaseFileTree = ({
  entries,
  onSelect,
  selectedPath,
}: {
  entries: TReleaseFileEntry[]
  onSelect: (path: string) => void
  selectedPath?: string
}) => {
  const entriesRef = useRef(entries)
  entriesRef.current = entries
  const paths = useMemo(() => entries.map(({ path }) => path), [entries])
  const gitStatus = useMemo<GitStatusEntry[]>(
    () =>
      entries.flatMap(({ change, path }) =>
        change === 'unchanged'
          ? []
          : [
              {
                path,
                status:
                  change === 'removed'
                    ? 'deleted'
                    : change === 'added'
                      ? 'added'
                      : 'modified',
              },
            ]
      ),
    [entries]
  )
  const { model } = useFileTree({
    fileTreeSearchMode: 'expand-matches',
    flattenEmptyDirectories: true,
    gitStatus,
    icons: 'standard',
    initialExpansion: 'open',
    initialSelectedPaths: selectedPath ? [selectedPath] : [],
    onSelectionChange: (selectedPaths) => {
      const path = selectedPaths.at(-1)
      if (path && entriesRef.current.some((entry) => entry.path === path)) {
        onSelect(path)
      }
    },
    paths,
    search: true,
  })

  useEffect(() => {
    model.resetPaths(paths)
    model.setGitStatus(gitStatus)
  }, [gitStatus, model, paths])

  useEffect(() => {
    for (const path of model.getSelectedPaths()) {
      if (path !== selectedPath) model.getItem(path)?.deselect()
    }
    if (selectedPath) {
      model.getItem(selectedPath)?.select()
      model.scrollToPath(selectedPath, { offset: 'nearest' })
    }
  }, [model, selectedPath])

  return (
    <FileTree
      className="h-144 w-full"
      model={model}
      style={
        {
          '--trees-bg-override': 'var(--background)',
          '--trees-border-color-override': 'var(--border-color)',
          '--trees-fg-override': 'var(--foreground)',
          '--trees-selected-bg-override': 'var(--background-neutral)',
        } as React.CSSProperties
      }
    />
  )
}

export const ReleaseFiles = ({
  currentContent,
  entries,
  isContentLoading,
  onPackageChange,
  onSelect,
  packageOptions,
  packageStatus,
  previousContent,
  selectedPackageId,
  selectedPath,
}: {
  currentContent?: TReleaseFileContent
  entries: TReleaseFileEntry[]
  isContentLoading?: boolean
  onPackageChange: (packageId: string) => void
  onSelect: (path: string) => void
  packageOptions: { id: string; platform: string; status?: string }[]
  packageStatus?: string
  previousContent?: TReleaseFileContent
  selectedPackageId?: string
  selectedPath?: string
}) => {
  const [changedOnly, setChangedOnly] = useState(false)
  const [category, setCategory] = useState('all')
  const selected = entries.find(({ path }) => path === selectedPath)
  const filtered = entries.filter((entry) => {
    if (changedOnly && entry.change === 'unchanged') return false
    if (category !== 'all' && entry.category !== category) return false
    return true
  })
  const changed = entries.filter(({ change }) => change !== 'unchanged')

  useEffect(() => {
    if (
      filtered.length > 0 &&
      !filtered.some(({ path }) => path === selectedPath)
    ) {
      onSelect(filtered[0].path)
    }
  }, [filtered, onSelect, selectedPath])

  return (
    <div className="flex flex-col gap-6 mt-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-2">
          <Text variant="h3" weight="strong">
            Release files
          </Text>
          <div className="flex flex-wrap gap-2">
            <Badge>{entries.length} files</Badge>
            <Badge theme="success">
              {changed.filter(({ change }) => change === 'added').length} added
            </Badge>
            <Badge theme="warn">
              {changed.filter(({ change }) => change === 'modified').length}{' '}
              modified
            </Badge>
            <Badge theme="error">
              {changed.filter(({ change }) => change === 'removed').length}{' '}
              removed
            </Badge>
          </div>
        </div>
        {packageOptions.length > 1 ? (
          <Select
            className="w-64"
            labelProps={{ labelText: 'Platform' }}
            options={packageOptions.map((option) => ({
              label: option.platform,
              value: option.id,
              badge: { label: option.status ?? 'unknown' },
            }))}
            onChange={onPackageChange}
            value={selectedPackageId}
          />
        ) : packageOptions[0] ? (
          <div className="flex flex-col gap-1">
            <Text variant="subtext" theme="neutral">
              Platform
            </Text>
            <Badge variant="code">{packageOptions[0].platform}</Badge>
          </div>
        ) : null}
      </div>

      {packageStatus && packageStatus !== 'active' ? (
        <Text theme="neutral">
          Portable package is {packageStatus}. Package contents will appear when
          publishing completes. Authored files are available now.
        </Text>
      ) : null}

      <div className="flex flex-wrap items-center justify-end gap-3">
        <Select
          className="w-48"
          options={[
            { label: 'All files', value: 'all' },
            { label: 'Source', value: 'source' },
            { label: 'Artifacts', value: 'artifact' },
            { label: 'Runtime', value: 'runtime' },
            { label: 'Documents', value: 'document' },
          ]}
          onChange={setCategory}
          value={category}
        />
        <CheckboxInput
          checked={changedOnly}
          labelProps={{ labelText: 'Changed only' }}
          onChange={(event) => setChangedOnly(event.currentTarget.checked)}
        />
      </div>

      <div className="flex min-h-144 flex-col overflow-hidden border rounded-md lg:flex-row">
        <div className="border-b lg:w-2/5 lg:border-b-0 lg:border-r">
          {filtered.length ? (
            <ReleaseFileTree
              entries={filtered}
              onSelect={onSelect}
              selectedPath={selectedPath}
            />
          ) : (
            <Text theme="neutral" className="p-4">
              No files match these filters. Clear the search or filters to view
              release files.
            </Text>
          )}
        </div>
        <div className="min-w-0 p-4 overflow-auto max-h-192 lg:w-3/5">
          {selected ? (
            <FilePreview
              currentContent={currentContent}
              entry={selected}
              isContentLoading={isContentLoading}
              previousContent={previousContent}
            />
          ) : (
            <Text theme="neutral">Select a file to inspect it.</Text>
          )}
        </div>
      </div>
    </div>
  )
}

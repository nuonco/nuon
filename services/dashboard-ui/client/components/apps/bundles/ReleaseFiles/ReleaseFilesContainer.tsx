import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAppReleaseFileContent } from '@/lib'
import type { TAppReleaseWithFiles, TReleaseFile } from '@/types'
import {
  ReleaseFiles,
  releaseFileEntryCanPreview,
  type TReleaseFileEntry,
  type TReleaseFileVersion,
} from './ReleaseFiles'

const sourceVersions = (files?: TReleaseFile[]) =>
  new Map<string, TReleaseFileVersion>(
    (files ?? []).map((file) => [
      file.path,
      {
        digest: file.digest,
        mediaType: file.media_type,
        size: file.size,
      },
    ])
  )

const changedEntry = (
  path: string,
  category: TReleaseFileEntry['category'],
  current?: TReleaseFileVersion,
  previous?: TReleaseFileVersion
): TReleaseFileEntry => ({
  path,
  category,
  current,
  previous,
  change: !previous
    ? 'added'
    : !current
      ? 'removed'
      : current.digest
        ? current.digest === previous.digest
          ? 'unchanged'
          : 'modified'
        : JSON.stringify(current.metadata) === JSON.stringify(previous.metadata)
          ? 'unchanged'
          : 'modified',
})

export const releaseFileEntries = (
  release: TAppReleaseWithFiles,
  previousRelease: TAppReleaseWithFiles | undefined
) => {
  const currentSource = sourceVersions(release.source_files)
  const previousSource = sourceVersions(previousRelease?.source_files)
  const entries: TReleaseFileEntry[] = []
  for (const path of new Set([
    ...currentSource.keys(),
    ...previousSource.keys(),
  ])) {
    entries.push(
      changedEntry(
        path,
        'source',
        currentSource.get(path),
        previousSource.get(path)
      )
    )
  }
  return entries.sort((a, b) => a.path.localeCompare(b.path))
}

export const ReleaseFilesContainer = ({
  appId,
  orgId,
  previousRelease,
  release,
}: {
  appId: string
  orgId: string
  previousRelease?: TAppReleaseWithFiles
  release: TAppReleaseWithFiles
}) => {
  const [selectedPath, setSelectedPath] = useState<string>()
  const entries = useMemo(
    () => releaseFileEntries(release, previousRelease),
    [previousRelease, release]
  )

  useEffect(() => {
    if (!selectedPath || !entries.some(({ path }) => path === selectedPath)) {
      setSelectedPath(
        entries.find(({ change }) => change !== 'unchanged')?.path ??
          entries[0]?.path
      )
    }
  }, [entries, selectedPath])

  const selected = entries.find(({ path }) => path === selectedPath)
  const currentSourcePath =
    selected?.category === 'source' &&
    selected.current &&
    releaseFileEntryCanPreview(selected)
      ? selected.path
      : undefined
  const previousSourcePath =
    selected?.category === 'source' &&
    selected.previous &&
    releaseFileEntryCanPreview(selected)
      ? selected.path
      : undefined
  const currentContentQuery = useQuery({
    queryKey: [
      'release-file-content',
      orgId,
      appId,
      release.id,
      currentSourcePath,
    ],
    queryFn: () =>
      getAppReleaseFileContent({
        appId,
        orgId,
        path: currentSourcePath!,
        releaseId: release.id!,
      }),
    enabled: !!currentSourcePath,
  })
  const previousContentQuery = useQuery({
    queryKey: [
      'release-file-content',
      orgId,
      appId,
      previousRelease?.id,
      previousSourcePath,
    ],
    queryFn: () =>
      getAppReleaseFileContent({
        appId,
        orgId,
        path: previousSourcePath!,
        releaseId: previousRelease!.id!,
      }),
    enabled: !!previousSourcePath && !!previousRelease?.id,
  })

  return (
    <ReleaseFiles
      currentContent={currentContentQuery.data}
      entries={entries}
      isContentLoading={
        currentContentQuery.isFetching || previousContentQuery.isFetching
      }
      onSelect={setSelectedPath}
      previousContent={previousContentQuery.data}
      selectedPath={selectedPath}
    />
  )
}

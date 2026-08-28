import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAppReleaseFileContent, getReleasePackages } from '@/lib'
import type {
  TAppReleaseWithFiles,
  TReleaseFile,
  TReleasePackage,
  TReleasePackageMember,
} from '@/types'
import {
  ReleaseFiles,
  releaseFileEntryCanPreview,
  type TReleaseFileEntry,
  type TReleaseFileVersion,
} from './ReleaseFiles'

const packageMemberPath = (member: TReleasePackageMember) => {
  const name = member.logical_name ?? 'unnamed'
  switch (member.kind) {
    case 'component':
    case 'image':
      return `package/components/${name}`
    case 'sandbox':
      return `package/sandbox/${name}`
    case 'action_step':
      return `package/actions/${name}`
    case 'stack_asset':
      return `package/stack/${name}`
    case 'portal_binary':
    case 'runner_binary':
    case 'runner_image':
      return `package/runtime/${member.kind}/${name}`
    case 'source_archive':
      return 'package/documents/release-source.json'
    default:
      return `package/artifacts/${member.kind ?? 'unknown'}/${name}`
  }
}

const packageVersions = (pkg?: TReleasePackage) => {
  const versions = new Map<string, TReleaseFileVersion>()
  if (pkg) {
    versions.set('package/documents/logical-manifest.json', {
      digest: pkg.manifest_digest,
      metadata: { digest: pkg.manifest_digest },
    })
    versions.set('package/documents/plan-envelope.json', {
      digest: pkg.plan_digest,
      metadata: { digest: pkg.plan_digest },
    })
    versions.set('package/documents/oci-index.json', {
      digest: pkg.oci_index_digest,
      metadata: { digest: pkg.oci_index_digest },
    })
    versions.set('package/archive.tar.zst', {
      digest: pkg.archive_checksum,
      size: pkg.archive_size,
      metadata: {
        checksum: pkg.archive_checksum,
        size: pkg.archive_size,
        format: pkg.format,
        platform: pkg.target_platform,
      },
    })
  }
  for (const member of pkg?.members ?? []) {
    versions.set(packageMemberPath(member), {
      digest: member.digest,
      mediaType: member.media_type,
      size: member.size,
      metadata: {
        kind: member.kind,
        logical_name: member.logical_name,
        digest: member.digest,
        media_type: member.media_type,
        size: member.size,
        repository: member.repository,
        platform_os: member.platform_os,
        platform_architecture: member.platform_architecture,
      },
    })
  }
  return versions
}

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
  previousRelease: TAppReleaseWithFiles | undefined,
  pkg?: TReleasePackage,
  previousPackage?: TReleasePackage
) => {
  const currentSource = sourceVersions(release.source_files)
  const previousSource = sourceVersions(previousRelease?.source_files)
  const currentPackage = packageVersions(pkg)
  const previousPackageVersions = packageVersions(previousPackage)
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
  for (const path of new Set([
    ...currentPackage.keys(),
    ...previousPackageVersions.keys(),
  ])) {
    const category = path.startsWith('package/runtime/')
      ? 'runtime'
      : path.startsWith('package/documents/') || path.endsWith('.tar.zst')
        ? 'document'
        : 'artifact'
    entries.push(
      changedEntry(
        path,
        category,
        currentPackage.get(path),
        previousPackageVersions.get(path)
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
  const [selectedPackageId, setSelectedPackageId] = useState<string>()
  const packagesQuery = useQuery({
    queryKey: ['release-packages', orgId, appId, release.id],
    queryFn: () =>
      getReleasePackages({ appId, orgId, releaseId: release.id! }),
    enabled: !!release.id,
  })
  const previousPackagesQuery = useQuery({
    queryKey: ['release-packages', orgId, appId, previousRelease?.id],
    queryFn: () =>
      getReleasePackages({
        appId,
        orgId,
        releaseId: previousRelease!.id!,
      }),
    enabled: !!previousRelease?.id,
  })
  const packages = packagesQuery.data ?? []
  const selectedPackage =
    packages.find(({ id }) => id === selectedPackageId) ?? packages[0]
  const previousPackage = previousPackagesQuery.data?.find(
    ({ target_platform }) =>
      target_platform === selectedPackage?.target_platform
  )
  const entries = useMemo(
    () =>
      releaseFileEntries(
        release,
        previousRelease,
        selectedPackage,
        previousPackage
      ),
    [previousPackage, previousRelease, release, selectedPackage]
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
      onPackageChange={setSelectedPackageId}
      onSelect={setSelectedPath}
      packageOptions={packages.map((pkg) => ({
        id: pkg.id!,
        platform: pkg.target_platform ?? 'unknown',
        status: pkg.status,
      }))}
      packageStatus={selectedPackage?.status}
      previousContent={previousContentQuery.data}
      selectedPackageId={selectedPackage?.id}
      selectedPath={selectedPath}
    />
  )
}

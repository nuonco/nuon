import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAppReleaseFileContent, getReleasePackage } from '@/lib'
import type {
  TAppReleaseWithFiles,
  TReleaseFile,
  TReleasePackage,
  TReleasePackageMember,
} from '@/types'
import {
  ReleaseFiles,
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

const packageMemberVersion = (
  member: TReleasePackageMember
): TReleaseFileVersion => ({
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

const packageDocumentVersions = (pkg?: TReleasePackage) => {
  if (!pkg) return new Map<string, TReleaseFileVersion>()
  return new Map<string, TReleaseFileVersion>([
    [
      'package/documents/logical-manifest.json',
      {
        digest: pkg.manifest_digest,
        metadata: { digest: pkg.manifest_digest },
      },
    ],
    [
      'package/documents/plan-envelope.json',
      { digest: pkg.plan_digest, metadata: { digest: pkg.plan_digest } },
    ],
    [
      'package/documents/qualification-report.json',
      { metadata: { included: true } },
    ],
    [
      'package/documents/bundle-provenance.json',
      { metadata: { included: true } },
    ],
    [
      'package/documents/oci-index.json',
      {
        digest: pkg.oci_index_digest,
        metadata: { digest: pkg.oci_index_digest },
      },
    ],
    [
      'package/archive.tar.zst',
      {
        digest: pkg.archive_checksum,
        size: pkg.archive_size,
        metadata: {
          checksum: pkg.archive_checksum,
          size: pkg.archive_size,
          format: pkg.format,
          platform: pkg.target_platform,
        },
      },
    ],
  ])
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

const packageVersions = (pkg?: TReleasePackage) => {
  const versions = packageDocumentVersions(pkg)
  for (const member of pkg?.members ?? []) {
    versions.set(packageMemberPath(member), packageMemberVersion(member))
  }
  return versions
}

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
  pkg: TReleasePackage | undefined,
  previousPackage: TReleasePackage | undefined
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
  const packages = release.packages ?? []
  const defaultPackageId =
    packages.find(({ status }) => status === 'active')?.id ?? packages[0]?.id
  const [selectedPackageId, setSelectedPackageId] = useState(defaultPackageId)
  const [selectedPath, setSelectedPath] = useState<string>()

  useEffect(() => {
    if (
      !selectedPackageId ||
      !packages.some(({ id }) => id === selectedPackageId)
    ) {
      setSelectedPackageId(defaultPackageId)
    }
  }, [defaultPackageId, packages, selectedPackageId])

  const { data: packageDetails } = useQuery({
    queryKey: ['release-package', orgId, selectedPackageId],
    queryFn: () => getReleasePackage({ orgId, packageId: selectedPackageId! }),
    enabled: !!selectedPackageId,
  })
  const platform = packages.find(
    ({ id }) => id === selectedPackageId
  )?.target_platform
  const previousPackageId = previousRelease?.packages?.find(
    ({ target_platform }) => target_platform === platform
  )?.id
  const { data: previousPackageDetails } = useQuery({
    queryKey: ['release-package', orgId, previousPackageId],
    queryFn: () => getReleasePackage({ orgId, packageId: previousPackageId! }),
    enabled: !!previousPackageId,
  })
  const entries = useMemo(
    () =>
      releaseFileEntries(
        release,
        previousRelease,
        packageDetails?.status === 'active' ? packageDetails : undefined,
        packageDetails?.status === 'active' &&
          previousPackageDetails?.status === 'active'
          ? previousPackageDetails
          : undefined
      ),
    [packageDetails, previousPackageDetails, previousRelease, release]
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
    selected?.category === 'source' && selected.current
      ? selected.path
      : undefined
  const previousSourcePath =
    selected?.category === 'source' && selected.previous
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
        platform: pkg.target_platform ?? 'Unknown platform',
        status: pkg.status,
      }))}
      packageStatus={packageDetails?.status}
      previousContent={previousContentQuery.data}
      selectedPackageId={selectedPackageId}
      selectedPath={selectedPath}
    />
  )
}

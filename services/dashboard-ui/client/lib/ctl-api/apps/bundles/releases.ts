import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'
import type {
  TAppRelease,
  TAppReleaseWithFiles,
  TCreateReleasePackageRequest,
  TCreateReleaseRequest,
  TPaginationParams,
  TReleasePackage,
  TReleasePackageDownloadGrant,
  TReleaseFileContent,
  TInstallAppConfigVersion,
} from '@/types'

type AppRequest = {
  appId: string
  orgId: string
}

export const createAppRelease = ({
  appId,
  body,
  orgId,
}: AppRequest & { body: TCreateReleaseRequest }) =>
  api<TAppReleaseWithFiles>({
    body,
    method: 'POST',
    orgId,
    path: `apps/${appId}/releases`,
  })

export const getAppReleaseFileContent = ({
  appId,
  orgId,
  path,
  releaseId,
}: AppRequest & { path: string; releaseId: string }) =>
  api<TReleaseFileContent>({
    method: 'GET',
    orgId,
    path: `apps/${appId}/releases/${releaseId}/files/content?path=${encodeURIComponent(path)}`,
  })

export const getAppReleases = ({
  appId,
  orgId,
  limit,
  offset,
}: AppRequest & TPaginationParams) =>
  api<TAppRelease[]>({
    method: 'GET',
    orgId,
    paginated: true,
    path: `apps/${appId}/releases${buildQueryParams({ limit, offset })}`,
  })

export const getAppRelease = ({
  appId,
  orgId,
  releaseId,
}: AppRequest & { releaseId: string }) =>
  api<TAppReleaseWithFiles>({
    method: 'GET',
    orgId,
    path: `apps/${appId}/releases/${releaseId}`,
  })

export const proposeAppRelease = ({
  installId,
  orgId,
  releaseId,
}: {
  installId: string
  orgId: string
  releaseId: string
}) =>
  api<TInstallAppConfigVersion>({
    body: { release_id: releaseId },
    method: 'POST',
    orgId,
    path: `installs/${installId}/release-updates`,
  })

export const createReleasePackage = ({
  appId,
  body,
  orgId,
  releaseId,
}: AppRequest & {
  body: TCreateReleasePackageRequest
  releaseId: string
}) =>
  api<TReleasePackage>({
    body,
    method: 'POST',
    orgId,
    path: `apps/${appId}/releases/${releaseId}/packages`,
  })

export const getReleasePackage = ({
  orgId,
  packageId,
}: {
  orgId: string
  packageId: string
}) =>
  api<TReleasePackage>({
    method: 'GET',
    orgId,
    path: `release-packages/${packageId}`,
  })

export const createReleasePackageDownloadGrant = ({
  orgId,
  packageId,
}: {
  orgId: string
  packageId: string
}) =>
  api<TReleasePackageDownloadGrant>({
    method: 'POST',
    orgId,
    path: `release-packages/${packageId}/download-grants`,
  })

import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'
import type {
  TAppRelease,
  TAppReleaseWithFiles,
  TCreateReleaseRequest,
  TPaginationParams,
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

export const getPreviousAppRelease = async ({
  appId,
  orgId,
  releaseId,
}: AppRequest & { releaseId: string }) => {
  const limit = 100
  let offset = 0
  let useFirstRelease = false

  while (true) {
    const result = await getAppReleases({ appId, orgId, limit, offset })
    if (useFirstRelease) {
      const previousRelease = result.data[0]
      return previousRelease?.id
        ? getAppRelease({ appId, orgId, releaseId: previousRelease.id })
        : undefined
    }

    const releaseIndex = result.data.findIndex(({ id }) => id === releaseId)
    const previousRelease = result.data[releaseIndex + 1]

    if (releaseIndex >= 0 && previousRelease?.id) {
      return getAppRelease({ appId, orgId, releaseId: previousRelease.id })
    }
    if (!result.pagination.hasNext) return undefined

    useFirstRelease = releaseIndex >= 0
    offset += limit
  }
}

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

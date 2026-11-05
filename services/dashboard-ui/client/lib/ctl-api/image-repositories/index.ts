import { api } from '@/lib/api'
import type {
  TCreateImageRepositoryRequest,
  TImageRepository,
  TPaginationParams,
} from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getImageRepositories = ({
  orgId,
  ...params
}: { orgId: string; q?: string; platform?: string } & TPaginationParams) =>
  api<TImageRepository[]>({
    orgId,
    path: `image-repositories${buildQueryParams(params)}`,
    paginated: true,
  })

export const getImageRepository = ({
  orgId,
  repositoryId,
}: {
  orgId: string
  repositoryId: string
}) =>
  api<TImageRepository>({ orgId, path: `image-repositories/${repositoryId}` })

export const createImageRepository = ({
  orgId,
  body,
}: {
  orgId: string
  body: TCreateImageRepositoryRequest
}) =>
  api<TImageRepository>({
    orgId,
    path: 'image-repositories',
    method: 'POST',
    body,
  })

export const deleteImageRepository = ({
  orgId,
  repositoryId,
}: {
  orgId: string
  repositoryId: string
}) =>
  api<void>({
    orgId,
    path: `image-repositories/${repositoryId}`,
    method: 'DELETE',
  })

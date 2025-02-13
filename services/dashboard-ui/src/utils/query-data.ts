import { API_URL } from '@/utils/configs'
import { getFetchOpts } from '@/utils/get-fetch-opts'

export interface IQueryData {
  errorMessage?: string
  orgId?: string
  path: string
  pathVersion?: 'v1'
}

export async function queryData<T>({
  errorMessage = 'Encountered an issue retrieving this information, please refresh the page to try again.',
  orgId,
  path,
  pathVersion = 'v1',
}: IQueryData): Promise<T> {
  const res = await fetch(
    `${API_URL}/${pathVersion}/${path}`,
    await getFetchOpts(orgId, {}, 10000)
  )

  if (!res.ok) {
    throw new Error(errorMessage)
  }

  return res.json()
}

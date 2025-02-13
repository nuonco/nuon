import { API_URL } from '@/utils/configs'
import { getFetchOpts } from '@/utils/get-fetch-opts'

export interface IMutateData {
  data?: Record<string, unknown>
  errorMessage?: string
  method?: 'POST' | 'PATCH' | 'DELETE'
  orgId?: string
  path: string
  pathVersion?: 'v1'
}

export async function mutateData<T>({
  data,
  errorMessage = 'Encountered an issue retrieving this information, please refresh the page to try again.',
  method = 'POST',
  orgId,
  path,
  pathVersion = 'v1',
}: IMutateData): Promise<T> {
  const res = await fetch(`${API_URL}/${pathVersion}/${path}`, {
    ...(await getFetchOpts(orgId, {}, 10000)),
    body: data && JSON.stringify(data),
    method,
  })

  if (!res.ok) {
    throw new Error(errorMessage)
  }

  return res.json()
}

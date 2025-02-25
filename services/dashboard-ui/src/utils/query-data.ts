import { API_URL } from '@/utils/configs'
import { getFetchOpts } from '@/utils/get-fetch-opts'

export type TResponseError = {
  description: string
  error: string
  user_error: boolean
}

export type TQueryError = {
  errorMessage: string
  status: Response['status']
  error: TResponseError
}

export interface IQueryData {
  errorMessage?: string
  orgId?: string
  path: string
  pathVersion?: 'v1'
  abortTimeout?: number
}

export async function queryData<T>({
  errorMessage = 'Encountered an issue retrieving this information, please refresh the page to try again.',
  orgId,
  path,
  pathVersion = 'v1',
  abortTimeout = 10000,
}: IQueryData): Promise<T> {
  const res = await fetch(
    `${API_URL}/${pathVersion}/${path}`,
    await getFetchOpts(orgId, {}, abortTimeout)
  )

  if (!res.ok) {
    throw new Error(errorMessage)
    // const error = await res.json()

    // if (res.status >= 500) {

    // } else {
    //   return {
    //     status: res.status,
    //     error,
    //     errorMessage,
    //   }
    // }
  }

  return res.json()
}

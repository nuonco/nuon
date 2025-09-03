import { API_URL, getFetchOpts } from '@/utils'

export async function getAPIVersion(): Promise<{
  git_ref: string
  version: string
}> {
  const data = await fetch(`${API_URL}/version`, await getFetchOpts()).catch(
    (error) => {
      console.error(error)
      return {
        ok: false,
      } as Response
    }
  )

  if (!data?.ok) {
    throw new Error('Failed to fetch api version')
  }

  return data && data?.json()
}
import type { TOrg } from '@/types'
import { mutateData } from '@/utils'

export async function postOrg(data: { name: string }) {
  return mutateData<TOrg>({
    data,
    errorMessage:
      'Unable to create your organization, refresh the page and try again.',
    path: 'orgs',
  })
}
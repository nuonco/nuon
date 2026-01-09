'use server'

import { revalidatePath } from 'next/cache'
import { getSession } from '@/lib/auth-server'
import { ADMIN_API_URL } from '@/configs/api'

export async function updateOrgFeature(
  orgId: string,
  formData: FormData,
  list: Array<string>,
  revalidatePathString?: string
) {
  const data = Object.fromEntries(formData)
  const features = data['all']
    ? { all: true }
    : list.reduce((acc, feat) => {
        acc[feat] = data.hasOwnProperty(feat)
        return acc
      }, {})
  const session = await getSession()
  const { user } = session || {}

  try {
    const result = await fetch(
      `${ADMIN_API_URL}/v1/orgs/${orgId}/admin-features`,
      {
        method: 'PATCH',
        body: JSON.stringify({ features }),
        headers: {
          'Content-Type': 'application/json',
          'X-Nuon-Admin-Email': user?.email,
        },
      }
    ).then((r) => r.json())
    
    // Revalidate the path if provided
    if (revalidatePathString) {
      revalidatePath(revalidatePathString)
    }
    
    return { status: 201, result }
  } catch (error) {
    throw new Error('Unable to update org features')
  }
}
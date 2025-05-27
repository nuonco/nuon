import { GITHUB_APP_NAME } from '@/utils'
import { redirect } from 'next/navigation'

export async function GET(request: any) {
  // You can add logic here to determine the redirect URL
  const { searchParams } = new URL(request.url)
  const orgId = searchParams.get('org_id')
  const externalUrl = `https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${orgId}`

  // Redirect to the external URL
  redirect(externalUrl)
}

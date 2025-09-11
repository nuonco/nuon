import { GITHUB_APP_NAME } from '@/configs/github-app'
import { redirect } from 'next/navigation'

export async function GET(request: any) {
  const { searchParams } = new URL(request.url)
  const orgId = searchParams.get('org_id')
  const externalUrl = `https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${orgId}`

  redirect(externalUrl)
}

import { redirect } from 'next/navigation'

export default async function InstallBreakGlass({ params }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  redirect(`/${orgId}/installs/${installId}/break-glass/generate-stack`)
}

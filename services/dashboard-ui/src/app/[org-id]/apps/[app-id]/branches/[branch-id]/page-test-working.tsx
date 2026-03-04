import type { TPageProps } from '@/types'

type TBranchPageProps = TPageProps<'org-id' | 'app-id' | 'branch-id'>

export default async function AppBranchDetailPageTest({
  params,
}: TBranchPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params

  return (
    <div style={{ padding: '2rem' }}>
      <h1>✅ Branch Page Test</h1>
      <p>Org ID: {orgId}</p>
      <p>App ID: {appId}</p>
      <p>Branch ID: {branchId}</p>
      <p>If you see this, routing is working correctly!</p>
    </div>
  )
}

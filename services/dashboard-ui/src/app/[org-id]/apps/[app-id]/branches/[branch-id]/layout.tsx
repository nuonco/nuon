import type { TLayoutProps } from '@/types'

interface IBranchLayout extends TLayoutProps<'org-id' | 'app-id' | 'branch-id'> {}

export default async function BranchLayout({
  children,
}: IBranchLayout) {
  // Simple pass-through layout
  return <>{children}</>
}
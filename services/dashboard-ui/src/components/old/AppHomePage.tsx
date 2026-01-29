'use client'

import Link from 'next/link'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { MainLayout } from '@/components/layout/MainLayout'
import { MainTopbar } from '@/components/layout/MainTopbar'
import { SidebarProvider } from '@/providers/sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { VERSION } from '@/utils'

const NoOrgs = () => (
  <EmptyState
    variant="diagram"
    emptyTitle="No organizations yet"
    emptyMessage="Complete the onboarding process to set up your first organization and start deploying applications."
    action={
      <Link href="/onboarding">
        <Button variant="primary">Begin Onboarding</Button>
      </Link>
    }
  />
)

export const AppHomePage = () => {
  return (
    <SidebarProvider initIsSidebarOpen={true}>
      <SurfacesProvider>
        <MainLayout
          versions={{
            api: { git_ref: 'unknown', version: 'unknown' },
            ui: { version: VERSION },
          }}
          hideOrgContent
        >
          <main className="flex flex-col h-screen w-full">
            <MainTopbar hideOrgSettings />
            <div className="flex-auto flex flex-col items-center justify-center overflow-y-auto">
              <NoOrgs />
            </div>
          </main>
        </MainLayout>
      </SurfacesProvider>
    </SidebarProvider>
  )
}

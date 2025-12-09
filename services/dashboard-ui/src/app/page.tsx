import { redirect } from 'next/navigation'
import { getOrgIdFromCookie } from '@/actions/orgs/org-session-cookie'
import { HomePageWithModal } from '@/components/old/HomePageWithModal'
import { AppHomePage } from '@/components/old/AppHomePage'
import { auth0 } from '@/lib/auth'
import { getOrgs, getOrgById } from '@/lib'

export default async function Home() {
  const session = await auth0.getSession()

  // If user doesn't have a session, show the HomePageWithModal
  if (!session) {
    return <HomePageWithModal showModal={false} />
  }

  // User has a session, check for org
  const orgIdFromCookie = await getOrgIdFromCookie()

  if (orgIdFromCookie) {
    // Check if the org from cookie exists using getOrgById
    const { data: org, error } = await getOrgById({ orgId: orgIdFromCookie })
    
    if (org && !error) {
      // Org exists, redirect to that org
      redirect(`/${orgIdFromCookie}/apps`)
    }
  }

  // Either no org cookie or org doesn't exist, fetch first org from getOrgs
  const { data: orgs } = await getOrgs({ limit:  1 })

  if (orgs && orgs.length > 0) {
    // Redirect to the first org
    redirect(`/${orgs[0].id}/apps`)
  } else {
    // No orgs available, show AppHomePage
    return <AppHomePage />
  }
}

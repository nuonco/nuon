import { redirect } from 'next/navigation'
import { getOrgIdFromCookie } from '@/actions/orgs/org-session-cookie'
import { HomePageWithModal } from '@/components/old/HomePageWithModal'
import { AppHomePage } from '@/components/old/AppHomePage'
import { auth0 } from '@/lib/auth'
import { getOrgs } from '@/lib'

export default async function Home() {
  const session = await auth0.getSession()

  if (session) {
    const { data: orgs } = await getOrgs()

    if (orgs && orgs?.length) {
      // User is authenticated and has organizations - redirect to app homepage
      const orgIdFromCookie = await getOrgIdFromCookie()
      const targetOrg = (orgIdFromCookie && orgs.find(org => org.id === orgIdFromCookie)) 
        ? orgIdFromCookie 
        : orgs[0].id
      
      redirect(`/${targetOrg}/apps`)
    } else {
      // User is authenticated but has no organizations - show app homepage with journey modal
      return <AppHomePage />
    }
  }

  // User is not authenticated - show marketing landing page
  return <HomePageWithModal showModal={false} />
}

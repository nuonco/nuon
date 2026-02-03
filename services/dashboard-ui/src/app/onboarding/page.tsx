import { redirect } from 'next/navigation'
import { getSession } from '@/lib/auth-server'
import { OnboardingPageContent } from './onboarding-page-content'

export const dynamic = 'force-dynamic'

export default async function OnboardingPage() {
  const session = await getSession()
  if (!session) {
    redirect('/')
  }
  return <OnboardingPageContent />
}

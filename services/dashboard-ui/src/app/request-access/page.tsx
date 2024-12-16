import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DashboardHeader, SignUpForm } from '@/components'

export default withPageAuthRequired(
  async function GettingStarted() {
    return (
      <div className="h-100vh p-6 max-w-6xl m-auto">
        <DashboardHeader />
        <SignUpForm />
      </div>
    )
  },
  { returnTo: '/' }
)

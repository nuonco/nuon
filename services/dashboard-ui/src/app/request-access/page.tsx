import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { Dashboard, SignUpForm } from '@/components'

export default withPageAuthRequired(
  async function GettingStarted() {
    return (
      <Dashboard>
        <SignUpForm />
      </Dashboard>
    )
  },
  { returnTo: '/' }
)

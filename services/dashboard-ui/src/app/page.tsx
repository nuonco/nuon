import { Dashboard, Heading, Link, Page } from '@/components'

export default function Home() {
  return (
    <Dashboard>
      <Page
        heading={
          <Heading level={1} variant="title">
            Login to get started with Nuon
          </Heading>
        }
      >
        <main className="flex flex-col gap-8">
          <Link href="/api/auth/login?returnTo=/dashboard">Login</Link>
        </main>
      </Page>
    </Dashboard>
  )
}

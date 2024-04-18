import { Dashboard, Heading, Page } from '@/components'

export default function Home() {
  return (
    <Dashboard>
      <Page
        header={
          <Heading level={1} variant="title">
            Login to get started with Nuon
          </Heading>
        }
      >
        <main className="flex flex-col gap-8">
          <a
            className="flex flex-initial items-center w-fit gap-1 text-fuchsia-700 hover:text-fuchsia-600 focus:text-fuchsia-600 active:text-fuchsia-800 dark:text-fuchsia-500 dark:hover:text-fuchsia-400 dark:focus:text-fuchsia-400 dark:active:text-fuchsia-600"
            href="/api/auth/login?returnTo=/dashboard"
          >
            Login
          </a>
        </main>
      </Page>
    </Dashboard>
  )
}

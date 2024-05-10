import { GoArrowRight } from 'react-icons/go'
import { Dashboard, Heading, Text } from '@/components'

export default function Home() {
  return (
    <Dashboard>
      <main className="flex h-full gap-6">
        <div className="flex flex-col gap-6 p-0 md:p-12 lg:p-24 lg:max-w-2xl">
          <Heading variant="title">BYOC for everyone.</Heading>
          <Text variant="base">
            Offer Bring Your Own Cloud in minutes, unlocking new customers,
            product capabilities and revenue.
          </Text>
          <a
            className="flex flex-initial items-center w-fit gap-1 bg-gradient-to-r
hover:scale-105 focus:scale-105 active:scale-95 from-indigo-500 via-purple-500 to-fuchsia-400 text-slate-50 dark:text-slate-950 drop-shadow-sm px-4 py-2 rounded-sm text-xl transition-transform duration-75"
            href="/api/auth/login?returnTo=/dashboard"
          >
            Login to get started <GoArrowRight />
          </a>
        </div>
        {/* <div className="flex flex-col gap-6 p-12 md:p-24">graphic</div> */}
      </main>
    </Dashboard>
  )
}

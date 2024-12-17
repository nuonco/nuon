import { GoArrowRight } from 'react-icons/go'
import { Dashboard, Text } from '@/components'

export default function Home() {
  return (
    <Dashboard>
      <main className="flex flex-col md:flex-row h-full gap-6">
        <div className="flex flex-col gap-6 p-0 md:p-12 lg:p-24 lg:max-w-2xl">
          <Text className="!text-[42px] !leading-[42px]" variant="semi-18" level={1}>Bring Your Own Cloud, <br />for everyone.</Text>
          <Text className="!text-xl !leading-loose">
            If you already have an account, please log in. Otherwise you will be directed to request access.
          </Text>
          <a
            className="flex flex-initial items-center w-fit gap-1 bg-gradient-to-r
hover:scale-105 focus:scale-105 active:scale-95 from-indigo-500 via-purple-500 to-fuchsia-400 text-slate-50 dark:text-slate-950 drop-shadow-sm px-4 py-2 rounded-md text-lg transition-transform duration-75"
            href="/api/auth/login?returnTo=/"
          >
            Login / request access <GoArrowRight />
          </a>
        </div>
        <div className="flex-col gap-6 p-12 md:p-24 hidden 2xl:!flex">
          <img className="w-full max-w-[600px] rounded shadow" src="https://website-v2-nuonco.vercel.app/_astro/about-diagram.BDte5ucH_Z1z1dOE.webp" />
        </div>
      </main>
    </Dashboard>
  )
}

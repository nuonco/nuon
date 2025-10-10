import { type ReactNode } from 'react'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Logo } from '@/components/common/Logo/Logo'
import { Text } from '@/components/common/Text'

export const LandingLayout = ({ children }: { children: ReactNode }) => {
  return (
    <div className={'landing-gradient'}>
      <div
        className={'w-full h-dvh overflow-auto flex flex-col landing-graphic'}
      >
        <div className="flex flex-col gap-6 p-6 xl:px-24 w-full max-w-6xl mx-auto flex-auto">
          <header className="flex flex-wrap items-center justify-between gap-6 pb-6">
            <div className="flex items-center gap-6">
              <Logo />
            </div>

            <div className="flex flex-col">
              <div className="flex gap-4 items-center">
                <Link
                  href="https://docs.nuon.co"
                  target="_blank"
                  variant="ghost"
                >
                  Docs
                </Link>
                {/* <Suspense fallback="Loading...">
              <HeaderVersions />
              </Suspense>
              <SignOutButton /> */}
              </div>
            </div>
          </header>

          {children}
        </div>
        <footer className="bg-[#1D0B2F] text-cool-grey-50">
          <div className="flex flex-col md:flex-row gap-8 md:justify-between px-6 py-12 xl:px-24 max-w-6xl mx-auto">
            <div className="flex flex-col gap-3">
              <Logo id="dark-logo" variant="dark" />
              <Text variant="subtext">
                &copy; {new Date().getFullYear()} Nuon. All rights reserved.
              </Text>
              <div className="flex gap-4 mt-6">
                <a
                  href="https://x.com/nuoninc"
                  target="_blank"
                  className="text-lg hover:text-white/75"
                >
                  <Icon variant="XLogo" size={20} />
                </a>

                <a
                  href="https://www.linkedin.com/company/nuonco"
                  target="_blank"
                  className="text-lg hover:text-white/75"
                >
                  <Icon variant="LinkedinLogo" size={20} />
                </a>

                <a
                  href="https://github.com/nuonco"
                  target="_blank"
                  className="text-lg hover:text-white/75"
                >
                  <Icon variant="GithubLogo" size={20} />
                </a>
              </div>
            </div>

            <div className="flex flex-col md:flex-row gap-8 md:gap-16">
              <div>
                <Text variant="base" weight="strong" className="mb-4">
                  Product
                </Text>
                <div className="flex flex-col gap-4">
                  <Link
                    className="text-sm !text-cool-grey-50 hover:text-cool-grey-100 hover:underline"
                    target="_blank"
                    href="https://nuon.co/about"
                  >
                    About
                  </Link>
                  <Link
                    className="text-sm !text-cool-grey-50 hover:text-cool-grey-100 hover:underline"
                    target="_blank"
                    href="https://docs.nuon.co/pricing"
                  >
                    Pricing
                  </Link>
                  <Link
                    className="text-sm !text-cool-grey-50 hover:text-cool-grey-100 hover:underline"
                    target="_blank"
                    href="https://docs.nuon.co/get-started/introduction"
                  >
                    Docs
                  </Link>
                  <Link
                    className="text-sm !text-cool-grey-50 hover:text-cool-grey-100 hover:underline"
                    target="_blank"
                    href="https://nuon.co/blog"
                  >
                    Blog
                  </Link>
                </div>
              </div>
              <div>
                <Text variant="base" weight="strong" className="mb-4">
                  Legal
                </Text>
                <div className="flex flex-col gap-4">
                  <Link
                    className="text-sm !text-cool-grey-50 hover:text-cool-grey-100 hover:underline"
                    target="_blank"
                    href="https://nuon.co/terms"
                  >
                    Terms & confitions
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </footer>
      </div>
    </div>
  )
}

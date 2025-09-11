import classNames from 'classnames'
import React, { type FC, Suspense } from 'react'
import { XLogo, LinkedinLogo, GithubLogo } from '@phosphor-icons/react/dist/ssr'
import { Link } from '@/components/Link'
import { Logo } from '@/components/Logo'
import { BreadcrumbNav, type TLink } from '@/components/Nav'
import { NuonVersions, type TNuonVersions } from '@/components/NuonVersions'
import { SignOutButton } from '@/components/Profile'
import { ID, Text } from '@/components/Typography'
import { getAPIVersion } from '@/lib'
import { VERSION } from '@/utils'

const HeaderVersions = async () => {
  const { data: apiVersion } = await getAPIVersion()

  const versions = {
    api: apiVersion || { git_ref: 'unknown', version: 'unknown' },
    ui: {
      version: VERSION,
    },
  }

  return (
    <NuonVersions
      className="justify-end flex-col !flex-nowrap !gap-0"
      {...(versions as TNuonVersions)}
    />
  )
}

export const DashboardHeader: FC = () => {
  return (
    <header className="flex flex-wrap items-center justify-between gap-6 pb-6">
      <div className="flex items-center gap-6">
        <Logo />
      </div>

      <div className="flex flex-col">
        <div className="flex gap-4 items-center">
          <Link href="https://docs.nuon.co" target="_blank" variant="ghost">
            Docs
          </Link>          
          <SignOutButton />
        </div>
      </div>
    </header>
  )
}

export const Dashboard: FC<{
  children: React.ReactElement
  isLandingPage?: boolean
}> = ({ children, isLandingPage = false }) => {
  return (
    <div className={classNames('landing-gradient', {})}>
      <div
        className={classNames('w-full h-dvh overflow-auto flex flex-col', {
          'landing-graphic': isLandingPage,
        })}
      >
        <div className="flex flex-col gap-6 p-6 xl:px-24 w-full max-w-6xl mx-auto flex-auto">
          <DashboardHeader />

          {children}
        </div>
        <footer className="bg-[#1D0B2F] text-cool-grey-50">
          <div className="flex flex-col md:flex-row gap-8 md:justify-between px-6 py-12 xl:px-24 max-w-6xl mx-auto">
            <div className="flex flex-col gap-3">
              <Logo isDarkModeOnly />
              <Text variant="reg-12">
                &copy; {new Date().getFullYear()} Nuon. All rights reserved.
              </Text>
              <div className="flex gap-4 mt-6">
                <a
                  href="https://x.com/nuoninc"
                  target="_blank"
                  className="text-lg hover:text-white/75"
                >
                  <XLogo size={20} />
                </a>

                <a
                  href="https://www.linkedin.com/company/nuonco"
                  target="_blank"
                  className="text-lg hover:text-white/75"
                >
                  <LinkedinLogo size={20} />
                </a>

                <a
                  href="https://github.com/nuonco"
                  target="_blank"
                  className="text-lg hover:text-white/75"
                >
                  <GithubLogo size={20} />
                </a>
              </div>
            </div>

            <div className="flex flex-col md:flex-row gap-8 md:gap-16">
              <div>
                <Text variant="med-14" className="mb-4">
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
                <Text variant="med-14" className="mb-4">
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

export const DashboardContent: FC<{
  breadcrumb: Array<TLink>
  banner?: React.ReactNode | null
  children: React.ReactElement
  heading?: React.ReactElement | string
  headingUnderline?: React.ReactElement | string
  headingMeta?: React.ReactElement | string
  statues?: React.ReactElement | null
  meta?: React.ReactElement | null
}> = ({
  breadcrumb,
  banner = null,
  children,
  heading,
  headingUnderline,
  headingMeta,
  statues = null,
  meta = null,
}) => {
  return (
    <>
      {banner}
      <header className="flex justify-between items-center border-b px-6 py-4 h-[75px] bg-white/50 dark:bg-white/[.02]">
        <BreadcrumbNav links={breadcrumb} />
        <div>
          <Link href="https://docs.nuon.co" target="_blank" variant="ghost">
            Docs
          </Link>
        </div>
      </header>
      <main
        className="overflow-x-auto flex flex-col"
        style={{ height: 'calc(100% - 75px)' }}
      >
        {heading && (
          <header
            className={classNames(
              'px-6 pt-8 flex flex-col pt-6 gap-6 border-b',
              {
                'pb-8': !Boolean(meta),
              }
            )}
          >
            <div className="flex items-start justify-between gap-4">
              <hgroup className="flex flex-col gap-2">
                <Text level={1} role="heading" variant="semi-18">
                  {heading}
                </Text>
                <span>
                  <ID id={headingUnderline} />
                  {headingMeta ? (
                    <Text
                      className="text-blue-800 dark:text-blue-600"
                      variant="reg-12"
                    >
                      {headingMeta}
                    </Text>
                  ) : null}
                </span>
              </hgroup>

              {statues}
            </div>
            {meta}
          </header>
        )}

        {children}
      </main>
    </>
  )
}

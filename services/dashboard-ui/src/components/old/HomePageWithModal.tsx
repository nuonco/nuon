import React from 'react'
import Image from 'next/image'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { LogoLight } from '@/components/common/Logo/LogoLight'
import { LogoDark } from '@/components/common/Logo/LogoDark'
import { Text } from '@/components/common/Text'
import { USE_AUTH_SERVICE, AUTH_SERVICE_URL, APP_URL } from '@/configs/auth'
import authRightLight from '@/assets/auth-right-light.png'
import authRightDark from '@/assets/auth-right-dark.png'

interface HomePageWithModalProps {
  showModal: boolean
}

export const HomePageWithModal: React.FC<HomePageWithModalProps> = ({
  showModal,
}) => {
  const authUrl = USE_AUTH_SERVICE
    ? `${AUTH_SERVICE_URL}/?url=${APP_URL}`
    : '/api/auth/login?returnTo=/'

  return (
    <div className="flex h-screen w-full">
      {/* Left Side */}
      <div className="flex flex-col gap-10 justify-center w-full lg:w-[835px] bg-cool-grey-50 dark:bg-dark-grey-950 px-8 md:px-20 py-[60px]">
        {/* Card */}
        <div className="bg-white dark:bg-dark-grey-800 border border-[rgba(158,168,179,0.24)] dark:border-dark-grey-500 rounded-[10px] shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08),0px_10px_32px_-10px_rgba(0,0,0,0.08)] px-8 md:px-10 py-[60px] w-full flex flex-col gap-6">
          {/* Logo */}
          <a href="https://nuon.co" className="w-[100px] h-[42px] flex items-center">
            <span className="sr-only">Nuon</span>
            <LogoLight className="block dark:hidden shrink-0" />
            <LogoDark className="hidden dark:block shrink-0" />
          </a>

          {/* Heading and Subtitle */}
          <div className="flex flex-col gap-6">
            <Text
              role="heading"
              level={1}
              variant="h1"
              weight="strong"
              className="!tracking-[-0.65px]"
            >
              Start deploying to customer clouds.
            </Text>
            <Text role="paragraph" variant="base" weight="strong" theme="neutral">
              Create an account or sign in to manage your deployments. Get
              Started!
            </Text>
          </div>

          {/* Sign Up Button */}
          {!showModal && (
            <a
              href={authUrl}
              className="flex items-center justify-center w-full h-12 rounded-lg bg-primary-600 text-white text-lg leading-[27px] tracking-[-0.2px] font-medium shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08),inset_0px_0px_0px_2px_rgba(255,255,255,0.12)] hover:bg-primary-700 transition-colors"
            >
              Sign up
            </a>
          )}

          {/* Divider */}
          <hr className="border-[rgba(158,168,179,0.24)] dark:border-dark-grey-500" />

          {/* Already have account */}
          <div className="flex flex-col gap-6">
            <Text variant="h3" weight="strong" theme="neutral">
              Already have an account?
            </Text>

            {/* Sign In Button */}
            {!showModal && (
              <a
                href={authUrl}
                className="flex items-center justify-center w-full h-12 rounded-lg bg-white dark:bg-dark-grey-700 border border-[rgba(158,168,179,0.24)] dark:border-dark-grey-500 text-primary-600 dark:text-primary-400 text-lg leading-[27px] tracking-[-0.2px] font-medium shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)] hover:bg-cool-grey-50 dark:hover:bg-dark-grey-500 transition-colors"
              >
                Sign in
              </a>
            )}
          </div>
        </div>

        {/* Divider */}
        <hr className="border-[rgba(158,168,179,0.24)] dark:border-dark-grey-500" />

        {/* Customer Portal Section */}
        <div className="flex items-center gap-6 bg-cool-grey-50 dark:bg-dark-grey-800 border border-[rgba(158,168,179,0.24)] dark:border-dark-grey-500 rounded-lg p-6">
          <Text variant="h3" weight="strong" theme="neutral" className="w-3/5 shrink-0">
            Create white-label portals with real-time install visibility
          </Text>
          <div className="flex-1 flex items-center justify-end">
            <Button
              variant="secondary"
              href="https://customers.nuon.co"
              className="!h-auto !px-3 !py-2 !text-lg !leading-[27px] !tracking-[-0.2px] !rounded-lg shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)]"
            >
              Customer portal
              <Icon variant="ArrowLineUpRightIcon" size={16} weight="bold" />
            </Button>
          </div>
        </div>
      </div>

      {/* Right Side - Branded Background */}
      <div className="hidden lg:flex relative flex-1 overflow-hidden">
        <Image
          src={authRightLight}
          alt="Nuon branded background"
          fill
          className="object-cover block dark:hidden"
          priority
          quality={100}
          unoptimized
        />
        <Image
          src={authRightDark}
          alt="Nuon branded background"
          fill
          className="object-cover hidden dark:block"
          priority
          quality={100}
          unoptimized
        />
      </div>
    </div>
  )
}

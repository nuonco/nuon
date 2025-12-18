'use client'

import React from 'react'
import Image from 'next/image'
import { ArrowUpRight } from '@phosphor-icons/react'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { LogoLight } from '@/components/common/Logo/LogoLight'
import { LogoDark } from '@/components/common/Logo/LogoDark'
import ossHeroImage from '@/assets/oss-hero.png'

interface HomePageWithModalProps {
  showModal: boolean
}

export const HomePageWithModal: React.FC<HomePageWithModalProps> = ({
  showModal,
}) => {
  return (
    <div className="flex h-screen w-full">
      {/* Left Side */}
      <div className="flex flex-col gap-10 justify-center w-full lg:w-[840px] bg-cool-grey-50 dark:bg-dark-grey-950 px-8 md:px-20 py-16">
        {/* Card */}
        <div className="bg-white dark:bg-dark-grey-800 border border-cool-grey-500/25 dark:border-dark-grey-500 rounded-lg shadow-sm px-8 md:px-[70px] py-16 md:py-20 w-full flex flex-col gap-10">
          {/* Logo */}
          <a href="https://nuon.co" className="w-fit">
            <span className="sr-only">Nuon</span>
            <LogoLight className="block dark:hidden shrink-0" />
            <LogoDark className="hidden dark:block shrink-0" />
          </a>

          {/* Heading and Subtitle */}
          <div className="flex flex-col gap-6">
            <Text role="heading" level={1} variant="h1" weight="strong">
              Start deploying to customer clouds.
            </Text>
            <Text role="paragraph" variant="h3" theme="neutral">
              Create an account or sign in to manage your deployments. Get
              Started!
            </Text>
          </div>

          {/* Sign Up Button */}
          {!showModal && (
            <Button
              variant="primary"
              size="lg"
              href="/api/auth/login?returnTo=/"
              className="w-full justify-center"
            >
              Sign up
            </Button>
          )}

          {/* Divider */}
          <hr />

          {/* Already have account */}
          <Text role="heading" level={2} variant="h2" weight="strong">
            Already have an account?
          </Text>

          {/* Sign In Button */}
          {!showModal && (
            <Button
              variant="secondary"
              size="lg"
              href="/api/auth/login?returnTo=/"
              className="w-full justify-center"
            >
              Sign in
            </Button>
          )}
        </div>

        {/* Learn More Link */}
        <Button
          variant="ghost"
          size="lg"
          href="https://docs.nuon.co"
          className="text-primary-600 dark:text-primary-500"
        >
          Learn more about how Nuon works
          <ArrowUpRight size={16} weight="bold" />
        </Button>
      </div>

      {/* Right Side - Branded Background */}
      <div className="hidden lg:flex relative flex-1 overflow-hidden">
        <Image
          src={ossHeroImage}
          alt="Nuon branded background"
          fill
          className="object-cover"
          priority
        />
      </div>
    </div>
  )
}

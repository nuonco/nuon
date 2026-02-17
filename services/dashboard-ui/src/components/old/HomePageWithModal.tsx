import React from 'react'
import Image from 'next/image'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { LogoLight } from '@/components/common/Logo/LogoLight'
import { LogoDark } from '@/components/common/Logo/LogoDark'
import { Text } from '@/components/common/Text'
import { USE_AUTH_SERVICE, AUTH_SERVICE_URL, APP_URL } from '@/configs/auth'
import { LEDMarquee } from '@/components/old/LEDMarquee'
import ossHeroImage from '@/assets/oss-hero.png'

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
              href={authUrl}
              className="w-full justify-center"
              target="_self"
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
              href={authUrl}
              className="w-full justify-center"
              target="_self"
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
          <Icon variant="ArrowUpRightIcon" size={16} weight="bold" />
        </Button>
      </div>

      {/* Right Side - Branded Background with Demo Preview */}
      <div className="hidden lg:flex flex-col flex-1 overflow-hidden">
        {/* Image - 2/3 of height */}
        <div className="relative overflow-hidden" style={{ flex: '2 1 0%' }}>
          <Image
            src={ossHeroImage}
            alt="Nuon branded background"
            fill
            className="object-cover object-top"
            priority
          />
        </div>

        {/* Demo section - 1/3 of height */}
        <div className="flex flex-col" style={{ flex: '1 1 0%' }}>
          {/* LED Marquee Divider */}
          <LEDMarquee text="See What Your Customers See" />

          {/* Dark Demo CTA */}
          <div
            className="demo-section relative flex flex-col items-center justify-center gap-5 px-10 flex-1 overflow-hidden"
          >
            <h3 className="demo-heading">
              Experience the customer installer
            </h3>
            <p className="demo-subtext">
              See a live installation flow powered by Nuon — exactly what your end customers will use.
            </p>
            <a
              href="https://customers.nuon.co/admin/login/"
              target="_blank"
              rel="noopener noreferrer"
              className="demo-cta-btn"
            >
              Launch Live Demo
              <svg
                width="16"
                height="16"
                viewBox="0 0 16 16"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M4 12L12 4M12 4H5M12 4V11" />
              </svg>
            </a>
          </div>
          <style>{`
            .demo-section {
              background:
                radial-gradient(circle at 50% 0%, rgba(34, 211, 238, 0.06) 0%, transparent 70%),
                #0a0a0a;
            }
            .demo-section::before {
              content: '';
              position: absolute;
              inset: 0;
              background-image:
                linear-gradient(rgba(34, 211, 238, 0.04) 1px, transparent 1px),
                linear-gradient(90deg, rgba(34, 211, 238, 0.04) 1px, transparent 1px);
              background-size: 24px 24px;
              pointer-events: none;
            }
            .demo-heading {
              position: relative;
              color: #ffffff;
              font-family: monospace;
              font-size: 22px;
              font-weight: 700;
              text-align: center;
              letter-spacing: -0.5px;
              line-height: 1.3;
            }
            .demo-subtext {
              position: relative;
              color: rgba(34, 211, 238, 0.55);
              font-family: monospace;
              font-size: 14px;
              text-align: center;
              max-width: 360px;
              line-height: 1.5;
            }
            .demo-cta-btn {
              position: relative;
              display: inline-flex;
              align-items: center;
              justify-content: center;
              gap: 10px;
              background: rgba(34, 211, 238, 0.15);
              border: 1px solid rgba(34, 211, 238, 0.5);
              color: #22d3ee;
              font-family: monospace;
              font-size: 15px;
              font-weight: 600;
              text-transform: uppercase;
              letter-spacing: 0.5px;
              padding: 0 32px;
              height: 50px;
              transition: all 0.2s ease;
            }
            .demo-cta-btn:hover {
              background: rgba(34, 211, 238, 0.25);
              border-color: rgba(34, 211, 238, 0.8);
              box-shadow: 0 0 20px rgba(34, 211, 238, 0.15);
            }
          `}</style>
        </div>
      </div>
    </div>
  )
}

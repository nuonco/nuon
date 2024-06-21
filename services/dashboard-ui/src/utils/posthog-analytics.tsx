'use client'

import { useUser } from '@auth0/nextjs-auth0/client'
import React, { type FC, useEffect, useState } from 'react'
import { GoX } from 'react-icons/go'
import posthog from 'posthog-js'
import { Button, Heading, Text } from '@/components'

export const InitPosthogAnalytics: FC = () => {
  const { user, error, isLoading } = useUser()
  const [showOptOut, setShowOptOut] = useState<boolean>(false)

  useEffect(() => {
    const phOpt = window.localStorage.getItem('ph_opt')
    posthog.init(
      process?.env?.NEXT_PUBLIC_POSTHOG_TOKEN ||
        'phc_1NEQAphH0jxCX7opmp4Iyq2O6mu4tM552kMjJz8uKkl',
      {
        api_host:
          process?.env?.NEXT_PUBLIC_POSTHOG_HOST || 'https://us.i.posthog.com',
        person_profiles: 'identified_only',
      }
    )

    if (phOpt === null) {
      posthog.opt_out_capturing()
      setShowOptOut(true)
    } else if (phOpt === 'out') {
      posthog.opt_out_capturing()
    }
  }, [])

  useEffect(() => {
    if (user && !isLoading && window.localStorage.getItem('ph_opt') === 'in') {
      posthog.identify(user.sub, {
        email: user.email,
        nuon_id: user.sub,
        name: user.name,
      })
    }

    if (error || (!user && !isLoading)) {
      posthog.reset()
    }
  }, [user, error, isLoading])

  return (
    <>
      {showOptOut && (
        <div className="absolute bottom-0 bg-opacity-95 bg-slate-50 dark:bg-slate-950 z-20 border-b shadow-sm w-full flex items-start p-6">
          <div className="flex flex-col gap-4 items-start p-6 max-w-xl m-auto">
            <div>
              <Heading>Enhance your experience with Nuon</Heading>
              <Text variant="caption">
                We use cookies to improve your experience and enhance our
                app&apos;s performance. By clicking &quot;Accept,&quot; you
                consent to the use of cookies for product analytics, helping us
                understand how you use our app and make it even better for you.
              </Text>
            </div>
            <div className="flex gap-4">
              <Button
                variant="primary"
                onClick={() => {
                  window.localStorage.setItem('ph_opt', 'in')
                  posthog.opt_in_capturing()
                  setShowOptOut(false)
                }}
              >
                Accept
              </Button>
              <Button
                onClick={() => {
                  window.localStorage.setItem('ph_opt', 'out')
                  posthog.opt_out_capturing()
                  setShowOptOut(false)
                }}
              >
                Deny
              </Button>
            </div>
          </div>
          <Button
            className="rounded-full"
            variant="ghost"
            title="Opt out and close"
            onClick={() => {
              window.localStorage.setItem('ph_opt', 'out')
              posthog.opt_out_capturing()
              setShowOptOut(false)
            }}
          >
            <GoX />
          </Button>
        </div>
      )}
    </>
  )
}

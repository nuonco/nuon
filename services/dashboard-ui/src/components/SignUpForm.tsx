'use client'

import React, { type FC, FormEvent, useState } from 'react'
import { WarningCircle, Spinner } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Link } from '@/components/Link'
import { Text } from '@/components/Typography'
import { requestWaitlistAccess } from '@/components/org-actions'

export const SignUpForm: FC = () => {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [serverError, setServerError] = useState(null)
  const [successfulRequest, setSuccessfulRequest] = useState(false)

  return (
    <div className="pt-12 w-full max-w-md flex flex-col gap-6">
      {serverError ? (
        <div className="border border-red-300 bg-red-50 text-red-800 dark:border-red-600/30 dark:bg-red-600/10 dark:text-red-500 rounded p-3 w-full mb-2">
          <Text className="items-center" variant="reg-14">
            <WarningCircle />
            {serverError}
          </Text>
        </div>
      ) : null}
      {successfulRequest ? (
        <>
          <Text level={1} role="heading" variant="semi-18">
            {"We'll be in touch soon"}
          </Text>
          <Text className="!text-xl !leading-loose !inline" variant="reg-14">
            Thanks for requesting access, we are busy onboarding customers and
            will get to you as soon as possible. In the meantime, please join
            our{' '}
            <Link
              className="!inline"
              href="https://join.slack.com/t/nuoncommunity/shared_invite/zt-1q323vw9z-C8ztRP~HfWjZx6AXi50VRA"
              target="_blank"
            >
              slack community
            </Link>{' '}
            for more information.
          </Text>
        </>
      ) : (
        <>
          <Text
            className="!text-[32px]"
            level={1}
            role="heading"
            variant="semi-18"
          >
            Request Access Today
          </Text>
          <Text className="!text-xl !leading-loose" variant="reg-14">
            Fill out the form below and we
            will get back to you ASAP!
          </Text>
          <form
            className="flex flex-col gap-6 mt-4 mb-12"
            onSubmit={(e: FormEvent<HTMLFormElement>) => {
              e.preventDefault()
              setIsSubmitting(true)
              requestWaitlistAccess(new FormData(e.currentTarget))
                .then(() => {
                  setIsSubmitting(false)
                  setSuccessfulRequest(true)
                })
                .catch((err) => {
                  setServerError(err)
                })
            }}
          >
            <label className="flex flex-col gap-2">
              <Text variant="med-14">Title</Text>
              <input
                className="px-3 py-2 text-sm rounded border bg-black/5 dark:bg-transparent shadow-sm [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600/300"
                required
                name="job_title"
                placeholder="E.g. CTO, VP or Director of Engineering, Engineering manager..."
                type="text"
              />
            </label>

            <label className="flex flex-col gap-2">
              <Text variant="med-14">Company name</Text>
              <input
                className="px-3 py-2 text-sm rounded border bg-black/5 dark:bg-transparent shadow-sm [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600/300"
                required
                placeholder="Enter company name"
                name="org_name"
                type="text"
              />
            </label>

            <label className="flex flex-col gap-2">
              <Text variant="med-14">Tell us more</Text>
              <textarea
                className="px-3 py-2 text-sm rounded border bg-black/5 dark:bg-transparent shadow-sm [&:user-invalid]:border-red-300 [&:user-invalid]:dark:border-red-600/300 min-h-32"
                required
                name="tell_us_more"
                placeholder="What do you want to speak with Nuon about? Feel free to share your software application, infrastructure and cloud providers and how Nuon can help with that?"
              ></textarea>
            </label>

            <Button
              className="!inline-flex w-fit !text-base font-medium"
              variant="primary"
              disabled={isSubmitting}
            >
              {isSubmitting ? (
                <span className="flex items-center gap-2 justify-center">
                  <Spinner className="animate-[spin_3000ms_linear_infinite]" />{' '}
                  Requesting access...
                </span>
              ) : (
                'Request access'
              )}
            </Button>
          </form>
        </>
      )}
    </div>
  )
}

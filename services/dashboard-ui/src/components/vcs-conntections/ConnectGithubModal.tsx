'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { FaGithub } from 'react-icons/fa'
import { CaretLeft, Plus } from '@phosphor-icons/react'
import { Badge } from '@/components/Badge'
import { Button } from '@/components/Button'
import { Input } from '@/components/Input'
import { SpinnerSVG } from '@/components/Loading'
import { Link } from '@/components/Link'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { connectGitHubToOrg } from '@/components/org-actions'
import { GITHUB_APP_NAME } from '@/utils'
import { useOrg } from '@/components/Orgs/org-context'

interface IConnectGithubModal {}

export const ConnectGithubModal: FC<IConnectGithubModal> = ({}) => {
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [isManual, setIsManual] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>()

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-xl"
              contentClassName="!max-h-max"
              isOpen={isOpen}
              heading={
                <span className="flex items-center gap-3">
                  <FaGithub className="text-lg" /> Connect GitHub to Nuon
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
            >
              {isManual ? (
                <div>
                  <Button
                    className="text-sm flex items-center gap-2 mb-4 !p-1 !pl-0.5"
                    variant="ghost"
                    onClick={() => {
                      setIsManual(false)
                    }}
                  >
                    <CaretLeft />
                    Back
                  </Button>

                  <form
                    onSubmit={(e: React.FormEvent<HTMLFormElement>) => {
                      e.preventDefault()
                      setIsLoading(true)
                      const formData = Object.fromEntries(
                        new FormData(e.currentTarget)
                      ) as {
                        github_install_id: string
                        org_id: string
                      }

                      connectGitHubToOrg(formData)
                        .then(({ error }) => {
                          setIsLoading(false)
                          if (error) {
                            setError(
                              error?.error ||
                                'An error occured while trying to connect the GitHub account. Please try again.'
                            )
                          } else {
                            setIsManual(false)
                            setError(undefined)
                            setIsOpen(false)
                          }
                        })
                        .catch(() => {
                          setError(
                            'An error occured while trying to connect the GitHub account. Please try again.'
                          )
                        })
                    }}
                  >
                    <div className="flex flex-col gap-4">
                      {error ? <Notice>{error}</Notice> : null}
                      <Input
                        type="hidden"
                        name="org_id"
                        defaultValue={org?.id}
                        required
                      />
                      <label className="w-full flex flex-col gap-2">
                        <Text variant="med-14">GitHub install ID</Text>
                        <Input
                          placeholder="github_installation_id"
                          type="text"
                          name="github_install_id"
                          required
                        />
                      </label>
                    </div>
                    <div className="mt-6 flex gap-3 justify-end">
                      <Button
                        className="text-sm"
                        onClick={() => {
                          setIsManual(false)
                          setError(undefined)
                          setIsLoading(false)
                          setIsOpen(false)
                        }}
                        type="button"
                      >
                        Cancel
                      </Button>
                      <Button
                        className="text-sm flex items-center gap-2 font-medium"
                        disabled={isLoading}
                        variant="primary"
                      >
                        {isLoading ? (
                          <>
                            <SpinnerSVG /> Adding GitHub connection...
                          </>
                        ) : (
                          <>
                            <Plus size="16" /> Add GitHub connection
                          </>
                        )}
                      </Button>
                    </div>
                  </form>
                </div>
              ) : (
                <>
                  <div className="flex flex-col m-auto gap-8 mb-6">
                    <Link
                      className="flex flex-col items-center justify-center gap-4 border !p-8"
                      href={`https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${org.id}`}
                      variant="ghost"
                    >
                      <Text variant="med-14">New GitHub connection</Text>
                      <Text
                        className="!inline-block text-balance text-center !leading-loose"
                        variant="reg-14"
                      >
                        Add a new GitHub connection to your Nuon org by
                        installing the{' '}
                        <Badge
                          className="!inline-block !py-0.5 !px-1.5 !leading-normal"
                          variant="code"
                        >
                          {GITHUB_APP_NAME}
                        </Badge>{' '}
                        GitHub app and allowing access to the repositories of
                        your choice.
                      </Text>
                    </Link>

                    <div className="relative">
                      <hr />
                      <Text className="shadow-sm px-2 border w-fit rounded-lg bg-white text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50 absolute inset-0 m-auto h-[18px]">
                        OR
                      </Text>
                    </div>

                    <Button
                      className="flex flex-col items-center gap-4 !p-8"
                      onClick={() => {
                        setIsManual(true)
                      }}
                    >
                      <Text variant="med-14">Existing GitHub connection</Text>
                      <Text
                        className="!inline-block text-balance !leading-loose"
                        variant="reg-14"
                      >
                        Add an existing GitHub connection to your Nuon org by
                        manually entering the GitHub{' '}
                        <Badge
                          className="!inline-block !py-0.5 !px-1.5 !leading-normal"
                          variant="code"
                        >
                          github_install_id
                        </Badge>
                      </Text>
                    </Button>
                  </div>
                  <div className="flex gap-3 justify-end">
                    <Button
                      onClick={() => {
                        setIsOpen(false)
                      }}
                      className="text-sm"
                    >
                      Cancel
                    </Button>
                  </div>
                </>
              )}
            </Modal>,
            document.body
          )
        : null}

      <Button
        className="text-sm !font-medium !py-2 !pr-2 !pl-1 h-[32px] flex items-center gap-2 w-fit !text-primary-600 dark:!text-primary-400"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <Plus className="text-lg" />
        Add
      </Button>
    </>
  )
}

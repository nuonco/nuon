'use client'

import React, { type FC, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/Button'
import { CheckboxInput } from '@/components/Input'
import { Loading, SpinnerSVG } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { updateOrgFeature } from '@/components/admin-actions'
import type { TOrg } from '@/types'

export const AdminOrgFeatures: FC<{ org: TOrg }> = ({ org }) => {
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [featuresList, setFeaturesList] = useState([])
  const [error, setError] = useState()
  const router = useRouter()

  useEffect(() => {
    fetch(`/api/${org?.id}/features`)
      .then((res) =>
        res.json().then((feats) => {
          setIsLoading(false)
          setFeaturesList(feats)
        })
      )
      .catch((err) => {
        setIsLoading(false)
        setError(err?.message || 'Unable to fetch org features list')
      })
  }, [isOpen])

  return (
    <>
      <Modal
        className="!max-w-3xl"
        heading="Org features"
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false)
        }}
      >
        <div className="flex flex-col gap-3">
          {error ? <Notice>{error}</Notice> : null}
          {isLoading ? (
            <Loading loadingText="Loading org features..." variant="stack" />
          ) : featuresList?.length ? (
            <form
              onSubmit={(e: React.FormEvent<HTMLFormElement>) => {
                e.preventDefault()
                setIsSubmitting(true)
                const formData = new FormData(e.currentTarget)

                updateOrgFeature(org?.id, formData, featuresList)
                  .then(() => {
                    setIsSubmitting(false)
                    router.refresh()
                  })
                  .catch((err) => {
                    setIsSubmitting(false)
                    setError(err)
                  })
              }}
              className="flex flex-col gap-2"
            >
              <div className="w-fit">
                <CheckboxInput
                  labelText="All features"
                  name="all"
                  defaultChecked={Object.keys(org?.features).every(
                    (key) => org.features?.[key]
                  )}
                />
              </div>
              <div className="grid grid-cols-4">
                {featuresList.map((feature) => (
                  <CheckboxInput
                    key={feature}
                    labelText={feature}
                    name={feature}
                    defaultChecked={org?.features?.[feature]}
                  />
                ))}
              </div>
              <div className="flex items-center gap-3 self-end">
                <Button
                  className="text-sm"
                  type="button"
                  onClick={() => {
                    setIsOpen(false)
                  }}
                >
                  Cancel
                </Button>
                <Button
                  className="text-sm"
                  variant="primary"
                  disabled={isSubmitting}
                  type="submit"
                >
                  {isSubmitting ? (
                    <span className="flex items-center gap-3">
                      <SpinnerSVG /> Updating
                    </span>
                  ) : (
                    'Update'
                  )}
                </Button>
              </div>
            </form>
          ) : null}
        </div>
      </Modal>
      <div className="flex flex-col gap-2">
        <Text variant="reg-14">Manage org features</Text>
        <Button
          className="text-base w-full"
          onClick={() => {
            setIsOpen(true)
          }}
        >
          Manage org features
        </Button>
      </div>
    </>
  )
}

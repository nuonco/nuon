'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { Button } from '@/components/Button'
import { Link } from '@/components/Link'
import { Modal } from '@/components/Modal'
import { useOrg } from '@/components/Orgs'
import { Text } from '@/components/Typography'
// eslint-disable-next-line import/no-cycle
import type { TComponent } from '@/types'

export const ComponentDependencies: FC<{
  deps: Array<TComponent>
  name: string
  installId?: string
}> = ({ deps, name, installId }) => {
  return deps?.length > 2 ? (
    <MultiDependencies installId={installId} deps={deps} name={name} />
  ) : (
    deps.map((dep, i) => (
      <DependencyLink key={`${dep.id}-${i}`} dep={dep} installId={installId} />
    ))
  )
}

const DependencyLink: FC<{ dep: TComponent; installId?: string }> = ({
  dep,
  installId,
}) => {
  const { org } = useOrg()
  const path = installId
    ? `installs/${installId}/components/${dep.id}`
    : `apps/${dep?.app_id}/components/${dep?.id}`
  return (
    <Text className="bg-gray-500/10 px-2 py-1 rounded-lg border w-fit">
      <Link href={`/${org.id}/${path}`}>{dep?.name}</Link>
    </Text>
  )
}

const MultiDependencies: FC<{
  name: string
  deps: Array<TComponent>
  installId?: string
}> = ({ deps, name, installId }) => {
  const [isOpen, setIsOpen] = useState(false)
  const firstDeps = deps.slice(0, 2)
  const remainingDeps = deps.slice(2)

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-4xl"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
              heading={`${name} dependencies`}
              contentClassName="!p-0"
            >
              <div className="flex flex-wrap gap-4 px-6 py-4">
                {deps.map((dep, i) => (
                  <DependencyLink
                    key={`${dep.id}-${i}`}
                    dep={dep}
                    installId={installId}
                  />
                ))}
              </div>
              <div className="flex justify-end px-6 py-4 border-t">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-base"
                >
                  Close
                </Button>
              </div>
            </Modal>,
            document.body
          )
        : null}
      {firstDeps.map((dep, i) => (
        <DependencyLink
          key={`${dep.id}-${i}`}
          dep={dep}
          installId={installId}
        />
      ))}
      <Button
        className="!px-2 !py-1 text-sm"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        + {remainingDeps?.length}
      </Button>
    </>
  )
}

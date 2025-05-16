'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Graphviz } from '@hpcc-js/wasm'
import { Button } from '@/components/Button'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { useOrg } from '@/components/Orgs'
import { Text, Code } from '@/components/Typography'

export const AppConfigGraph: FC<{ appId: string; configId: string }> = ({
  appId,
  configId,
}) => {
  const { org } = useOrg()
  const [graph, setGraph] = useState<string>()
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()
  const [svg, setSvg] = useState('')

  const fetchData = () => {
    fetch(`/api/${org?.id}/apps/${appId}/configs/${configId}/graph`).then((r) =>
      r.json().then((res) => {
        setIsLoading(false)
        if (res?.error) {
          setError(res?.error?.error)
        } else {
          setGraph(res.data)
        }
      })
    )
  }

  useEffect(() => {
    fetchData()
  }, [])

  useEffect(() => {
    const renderGraph = async () => {
      try {
        const graphviz = await Graphviz.load()
        const svgOutput = await graphviz.layout(graph, 'svg', 'dot')
        setSvg(svgOutput)
      } catch (error) {
        console.error('Error rendering Graphviz:', error)
      }
    }

    renderGraph()
  }, [graph])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="w-full max-w-[calc(100%-4rem)] mx-6 xl:mx-auto"
              heading={
                <span>
                  <Text variant="med-14">App component dependency graph</Text>
                </span>
              }
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-2 mb-6">
                <Text variant="reg-14">
                  Nuon automatically creates a graph of all of the components in
                  your application.
                </Text>

                <ul className="flex flex-col gap-1 list-disc pl-4">
                  <li className="text-sm max-w-xl">
                    Dependencies are from root to dependencies (so a red-arrow
                    from a to b, means that b depends on a, or that when a
                    changes, b would be updated when{' '}
                    <Code
                      className="!inline-block !align-middle !py-0 !text-sm"
                      variant="inline"
                    >
                      select-dependencies
                    </Code>{' '}
                    is true)
                  </li>
                  <li className="text-sm">
                    Blue nodes mean that the current config version has changes
                    to that component
                  </li>
                </ul>
              </div>
              {isLoading ? (
                <Loading
                  loadingText="Loading component graph..."
                  variant="stack"
                />
              ) : error ? (
                <Notice>{error}</Notice>
              ) : (
                <div className="max-w-full overflow-auto">
                  <div dangerouslySetInnerHTML={{ __html: svg }} />
                </div>
              )}
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        View dependency graph
      </Button>
    </>
  )
}

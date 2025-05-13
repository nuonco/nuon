'use client'

import React, { type FC, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Graphviz } from '@hpcc-js/wasm'
import { Button } from '@/components/Button'
import { Modal } from '@/components/Modal'
import { Text, Code } from '@/components/Typography'

export const AppConfigGraph: FC<{ graph: string }> = ({ graph }) => {
  const [isOpen, setIsOpen] = useState(false)
  const [svg, setSvg] = useState('')

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
              className="!max-w-fit"
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
              <div dangerouslySetInnerHTML={{ __html: svg }} />
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

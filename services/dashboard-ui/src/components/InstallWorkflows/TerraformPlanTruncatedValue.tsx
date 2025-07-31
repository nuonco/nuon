'use client'

import React, { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import YAML from 'yaml'
import { MagnifyingGlassPlus } from '@phosphor-icons/react'
import { Loading } from '../Loading'
import { Modal } from '../Modal'
import { Code } from '../Typography'

// Helper: strip leading/trailing quotes and replace escaped newlines
function cleanString(str: string): string {
  let s = str ?? ''
  if (s.startsWith('"') && s.endsWith('"')) {
    s = s.slice(1, -1)
  }
  // Replace double-escaped newlines with real newlines
  s = s.replace(/\\n/g, '\n')
  return s
}

// Only consider YAML if it parses to an object or array
function isStringYaml(str: string): boolean {
  try {
    const parsed = YAML.parse(str)
    return typeof parsed === 'object' && parsed !== null
  } catch {
    return false
  }
}

// Only consider JSON if it parses to an object or array
function isStringJson(str: string): boolean {
  try {
    const parsed = JSON.parse(str)
    return typeof parsed === 'object' && parsed !== null
  } catch {
    return false
  }
}

export const TruncateValue = ({
  title,
  children,
}: {
  title: string
  children: string
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const [loading, setLoading] = useState<boolean>(false)

  // Normalize child to string for display
  let child =
    typeof children === 'object'
      ? JSON.stringify(children, null, 2)
      : children?.toString() || 'null'

  // Clean up quotes and newlines for all logic!
  child = cleanString(child)

  // Check if it is YAML or JSON (but not scalar)
  const isYAML = isStringYaml(child)
  const isJSON = isStringJson(child)

  // Show modal for long value, or if it's JSON/YAML (and not scalar)
  const shouldShowModal = child.length > 100 || isYAML || isJSON

  useEffect(() => {
    let active = true
    if (isOpen) {
      setLoading(true)
      setTimeout(() => {
        if (active) setLoading(false)
      }, 600)
    }
    return () => {
      active = false
    }
  }, [isOpen, child])

  if (shouldShowModal) {
    return (
      <span>
        <span
          onClick={() => setIsOpen(true)}
          className="cursor-pointer inline-flex gap-2 transition-all items-center hover:bg-black/5 dark:hover:bg-white/5 p-[1px]"
          title={`Click to view ${title}`}
        >
          <span
            className={`truncate inline-block max-w-[550px] ${isYAML || isJSON ? 'font-mono' : ''}`}
          >
            {child}
          </span>
          <MagnifyingGlassPlus size="16" />
        </span>
        {isOpen &&
          createPortal(
            <Modal
              heading={title}
              isOpen={isOpen}
              onClose={() => setIsOpen(false)}
            >
              <div>
                {loading ? (
                  <div className="flex items-center justify-center h-40">
                    <Loading
                      variant="stack"
                      loadingText={`Loading ${title}...`}
                    />
                  </div>
                ) : (
                  <span>
                    <Code className="max-w-full" variant="preformated">
                      {isJSON
                        ? JSON.stringify(JSON.parse(child), null, 2)
                        : child}
                    </Code>
                  </span>
                )}
              </div>
            </Modal>,
            document.body
          )}
      </span>
    )
  }

  return <>{child}</>
}

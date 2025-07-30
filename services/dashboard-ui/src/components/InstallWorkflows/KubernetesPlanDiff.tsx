import React, { useState } from 'react'
import { Text } from '@/components/Typography'
import { CaretRightIcon } from '@phosphor-icons/react' // Replace with your actual icon library import

interface IDiff {
  group_version_kind: {
    Group: string
    Version: string
    Kind: string
  }
  name: string
  namespace: string
  before: string
  after: string
}

interface IKubernetesManifestDiffViewerProps {
  approvalContents: {
    plan: IDiff[]
  }
}

export const KubernetesManifestDiffViewer: React.FC<
  IKubernetesManifestDiffViewerProps
> = ({ approvalContents }) => {
  const [isExpanded, setIsExpanded] = useState(false)
  const plan = approvalContents.plan

  return (
    <div className="flex flex-col gap-2">
      {plan.map((diff, index) => (
        <div
          key={index}
          className="flex flex-col gap-4 p-4 border rounded-md bg-gray-50 dark:bg-gray-800"
        >
          <div className="flex flex-col gap-1">
            <Text variant="med-14" className="font-bold text-primary-500">
              {diff.group_version_kind.Group}/{diff.group_version_kind.Version}{' '}
              - {diff.group_version_kind.Kind}
            </Text>
            <Text
              variant="med-14"
              className="text-gray-900 dark:text-gray-100 font-medium flex justify-between items-center"
            >
              <span>
                {diff.name}
                <span className="text-sm text-gray-700 dark:text-gray-300 ml-2">
                  ({diff.namespace || 'default'})
                </span>
              </span>
            </Text>
          </div>
          <details
            className="p-2 bg-gray-100 dark:bg-gray-900 rounded-md overflow-hidden text-sm transition-all duration-300 ease-in-out"
            onToggle={(e) => {
              const details = e.currentTarget
              const content = details.querySelector(
                '.details-lines'
              ) as HTMLElement
              const icon = details.querySelector(
                '.details-arrow-icon'
              ) as HTMLElement

              if (!isExpanded) {
                setIsExpanded(true)
                content.style.maxHeight = `${content.scrollHeight}px`
                content.style.transition = 'max-height 0.3s ease-in-out'
                icon.style.transform = 'rotate(90deg)'
              } else {
                setIsExpanded(false)
                content.style.maxHeight = `0px`
                content.style.transition = 'max-height 0.3s ease-in-out'
                icon.style.transform = 'rotate(0deg)'
              }
            }}
          >
            <summary className="cursor-pointer text-gray-700 dark:text-gray-300 font-medium transition-all m-1 flex items-center">
              <Text variant="med-12">View Diff</Text>
              <span className="ml-auto">
                <CaretRightIcon
                  className="details-arrow-icon transition-transform duration-300 ease-in-out"
                  weight="regular"
                  size={20}
                />
              </span>
            </summary>
            <div className="details-lines max-h-0 overflow-hidden transition-all duration-300 ease-in-out">
              <pre>
                {(diff.before || '')
                  .trim()
                  .split('\n')
                  .map((line, i) => {
                    const afterLine =
                      (diff.after || '').trim().split('\n')[i] || ''
                    if (line !== afterLine) {
                      if (line === '') {
                        return null
                      }
                      return (
                        <div key={i} className="mb-1">
                          {line && (
                            <div className="bg-red-100 dark:bg-red-900 p-1 rounded">
                              <span className="text-red-600 dark:text-red-200">
                                {String(i + 1).padStart(3, ' ')} {line}
                              </span>
                            </div>
                          )}
                          {afterLine && (
                            <div className="bg-green-100 dark:bg-green-900 p-1 rounded mt-1">
                              <span className="text-green-600 dark:text-green-200">
                                {String(i + 1).padStart(3, ' ')} {afterLine}
                              </span>
                            </div>
                          )}
                        </div>
                      )
                    }
                    return (
                      <div
                        key={i}
                        className="bg-gray-50 dark:bg-gray-800 p-1 rounded mb-1"
                      >
                        <span className="text-gray-800 dark:text-gray-200">
                          {String(i + 1).padStart(3, ' ')} {line}
                        </span>
                      </div>
                    )
                  })}
                {(diff.before || '').trim() === '' &&
                  (diff.after || '')
                    .trim()
                    .split('\n')
                    .map((line, i) => (
                      <div
                        key={`after-${i}`}
                        className="bg-green-100 dark:bg-green-900 p-1 rounded mb-1"
                      >
                        {line && (
                          <span className="text-green-600 dark:text-green-200">
                            {String(i + 1).padStart(3, ' ')} {line}
                          </span>
                        )}
                      </div>
                    ))}
              </pre>
            </div>
          </details>
        </div>
      ))}
    </div>
  )
}

import React, { useState } from 'react'
import { Text } from '@/components/old/Typography'
import { CaretRightIcon } from '@phosphor-icons/react'

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
  const plan = approvalContents.plan
  const [expanded, setExpanded] = useState<number | null>(null)

  return (
    <div className="flex flex-col gap-2">
      {plan?.length ? (
        plan.map((diff, index) => {
          const isOpen = expanded === index
          return (
            <div
              key={index}
              className="flex flex-col gap-4 p-4 border rounded-md bg-cool-grey-50 dark:bg-dark-grey-200"
            >
              <div className="flex flex-col gap-1">
                <Text variant="med-14" className="font-bold text-blue-500">
                  {diff.group_version_kind.Group}/
                  {diff.group_version_kind.Version} -{' '}
                  {diff.group_version_kind.Kind}
                </Text>
                <Text
                  variant="med-14"
                  className="text-cool-grey-900 dark:text-cool-grey-100 font-medium flex justify-between items-center"
                >
                  <span>
                    {diff.name}
                    <span className="text-sm text-cool-grey-700 dark:text-cool-grey-400 ml-2">
                      ({diff.namespace || 'default'})
                    </span>
                  </span>
                </Text>
              </div>
              <button
                type="button"
                className="cursor-pointer text-cool-grey-700 dark:text-cool-grey-400 font-medium transition-all m-1 flex items-center w-full bg-cool-grey-100 dark:bg-dark-grey-400 rounded-md px-3 py-2"
                onClick={() => setExpanded(isOpen ? null : index)}
                aria-expanded={isOpen}
              >
                <Text variant="med-12">View Diff</Text>
                <span className="ml-auto">
                  <CaretRightIcon
                    className={`transition-transform duration-300 ease-in-out ${isOpen ? 'rotate-90' : ''}`}
                    weight="regular"
                    size={20}
                  />
                </span>
              </button>
              {isOpen && (
                <div className="details-lines overflow-auto transition-all duration-300 ease-in-out">
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
                                <div className="bg-red-100 dark:bg-red-900 p-0.5 rounded">
                                  <span className="text-red-600 dark:text-red-200 text-sm">
                                    {String(i + 1).padStart(3, ' ')} {line}
                                  </span>
                                </div>
                              )}
                              {afterLine && (
                                <div className="bg-green-100 dark:bg-green-900 p-0.5 rounded mt-1">
                                  <span className="text-green-600 dark:text-green-200 text-sm">
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
                            className="bg-cool-grey-200 dark:bg-dark-grey-700 p-0.5 rounded mb-1"
                          >
                            <span className="text-cool-grey-800 dark:text-cool-grey-200 text-sm">
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
                            className="bg-green-100 dark:bg-green-900 p-0.5 rounded mb-1"
                          >
                            {line && (
                              <span className="text-green-600 dark:text-green-200 text-sm">
                                {String(i + 1).padStart(3, ' ')} {line}
                              </span>
                            )}
                          </div>
                        ))}
                  </pre>
                </div>
              )}
            </div>
          )
        })
      ) : (
        <span>No plan data</span>
      )}
    </div>
  )
}

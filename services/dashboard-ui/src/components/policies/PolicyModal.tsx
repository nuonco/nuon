'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ModalBase } from '@/components/surfaces/Modal'

interface IPolicyModalProps {
  isOpen: boolean
  onClose: () => void
  orgId: string
  appId: string
  componentNameToId: Record<string, string>
  policy: {
    id: string
    name: string
    type: string
    engine: string
    components: string[]
    contents: string
    createdAt: string
  }
}

function formatPolicyType(type: string): string {
  return type
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function getCodeLanguage(engine: string): 'yaml' | 'hcl' {
  return engine === 'opa' ? 'hcl' : 'yaml'
}

function ComponentsDropdown({
  components,
  orgId,
  appId,
  componentNameToId,
}: {
  components: string[]
  orgId: string
  appId: string
  componentNameToId: Record<string, string>
}) {
  const [search, setSearch] = useState('')
  const [isOpen, setIsOpen] = useState(false)

  const filteredComponents = useMemo(() => {
    if (!search.trim()) return components
    return components.filter((comp) =>
      comp.toLowerCase().includes(search.toLowerCase())
    )
  }, [components, search])

  return (
    <div className="flex flex-col gap-2">
      <Text variant="subtext" weight="strong">
        Applied to components ({components.length}):
      </Text>
      <div className="relative">
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="flex w-full items-center justify-between gap-2 rounded-md border border-grey-200 bg-white px-3 py-2 text-left text-sm dark:border-dark-grey-700 dark:bg-dark-grey-900"
        >
          <span className="text-grey-600 dark:text-grey-400">
            Search {components.length} components...
          </span>
          <Icon variant={isOpen ? 'CaretUp' : 'CaretDown'} size="16" />
        </button>

        {isOpen && (
          <div className="absolute z-50 mt-1 w-full rounded-md border border-grey-200 bg-white shadow-lg dark:border-dark-grey-700 dark:bg-dark-grey-900">
            <div className="border-b border-grey-200 p-2 dark:border-dark-grey-700">
              <Input
                placeholder="Search components..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full"
                autoFocus
              />
            </div>
            <div className="max-h-48 overflow-y-auto p-2">
              {filteredComponents.length > 0 ? (
                <div className="flex flex-col gap-1">
                  {filteredComponents.map((comp) => {
                    const componentId = componentNameToId[comp]
                    return componentId ? (
                      <Link
                        key={comp}
                        href={`/${orgId}/apps/${appId}/components/${componentId}`}
                        className="flex items-center gap-2 rounded px-2 py-1.5 text-sm text-grey-700 hover:bg-grey-100 hover:text-grey-900 dark:text-grey-300 dark:hover:bg-dark-grey-800 dark:hover:text-grey-100"
                      >
                        <Icon variant="ArrowSquareOut" size="14" />
                        {comp}
                      </Link>
                    ) : (
                      <div
                        key={comp}
                        className="px-2 py-1.5 text-sm text-grey-600 dark:text-grey-400"
                      >
                        {comp}
                      </div>
                    )
                  })}
                </div>
              ) : (
                <Text
                  variant="subtext"
                  className="px-2 py-1.5 text-center italic"
                >
                  No components match "{search}"
                </Text>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export const PolicyModal = ({
  isOpen,
  onClose,
  orgId,
  appId,
  componentNameToId,
  policy,
}: IPolicyModalProps) => {
  return (
    <ModalBase
      isVisible={isOpen}
      onClose={onClose}
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="ShieldCheck" size="24" />
          {policy.name}
        </Text>
      }
      className="!max-w-4xl"
      size="3/4"
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-wrap items-center gap-3">
          <Badge theme="default" size="md">
            {formatPolicyType(policy.type)}
          </Badge>
          <Badge
            theme={policy.engine === 'kyverno' ? 'brand' : 'info'}
            size="md"
          >
            {policy.engine.toUpperCase()}
          </Badge>
          {policy.createdAt && (
            <Text variant="subtext" className="flex items-center gap-1">
              Created <Time time={policy.createdAt} format="relative" />
            </Text>
          )}
        </div>

        {policy.components &&
          policy.components.length > 0 &&
          !(policy.components.length === 1 && policy.components[0] === '*') && (
            <ComponentsDropdown
              components={policy.components}
              orgId={orgId}
              appId={appId}
              componentNameToId={componentNameToId}
            />
          )}

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <Text variant="subtext" weight="strong">
              Policy definition:
            </Text>
            <ClickToCopyButton
              textToCopy={policy.contents}
              className="w-fit"
            />
          </div>
          <CodeBlock
            language={getCodeLanguage(policy.engine)}
            className="max-h-[500px]"
            showLineNumbers
          >
            {policy.contents}
          </CodeBlock>
        </div>
      </div>
    </ModalBase>
  )
}

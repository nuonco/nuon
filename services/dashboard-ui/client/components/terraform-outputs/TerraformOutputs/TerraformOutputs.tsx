import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton, ClickToCopy } from '@/components/common/ClickToCopy'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import type { TKeyValue } from '@/types'

type TViewMode = 'grid' | 'json'

const STORAGE_KEY = 'nuon:terraform-outputs-view'

function getStoredViewMode(): TViewMode {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'grid' || stored === 'json') return stored
  } catch {}
  return 'grid'
}

export const TerraformOutputs = ({
  heading = 'Outputs',
  outputs,
}: {
  heading?: string
  outputs: Record<string, unknown>
}) => {
  const [viewMode, setViewMode] = useState<TViewMode>(getStoredViewMode)

  const updateViewMode = (mode: TViewMode) => {
    setViewMode(mode)
    try {
      localStorage.setItem(STORAGE_KEY, mode)
    } catch {}
  }

  const entries = Object.entries(outputs)
  const isFlat = entries.every(([, v]) => typeof v !== 'object' || v === null)

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <Text weight="strong">{heading}</Text>
        <div className="flex items-center gap-2">
          <ToggleButton
            value={viewMode}
            onChange={updateViewMode}
            options={[
              { value: 'grid', label: <Icon variant="ListDashesIcon" size={16} />, ariaLabel: 'Grid view' },
              { value: 'json', label: <Icon variant="BracketsCurlyIcon" size={16} />, ariaLabel: 'JSON view' },
            ]}
          />
          <ClickToCopyButton
            textToCopy={JSON.stringify(outputs, null, 2)}
            className="w-fit"
          />
        </div>
      </div>

      {viewMode === 'json' ? (
        <JSONViewer data={outputs} expanded={2} showDataTypes={false} className="w-full" />
      ) : isFlat ? (
        <KeyValueList values={toKeyValues(outputs)} />
      ) : (
        <Card className="!p-0 !gap-0 divide-y w-full">
          {entries.map(([key, value]) => (
            <SectionExpand key={key} name={key} value={value} />
          ))}
        </Card>
      )}
    </div>
  )
}

function toKeyValues(obj: Record<string, unknown>): TKeyValue[] {
  return Object.entries(obj).map(([k, v]) => {
    if (v === null || v === undefined) {
      return { key: k, value: '' }
    }
    if (Array.isArray(v)) {
      return { key: k, value: JSON.stringify(v), type: 'array' }
    }
    if (typeof v === 'object') {
      return { key: k, value: JSON.stringify(v), type: 'object' }
    }
    return { key: k, value: String(v) }
  })
}

const SectionContent = ({ value }: { value: unknown }) => {
  if (typeof value !== 'object' || value === null) {
    return (
      <Text variant="subtext" family="mono">
        <ClickToCopy>{String(value)}</ClickToCopy>
      </Text>
    )
  }

  if (Array.isArray(value)) {
    return (
      <CodeBlock language="json" showCopy>
        {JSON.stringify(value, null, 2)}
      </CodeBlock>
    )
  }

  return <KeyValueList values={toKeyValues(value as Record<string, unknown>)} />
}

const SectionExpand = ({
  name,
  value,
}: {
  name: string
  value: unknown
}) => {
  const badge = Array.isArray(value)
    ? `${value.length} items`
    : typeof value === 'object' && value !== null
      ? `${Object.keys(value).length} keys`
      : typeof value

  return (
    <Expand
      id={`tf-output-${name}`}
      headerClassName="px-6"
      heading={
        <div className="flex items-center justify-between w-full">
          <Text weight="strong">{name}</Text>
          <Badge variant="code" size="sm" theme="neutral">
            {badge}
          </Badge>
        </div>
      }
    >
      <div className="border-t bg-black/[0.01] dark:bg-white/[0.01] px-6 py-3">
        <SectionContent value={value} />
      </div>
    </Expand>
  )
}

import type { IEmptyState } from '@/components/common/EmptyState'
import {
  KeyValueList,
  type IKeyValueList,
} from '@/components/common/KeyValueList'
import { Tabs } from '@/components/common/Tabs'
import type { TOTELLog } from '@/types'
import { cn } from '@/utils/classnames'
import { objectToKeyValueArray } from '@/utils/data-utils'

type TAttributeType = 'log' | 'resource' | 'scope'

interface IAttributesDisplay
  extends Omit<IKeyValueList, 'emptyStateProps' | 'values'> {
  attributes?: Record<string, string>
  type: TAttributeType
}

const ATTRIBUTE_CONFIG: Record<TAttributeType, IEmptyState> = {
  log: {
    emptyTitle: 'No log attributes',
    emptyMessage: "Log line doesn't have any log attributes.",
  },
  resource: {
    emptyTitle: 'No resource attributes',
    emptyMessage: "Log line doesn't have any resource attributes.",
  },
  scope: {
    emptyTitle: 'No scope attributes',
    emptyMessage: "Log line doesn't have any scope attributes.",
  },
} as const

export const AttributesDisplay = ({
  className,
  attributes,
  type,
  ...props
}: IAttributesDisplay) => {
  return (
    <KeyValueList
      className={cn('py-4', className)}
      values={objectToKeyValueArray(attributes)}
      emptyStateProps={ATTRIBUTE_CONFIG[type]}
      {...props}
    />
  )
}

export const AttributesTabs = ({ log }: { log: TOTELLog }) => {
  return (
    <Tabs
      tabs={{
        logAttributes: (
          <AttributesDisplay attributes={log.log_attributes} type="log" />
        ),
        resourceAttributes: (
          <AttributesDisplay
            attributes={log.resource_attributes}
            type="resource"
          />
        ),
        scopeAttributes: (
          <AttributesDisplay attributes={log.scope_attributes} type="scope" />
        ),
      }}
    />
  )
}

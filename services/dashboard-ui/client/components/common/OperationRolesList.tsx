import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import type { TKeyValue } from '@/types'

interface IOperationRolesList {
  operationRoles?: Record<string, string> | null
  emptyMessage?: string
}

export const OperationRolesList = ({
  operationRoles,
  emptyMessage = 'No operation roles configured',
}: IOperationRolesList) => {
  if (!operationRoles || Object.keys(operationRoles).length === 0) {
    return (
      <Text variant="subtext" theme="neutral">
        {emptyMessage}
      </Text>
    )
  }

  const operationLabels: Record<string, string> = {
    provision: 'Provision',
    deprovision: 'Deprovision',
    deploy: 'Deploy',
    teardown: 'Teardown',
    reprovision: 'Reprovision',
    trigger: 'Trigger',
  }

  const keyValuePairs: TKeyValue[] = Object.entries(operationRoles).map(
    ([operation, role]) => ({
      key: operationLabels[operation] || operation,
      value: role,
      type: 'string',
    })
  )

  return <KeyValueList values={keyValuePairs} />
}

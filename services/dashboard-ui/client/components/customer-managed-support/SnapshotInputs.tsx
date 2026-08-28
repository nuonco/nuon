import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { InputValue } from '@/components/installs/management/InputValue'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import type { TCustomerManagedCapturedInput } from '@/lib/ctl-api/installs/customer-managed-support-snapshots'
import { getInputDisplayName } from '@/utils/install-utils'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

const capturedValue = (input: TCustomerManagedCapturedInput) => {
  if (input.value_status === 'redacted') {
    return (
      <Badge variant="code" theme="neutral">
        Redacted
      </Badge>
    )
  }
  if (input.value_status === 'embedded-in-bundle') {
    return (
      <Text variant="subtext" theme="neutral">
        Embedded in bundle
      </Text>
    )
  }
  if (!input.value_available)
    return <InputValue name={input.name} value={null} />
  return <InputValue name={input.name} value={input.value ?? ''} />
}

export const CustomerManagedSnapshotInputs = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const captured = snapshot?.snapshot?.current_inputs
  const inputs = captured?.inputs ?? []

  return (
    <CustomerManagedSnapshotContent>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Current inputs
          </Text>
          <Text variant="subtext" theme="neutral">
            Safe input metadata captured by the customer runner.
          </Text>
        </HeadingGroup>
        {captured?.observed_at ? (
          <span className="flex items-center gap-2">
            <Text variant="subtext" theme="neutral">
              Observed
            </Text>
            <Time
              time={captured.observed_at}
              format="relative"
              variant="subtext"
            />
          </span>
        ) : null}
      </div>

      {!captured ? (
        <EmptyState
          variant="table"
          emptyTitle="No inputs captured"
          emptyMessage="Inputs will appear after the customer captures a snapshot with a compatible runner."
        />
      ) : inputs.length === 0 ? (
        <EmptyState
          variant="table"
          emptyTitle="No inputs configured"
          emptyMessage="This bundle does not define install inputs."
        />
      ) : (
        <PropertyGrid
          align="start"
          columns={[
            { key: 'name', header: 'Name' },
            { key: 'value', header: 'Captured value' },
            { key: 'default', header: 'Default' },
            { key: 'type', header: 'Type' },
            { key: 'required', header: 'Required' },
          ]}
          gridTemplate="minmax(150px, 1.5fr) minmax(180px, 2fr) minmax(140px, 1.5fr) minmax(90px, 1fr) minmax(80px, max-content)"
          values={inputs.map((input) => ({
            ...input,
            name: (
              <span className="flex flex-col">
                <Text variant="subtext" weight="strong">
                  {getInputDisplayName(input.name)}
                </Text>
                {getInputDisplayName(input.name) !== input.name ? (
                  <Text variant="label" family="mono" theme="neutral">
                    {input.name}
                  </Text>
                ) : null}
                {input.description ? (
                  <Text variant="label" theme="neutral">
                    {input.description}
                  </Text>
                ) : null}
              </span>
            ),
            value: capturedValue(input),
            default: input.secret ? (
              <Badge variant="code" theme="neutral">
                Redacted
              </Badge>
            ) : (
              <InputValue name={input.name} value={input.default} />
            ),
            type: (
              <Badge variant="code" theme="neutral">
                {input.type || 'string'}
              </Badge>
            ),
            required: (
              <Icon variant={input.required ? 'CheckIcon' : 'MinusIcon'} />
            ),
          }))}
        />
      )}
    </CustomerManagedSnapshotContent>
  )
}

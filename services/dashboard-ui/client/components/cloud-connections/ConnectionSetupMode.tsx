import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'

export const ConnectionSetupMode = ({
  platform,
  existingPrincipal,
  onChange,
}: {
  platform: string
  existingPrincipal: boolean
  onChange: (existing: boolean) => void
}) =>
  platform !== 'azure' ? (
    <div className="flex flex-col items-start gap-3">
      <Text weight="strong">Identity setup</Text>
      <ToggleButton
        value={existingPrincipal ? 'existing' : 'new'}
        onChange={(value) => onChange(value === 'existing')}
        options={[
          {
            value: 'new',
            label: platform === 'gcp' ? 'New service account' : 'New role',
          },
          {
            value: 'existing',
            label:
              platform === 'gcp' ? 'Existing service account' : 'Existing role',
          },
        ]}
      />
      <Text variant="subtext" theme="neutral">
        {existingPrincipal
          ? platform === 'gcp'
            ? 'Reuse the service account and add federation for this connection. Existing IAM bindings stay in place.'
            : 'Add this connection’s exact trust statement to the existing role. Keep one Terraform state or CloudFormation stack responsible for that role; merge the generated trust contribution into its current owner rather than creating a second role resource.'
          : 'Create a dedicated principal for this connection. If it already exists, choose the existing identity instructions instead.'}
      </Text>
    </div>
  ) : null

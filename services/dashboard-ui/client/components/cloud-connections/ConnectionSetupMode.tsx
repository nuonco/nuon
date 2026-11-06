import { Label } from '@/components/common/form/Label'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Text } from '@/components/common/Text'

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
    <div className="flex flex-col items-start gap-2">
      <Label>Identity setup</Label>
      <div className="flex flex-col gap-1">
        <RadioInput
          name="connection-setup-mode"
          value="new"
          checked={!existingPrincipal}
          onChange={() => onChange(false)}
          labelProps={{
            labelText: platform === 'gcp' ? 'New service account' : 'New role',
          }}
        />
        <RadioInput
          name="connection-setup-mode"
          value="existing"
          checked={existingPrincipal}
          onChange={() => onChange(true)}
          labelProps={{
            labelText:
              platform === 'gcp' ? 'Existing service account' : 'Existing role',
          }}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        {existingPrincipal
          ? platform === 'gcp'
            ? 'Reuse the service account and add federation for this connection. Existing IAM bindings stay in place.'
            : 'Adds this connection’s trust statement to the existing role. Keep one Terraform state or CloudFormation stack owning that role.'
          : 'Create a dedicated principal for this connection. If it already exists, use the existing role or service account setup instead.'}
      </Text>
    </div>
  ) : null

import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { Text } from '@/components/common/Text'
import type { InstallFormApi } from './useInstallForm'

export const StackOnlyCheckbox = ({ form }: { form: InstallFormApi }) => (
  <form.Field name="stackOnly">
    {(field) => (
      <FormCheckbox
        field={field}
        labelProps={{
          className:
            'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-1 gap-3 max-w-none items-start',
          labelText: (
            <span className="flex flex-col gap-1">
              <Text variant="base" weight="stronger" className="!leading-none">
                Stack and runner only
              </Text>
              <Text
                variant="subtext"
                theme="neutral"
                className="!leading-none"
              >
                Provision the stack and runner, and stop there. The sandbox and
                components stay unprovisioned until you provision the install.
              </Text>
            </span>
          ),
          labelTextProps: { as: 'div' },
        }}
      />
    )}
  </form.Field>
)

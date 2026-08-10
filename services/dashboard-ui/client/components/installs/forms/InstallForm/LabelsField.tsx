import { Button } from '@/components/common/Button'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import type { InstallFormApi } from './useInstallForm'

interface ILabelsField {
  form: InstallFormApi
}

export const LabelsField = ({ form }: ILabelsField) => (
  <Expand
    id="install-labels"
    heading="Labels"
    headerClassName="!px-4 bg-code"
    className="border rounded-md"
  >
    <div className="flex flex-col gap-3 p-4 border-t">
      <Text variant="subtext" theme="neutral">
        Labels connect this install to app branch install groups.
      </Text>

      <form.Field name="labels" mode="array">
        {(labelsField) => (
          <>
            {labelsField.state.value.map((_, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <form.Field name={`labels[${idx}].key`}>
                  {(field) => (
                    <Input
                      placeholder="Key"
                      value={(field.state.value as string | undefined) ?? ''}
                      onChange={(e) => field.handleChange(e.target.value)}
                      className="flex-1"
                    />
                  )}
                </form.Field>
                <form.Field name={`labels[${idx}].value`}>
                  {(field) => (
                    <Input
                      placeholder="Value"
                      value={(field.state.value as string | undefined) ?? ''}
                      onChange={(e) => field.handleChange(e.target.value)}
                      className="flex-1"
                    />
                  )}
                </form.Field>
                <Button
                  variant="icon"
                  type="button"
                  onClick={() => labelsField.removeValue(idx)}
                >
                  <Icon variant="XIcon" size={14} />
                </Button>
              </div>
            ))}
            <Button
              variant="secondary"
              type="button"
              className="w-fit"
              onClick={() => labelsField.pushValue({ key: '', value: '' })}
            >
              <Icon variant="PlusIcon" size={14} />
              Add label
            </Button>
          </>
        )}
      </form.Field>
    </div>
  </Expand>
)

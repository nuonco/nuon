import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ICustomNestedStackEntry {
  name: string
  template_url: string
  index: number
  parameters?: Record<string, string>
}

interface IEditStackOverridesModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: any
  currentVpcUrl: string
  currentRunnerUrl: string
  currentCustomStacks: ICustomNestedStackEntry[]
  appDefaultVpcUrl: string
  appDefaultRunnerUrl: string
  onSubmit: (data: {
    vpc_nested_template_url?: string
    runner_nested_template_url?: string
    custom_nested_stacks?: ICustomNestedStackEntry[]
  }) => void
}

export const EditStackOverridesModal = ({
  isPending,
  error,
  currentVpcUrl,
  currentRunnerUrl,
  currentCustomStacks,
  appDefaultVpcUrl,
  appDefaultRunnerUrl,
  onSubmit,
  ...props
}: IEditStackOverridesModal) => {
  const [vpcUrl, setVpcUrl] = useState(currentVpcUrl)
  const [runnerUrl, setRunnerUrl] = useState(currentRunnerUrl)
  const [customStacks, setCustomStacks] = useState<ICustomNestedStackEntry[]>(
    currentCustomStacks.length > 0
      ? currentCustomStacks
      : []
  )

  const handleAddStack = () => {
    setCustomStacks([
      ...customStacks,
      { name: '', template_url: '', index: customStacks.length },
    ])
  }

  const handleRemoveStack = (idx: number) => {
    setCustomStacks(customStacks.filter((_, i) => i !== idx))
  }

  const handleStackChange = (
    idx: number,
    field: keyof ICustomNestedStackEntry,
    value: string | number
  ) => {
    setCustomStacks(
      customStacks.map((s, i) => (i === idx ? { ...s, [field]: value } : s))
    )
  }

  const handleSubmit = () => {
    const data: Parameters<typeof onSubmit>[0] = {}
    if (vpcUrl !== currentVpcUrl) {
      data.vpc_nested_template_url = vpcUrl
    }
    if (runnerUrl !== currentRunnerUrl) {
      data.runner_nested_template_url = runnerUrl
    }
    const validStacks = customStacks.filter((s) => s.name && s.template_url)
    if (
      validStacks.length > 0 ||
      (currentCustomStacks.length > 0 && customStacks.length === 0)
    ) {
      data.custom_nested_stacks = validStacks
    }
    onSubmit(data)
  }

  const inputClassName =
    'w-full rounded-md border border-cool-grey-300 bg-white px-3 py-2 text-sm dark:border-dark-grey-500 dark:bg-dark-grey-800 dark:text-cool-grey-100 focus:outline-none focus:ring-1 focus:ring-primary-500'

  return (
    <Modal
      size="lg"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="StackSimple" size="24" />
          Edit stack overrides
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving...
          </span>
        ) : (
          'Save overrides'
        ),
        onClick: handleSubmit,
        disabled: isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to save stack overrides'}
          </Banner>
        ) : null}

        <Banner theme="info">
          <Text variant="body">
            Override the default stack template URLs for this install. Leave empty to use the app-level default.
          </Text>
        </Banner>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <Text variant="subtext" weight="strong">VPC nested template URL</Text>
            <input
              type="text"
              className={inputClassName}
              value={vpcUrl}
              onChange={(e) => setVpcUrl(e.target.value)}
              placeholder={appDefaultVpcUrl || 'https://s3.amazonaws.com/bucket/vpc-template.yaml'}
            />
            {appDefaultVpcUrl && (
              <Text variant="subtext" theme="neutral">
                App default: {appDefaultVpcUrl}
              </Text>
            )}
          </div>

          <div className="flex flex-col gap-1">
            <Text variant="subtext" weight="strong">Runner nested template URL</Text>
            <input
              type="text"
              className={inputClassName}
              value={runnerUrl}
              onChange={(e) => setRunnerUrl(e.target.value)}
              placeholder={appDefaultRunnerUrl || 'https://s3.amazonaws.com/bucket/runner-template.yaml'}
            />
            {appDefaultRunnerUrl && (
              <Text variant="subtext" theme="neutral">
                App default: {appDefaultRunnerUrl}
              </Text>
            )}
          </div>
        </div>

        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <Text variant="subtext" weight="strong">Custom nested stacks</Text>
            <Button variant="ghost" size="sm" onClick={handleAddStack}>
              <Icon variant="Plus" />
              Add stack
            </Button>
          </div>

          {customStacks.length === 0 ? (
            <Text variant="subtext" theme="neutral">
              No custom nested stack overrides configured.
            </Text>
          ) : (
            <div className="flex flex-col gap-3">
              {customStacks.map((stack, idx) => (
                <div
                  key={idx}
                  className="flex flex-col gap-2 rounded-md border border-cool-grey-200 p-3 dark:border-dark-grey-600"
                >
                  <div className="flex items-center justify-between">
                    <Text variant="subtext" weight="strong">Stack {idx + 1}</Text>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRemoveStack(idx)}
                    >
                      <Icon variant="Trash" />
                    </Button>
                  </div>
                  <div className="grid grid-cols-[1fr_2fr_auto] gap-2">
                    <div className="flex flex-col gap-1">
                      <Text variant="subtext" theme="neutral">Name</Text>
                      <input
                        type="text"
                        className={inputClassName}
                        value={stack.name}
                        onChange={(e) =>
                          handleStackChange(idx, 'name', e.target.value)
                        }
                        placeholder="e.g. k8s_namespaces"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <Text variant="subtext" theme="neutral">Template URL</Text>
                      <input
                        type="text"
                        className={inputClassName}
                        value={stack.template_url}
                        onChange={(e) =>
                          handleStackChange(idx, 'template_url', e.target.value)
                        }
                        placeholder="https://s3.amazonaws.com/..."
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <Text variant="subtext" theme="neutral">Index</Text>
                      <input
                        type="number"
                        className={inputClassName + ' w-20'}
                        value={stack.index}
                        onChange={(e) =>
                          handleStackChange(idx, 'index', parseInt(e.target.value) || 0)
                        }
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </Modal>
  )
}

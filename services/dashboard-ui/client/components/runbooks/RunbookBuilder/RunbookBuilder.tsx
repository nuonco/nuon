import { useMemo, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks/get-runbooks'
import { downloadFileOnClick } from '@/utils/file-download'
import {
  importRunbook,
  operationLabels,
  serializeRunbook,
  serializeRunbookReadme,
  slugifyRunbookName,
  validateRunbook,
  type TBuilderOperation,
  type TBuilderStep,
  type TOption,
} from './helpers'

interface IRunbookBuilder {
  components: TOption[]
  actions: TOption[]
  runbooks: TRunbook[]
  loading?: boolean
  loadingError?: boolean
}

const operations = Object.keys(operationLabels) as TBuilderOperation[]

export function RunbookBuilder({
  components,
  actions,
  runbooks,
  loading,
  loadingError,
}: IRunbookBuilder) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [steps, setSteps] = useState<TBuilderStep[]>([])
  const [selected, setSelected] = useState<string>()
  const [runbookId, setRunbookId] = useState('')
  const [messages, setMessages] = useState<string[]>([])
  const [copied, setCopied] = useState<'toml' | 'markdown'>()
  const selectedStep = steps.find((step) => step.key === selected)
  const errors = validateRunbook(name, steps)
  const toml = useMemo(
    () => serializeRunbook(name, description, steps),
    [name, description, steps]
  )
  const hasDocumentation = steps.some((step) => step.documentation?.trim())
  const markdown = useMemo(
    () => serializeRunbookReadme(name, description, steps),
    [name, description, steps]
  )
  const update = (patch: Partial<TBuilderStep>) =>
    setSteps((current) =>
      current.map((step) =>
        step.key === selected ? { ...step, ...patch } : step
      )
    )
  const add = (operation: TBuilderOperation) => {
    const step = {
      key: crypto.randomUUID(),
      operation,
      name: operationLabels[operation],
    }
    setSteps((current) => [...current, step])
    setSelected(step.key)
  }
  const move = (index: number, delta: number) =>
    setSteps((current) => {
      const next = [...current]
      const target = index + delta
      if (target < 0 || target >= next.length) return current
      ;[next[index], next[target]] = [next[target], next[index]]
      return next
    })
  const importSelected = () => {
    const runbook = runbooks.find((item) => item.id === runbookId)
    if (!runbook) return
    const result = importRunbook(runbook, actions)
    setSteps((current) => [...current, ...result.steps])
    setMessages(
      result.errors.length
        ? result.errors
        : [`Imported ${result.steps.length} steps from ${runbook.name}.`]
    )
    if (result.steps[0]) setSelected(result.steps[0].key)
  }

  return (
    <div className="flex flex-col gap-6">
      {loadingError ? (
        <Banner theme="warn">
          Some app resources could not be loaded. You can still build with
          available options.
        </Banner>
      ) : null}
      <Card className="!p-4 !gap-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            id="runbook-name"
            labelProps={{ labelText: 'Runbook name' }}
            value={name}
            onChange={(event) => setName(event.target.value)}
            required
          />
          <Input
            id="runbook-description"
            labelProps={{ labelText: 'Description' }}
            value={description}
            onChange={(event) => setDescription(event.target.value)}
          />
        </div>
        <Text variant="subtext" theme="neutral">
          Suggested path: runbooks/{slugifyRunbookName(name)}.toml
        </Text>
        {hasDocumentation ? (
          <Text variant="subtext" theme="neutral">
            README path: runbooks/{slugifyRunbookName(name)}.md
          </Text>
        ) : null}
      </Card>

      <div className="grid grid-cols-1 xl:grid-cols-[minmax(13rem,1fr)_minmax(18rem,1.4fr)_minmax(16rem,1fr)] gap-4 items-start">
        <Card className="!p-4 !gap-4">
          <Text weight="strong">Workflow palette</Text>
          {operations.map((operation) => (
            <Button
              key={operation}
              variant="secondary"
              className="w-full justify-start"
              onClick={() => add(operation)}
            >
              <Icon variant="PlusIcon" size="16" />
              {operationLabels[operation]}
            </Button>
          ))}
          <Text weight="strong" className="mt-2">
            Import existing runbook steps
          </Text>
          <Text variant="subtext" theme="neutral">
            Appends a snapshot of the latest steps. It does not retain a
            reference.
          </Text>
          <Select
            options={runbooks.map((runbook) => ({
              value: runbook.id,
              label: runbook.name,
            }))}
            value={runbookId}
            placeholder={loading ? 'Loading runbooks' : 'Select a runbook'}
            onChange={(event) => setRunbookId(event.target.value)}
            disabled={loading || !runbooks.length}
          />
          <Button
            variant="secondary"
            disabled={!runbookId}
            onClick={importSelected}
          >
            Import steps
          </Button>
        </Card>

        <Card className="!p-4 !gap-4">
          <Text weight="strong">Ordered workflow</Text>
          {!steps.length ? (
            <Text variant="subtext" theme="neutral">
              Add an operation from the palette.
            </Text>
          ) : (
            steps.map((step, index) => (
              <div key={step.key} className="flex items-center gap-2">
                <Button
                  className="flex-1 justify-start"
                  variant={selected === step.key ? 'primary' : 'secondary'}
                  onClick={() => setSelected(step.key)}
                >
                  {index + 1}. {operationLabels[step.operation]}
                </Button>
                <Button
                  aria-label="Move step up"
                  size="sm"
                  disabled={index === 0}
                  onClick={() => move(index, -1)}
                >
                  <Icon variant="ArrowUpIcon" size="14" />
                </Button>
                <Button
                  aria-label="Move step down"
                  size="sm"
                  disabled={index === steps.length - 1}
                  onClick={() => move(index, 1)}
                >
                  <Icon variant="ArrowDownIcon" size="14" />
                </Button>
                <Button
                  aria-label="Duplicate step"
                  size="sm"
                  onClick={() =>
                    setSteps((current) =>
                      current.toSpliced(index + 1, 0, {
                        ...step,
                        key: crypto.randomUUID(),
                      })
                    )
                  }
                >
                  <Icon variant="CopyIcon" size="14" />
                </Button>
                <Button
                  aria-label="Remove step"
                  size="sm"
                  variant="danger"
                  onClick={() => {
                    setSteps((current) =>
                      current.filter((item) => item.key !== step.key)
                    )
                    if (selected === step.key) setSelected(undefined)
                  }}
                >
                  <Icon variant="TrashIcon" size="14" />
                </Button>
              </div>
            ))
          )}
        </Card>

        <Card className="!p-4 !gap-4">
          <Text weight="strong">Step settings</Text>
          {!selectedStep ? (
            <Text variant="subtext" theme="neutral">
              Select a step to configure it.
            </Text>
          ) : (
            <>
              <Text>{operationLabels[selectedStep.operation]}</Text>
              <Input
                id="step-name"
                labelProps={{ labelText: 'Step name' }}
                value={selectedStep.name}
                onChange={(event) => update({ name: event.target.value })}
                required
              />
              <Textarea
                id="step-documentation"
                labelProps={{ labelText: 'Step documentation (Markdown)' }}
                helperText="This appears under the step in the generated runbook README."
                placeholder="Explain what this step does and what operators should verify."
                value={selectedStep.documentation ?? ''}
                autoResize
                minRows={4}
                onChange={(event) =>
                  update({ documentation: event.target.value })
                }
              />
              {[
                'deploy-component',
                'check-component-drift',
                'tear-down-component',
              ].includes(selectedStep.operation) ? (
                <Select
                  labelProps={{ labelText: 'Component' }}
                  options={components.map((component) => ({
                    value: component.name,
                    label: component.name,
                  }))}
                  value={selectedStep.componentName ?? ''}
                  placeholder="Select a component"
                  searchable
                  onChange={(event) =>
                    update({ componentName: event.target.value })
                  }
                />
              ) : null}
              {['deploy-component', 'check-component-drift'].includes(
                selectedStep.operation
              ) ? (
                <CheckboxInput
                  checked={!!selectedStep.deployDependents}
                  onChange={(event) =>
                    update({ deployDependents: event.target.checked })
                  }
                  labelProps={{ labelText: 'Deploy dependent components' }}
                />
              ) : null}
              {selectedStep.operation === 'tear-down-component' ? (
                <CheckboxInput
                  checked={!!selectedStep.tearDownDependents}
                  onChange={(event) =>
                    update({ tearDownDependents: event.target.checked })
                  }
                  labelProps={{ labelText: 'Tear down dependent components' }}
                />
              ) : null}
              {selectedStep.operation === 'reprovision-sandbox' ? (
                <CheckboxInput
                  checked={!!selectedStep.skipComponentDeploys}
                  onChange={(event) =>
                    update({ skipComponentDeploys: event.target.checked })
                  }
                  labelProps={{ labelText: 'Skip component deploys' }}
                />
              ) : null}
              {selectedStep.operation === 'configured-action' ? (
                <Select
                  labelProps={{ labelText: 'Action' }}
                  options={actions.map((action) => ({
                    value: action.name,
                    label: action.name,
                  }))}
                  value={selectedStep.actionName ?? ''}
                  placeholder="Select an action"
                  searchable
                  onChange={(event) =>
                    update({ actionName: event.target.value })
                  }
                />
              ) : null}
              {selectedStep.operation === 'command' ? (
                <>
                  <Input
                    id="step-command"
                    labelProps={{ labelText: 'Command' }}
                    value={selectedStep.command ?? ''}
                    onChange={(event) =>
                      update({ command: event.target.value })
                    }
                  />
                  <Input
                    id="step-timeout"
                    labelProps={{ labelText: 'Timeout' }}
                    placeholder="10m"
                    value={selectedStep.timeout ?? ''}
                    onChange={(event) =>
                      update({ timeout: event.target.value })
                    }
                  />
                  <Input
                    id="step-role"
                    labelProps={{ labelText: 'Role' }}
                    value={selectedStep.role ?? ''}
                    onChange={(event) => update({ role: event.target.value })}
                  />
                </>
              ) : null}
            </>
          )}
        </Card>
      </div>

      {messages.map((message) => (
        <Banner
          key={message}
          theme={message.startsWith('Imported') ? 'success' : 'warn'}
        >
          {message}
        </Banner>
      ))}
      {errors.length ? (
        <Banner theme="error">
          <div>
            <Text weight="strong">Runbook is not ready to export</Text>
            {errors.map((error) => (
              <Text key={error} variant="subtext">
                {error}
              </Text>
            ))}
          </div>
        </Banner>
      ) : null}
      <Card className="!p-4 !gap-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Text weight="strong">TOML preview</Text>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="primary"
              disabled={!!errors.length}
              title={errors.length ? 'Cannot copy — resolve validation errors' : undefined}
              onClick={async () => {
                await navigator.clipboard.writeText(toml)
                setCopied('toml')
                window.setTimeout(() => setCopied(undefined), 1500)
              }}
            >
              {copied === 'toml' ? 'Copied' : 'Copy TOML'}
            </Button>
            <Button
              variant="secondary"
              disabled={!!errors.length}
              title={errors.length ? 'Cannot download — resolve validation errors' : undefined}
              onClick={() =>
                downloadFileOnClick({
                  content: toml,
                  filename: `${slugifyRunbookName(name)}.toml`,
                })
              }
            >
              Download .toml
            </Button>
          </div>
        </div>
        <CodeBlock language="toml" showCopy={errors.length === 0}>
          {toml}
        </CodeBlock>
      </Card>
      {hasDocumentation ? (
        <Card className="!p-4 !gap-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Text weight="strong">README preview</Text>
            <div className="flex flex-wrap gap-2">
              <Button
                variant="primary"
                disabled={!!errors.length}
                title={
                  errors.length
                    ? 'Cannot copy — resolve validation errors'
                    : undefined
                }
                onClick={async () => {
                  await navigator.clipboard.writeText(markdown)
                  setCopied('markdown')
                  window.setTimeout(() => setCopied(undefined), 1500)
                }}
              >
                {copied === 'markdown' ? 'Copied' : 'Copy Markdown'}
              </Button>
              <Button
                variant="secondary"
                disabled={!!errors.length}
                title={
                  errors.length
                    ? 'Cannot download — resolve validation errors'
                    : undefined
                }
                onClick={() =>
                  downloadFileOnClick({
                    content: markdown,
                    filename: `${slugifyRunbookName(name)}.md`,
                    mimeType: 'text/markdown',
                  })
                }
              >
                Download .md
              </Button>
            </div>
          </div>
          <CodeBlock language="markdown" showCopy={errors.length === 0}>
            {markdown}
          </CodeBlock>
        </Card>
      ) : null}
    </div>
  )
}

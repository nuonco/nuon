import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import type {
  TBuilderStep,
  TOption,
} from '@/components/runbooks/RunbookBuilder/helpers'

interface IStepEditor {
  step: TBuilderStep
  components: TOption[]
  actions: TOption[]
  onChange: (patch: Partial<TBuilderStep>) => void
}

export function StepEditor({ step, components, actions, onChange }: IStepEditor) {
  return (
    <div className="flex flex-col gap-3">
      <Input
        id={`step-name-${step.key}`}
        labelProps={{ labelText: 'Step name' }}
        value={step.name}
        onChange={(event) => onChange({ name: event.target.value })}
        required
      />
      {[
        'deploy-component',
        'check-component-drift',
        'tear-down-component',
      ].includes(step.operation) ? (
        <Select
          labelProps={{ labelText: 'Component' }}
          options={components.map((component) => ({
            value: component.name,
            label: component.name,
          }))}
          value={step.componentName ?? ''}
          placeholder="Select a component"
          searchable
          onChange={(value) => onChange({ componentName: value })}
        />
      ) : null}
      {['deploy-component', 'check-component-drift'].includes(
        step.operation
      ) ? (
        <CheckboxInput
          checked={!!step.deployDependents}
          onChange={(event) =>
            onChange({ deployDependents: event.target.checked })
          }
          labelProps={{ labelText: 'Deploy dependent components' }}
        />
      ) : null}
      {step.operation === 'tear-down-component' ? (
        <CheckboxInput
          checked={!!step.tearDownDependents}
          onChange={(event) =>
            onChange({ tearDownDependents: event.target.checked })
          }
          labelProps={{ labelText: 'Tear down dependent components' }}
        />
      ) : null}
      {step.operation === 'reprovision-sandbox' ? (
        <CheckboxInput
          checked={!!step.skipComponentDeploys}
          onChange={(event) =>
            onChange({ skipComponentDeploys: event.target.checked })
          }
          labelProps={{ labelText: 'Skip component deploys' }}
        />
      ) : null}
      {step.operation === 'configured-action' ? (
        <Select
          labelProps={{ labelText: 'Action' }}
          options={actions.map((action) => ({
            value: action.name,
            label: action.name,
          }))}
          value={step.actionName ?? ''}
          placeholder="Select an action"
          searchable
          onChange={(value) => onChange({ actionName: value })}
        />
      ) : null}
      {step.operation === 'command' ? (
        <>
          <Input
            id={`step-command-${step.key}`}
            labelProps={{ labelText: 'Command' }}
            value={step.command ?? ''}
            onChange={(event) => onChange({ command: event.target.value })}
          />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <Input
              id={`step-timeout-${step.key}`}
              labelProps={{ labelText: 'Timeout' }}
              placeholder="10m"
              value={step.timeout ?? ''}
              onChange={(event) => onChange({ timeout: event.target.value })}
            />
            <Input
              id={`step-role-${step.key}`}
              labelProps={{ labelText: 'Role' }}
              value={step.role ?? ''}
              onChange={(event) => onChange({ role: event.target.value })}
            />
          </div>
        </>
      ) : null}
    </div>
  )
}

import { useState } from 'react'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Input } from '../../atoms/Input'
import { Text } from '../../atoms/Text'
import { Field } from '../../molecules/Field'
import type { IWizardDescriptor } from '../../../utils/wizard'
import { Wizard } from './Wizard'

export default {
  title: 'lite/templates/Wizard',
}

interface IDemoState {
  name: string
  connected: boolean
  provisioned: boolean
}

const NameStep = ({
  state,
  readOnly,
  onNameChange,
}: {
  state: IDemoState
  readOnly: boolean
  onNameChange: (value: string) => void
}) => (
  <div className="flex flex-col gap-3">
    <Text as="h2" variant="heading">
      Name the app
    </Text>
    <Field label="Name">
      <Input
        value={state.name}
        onChange={(event) => onNameChange(event.target.value)}
        disabled={readOnly}
        aria-label="Name"
      />
    </Field>
  </div>
)

const ConnectStep = ({
  readOnly,
  onConnect,
}: {
  readOnly: boolean
  onConnect: () => void
}) => (
  <div className="flex flex-col gap-3">
    <Text as="h2" variant="heading">
      Connect the branch
    </Text>
    <Text color="secondary">
      {readOnly
        ? 'The branch is connected.'
        : 'Connect a branch to continue.'}
    </Text>
    {readOnly ? null : (
      <button type="button" onClick={onConnect}>
        Connect branch
      </button>
    )}
  </div>
)

const ProvisionStep = ({
  state,
  readOnly,
  onProvision,
}: {
  state: IDemoState
  readOnly: boolean
  onProvision: () => void
}) => (
  <div className="flex flex-col gap-3">
    <Text as="h2" variant="heading">
      Watch provision
    </Text>
    <Text color="secondary">
      {state.provisioned
        ? 'Provision finished. Continue when you are ready.'
        : 'Waiting for provision to finish.'}
    </Text>
    {readOnly || state.provisioned ? null : (
      <button type="button" onClick={onProvision}>
        Finish provision
      </button>
    )}
  </div>
)

const demoDescriptor = (
  onNameChange: (value: string) => void,
  onConnect: () => void,
  onProvision: () => void
): IWizardDescriptor<IDemoState> => ({
  steps: [
    {
      id: 'name',
      label: 'Name',
      complete: (state) => state.name.trim().length > 0,
      render: ({ state, readOnly }) => (
        <NameStep state={state} readOnly={readOnly} onNameChange={onNameChange} />
      ),
    },
    {
      id: 'connect',
      label: 'Connect',
      complete: (state) => state.connected,
      render: ({ readOnly }) => (
        <ConnectStep readOnly={readOnly} onConnect={onConnect} />
      ),
    },
    {
      id: 'provision',
      label: 'Provision',
      complete: (state) => state.provisioned,
      render: ({ state, readOnly }) => (
        <ProvisionStep
          state={state}
          readOnly={readOnly}
          onProvision={onProvision}
        />
      ),
    },
  ],
})

const DemoWizard = ({ initial }: { initial: IDemoState }) => {
  const [state, setState] = useState(initial)
  const descriptor = demoDescriptor(
    (name) => setState((current) => ({ ...current, name })),
    () => setState((current) => ({ ...current, connected: true })),
    () => setState((current) => ({ ...current, provisioned: true }))
  )

  return (
    <Wizard
      descriptor={descriptor}
      state={state}
      exitAction={{
        children: 'Continue to app',
        variant: 'primary',
        onClick: () => {},
      }}
    />
  )
}

export const Overview = () => (
  <ComponentDocs
    name="Wizard"
    tier="template"
    summary="The setup template that derives the current step from server state and renders earlier steps read-only."
    use={[
      'Mount it from onboarding, app setup, and install setup once those flows supply a descriptor.',
    ]}
    avoid={[
      'Do not store the current step or a furthest-reached mark.',
      'Do not put a real app or install into the template — steps come from the descriptor.',
      'Do not auto-advance or redirect when a watcher step succeeds; enable the exit action and wait.',
    ]}
    rules={[
      'The current step is the first whose complete callback returns false.',
      'Earlier steps are done and render read-only; later steps are listed and unreachable.',
      'When every step is complete the wizard is finished and the exit action is enabled.',
      'State going backwards moves the current step back.',
      'The stepper is internal to Wizard and shows every step plus the total count.',
    ]}
    props={[
      {
        name: 'descriptor',
        type: 'IWizardDescriptor<TState>',
        description: 'Ordered steps with complete predicates and render bodies.',
      },
      {
        name: 'state',
        type: 'TState',
        description: 'Server-backed state the steps resolve against.',
      },
      {
        name: 'exitAction',
        type: 'IButton',
        description:
          'Shown throughout and enabled only when every step is complete.',
      },
    ]}
  />
)

export const CurrentStep = () => (
  <DemoWizard initial={{ name: 'Payments', connected: false, provisioned: false }} />
)

export const ReadOnlyEarlierSteps = () => (
  <DemoWizard initial={{ name: 'Payments', connected: true, provisioned: false }} />
)

export const Finished = () => (
  <DemoWizard
    initial={{ name: 'Payments', connected: true, provisioned: true }}
  />
)

export const Regression = () => (
  <DemoWizard initial={{ name: '', connected: true, provisioned: true }} />
)

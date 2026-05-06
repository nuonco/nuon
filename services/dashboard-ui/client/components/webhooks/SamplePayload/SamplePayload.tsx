import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'

export const SAMPLE_PAYLOAD = `{
  "specversion": "1.0",
  "id": "3487f2f5-ec33-45db-8200-6cc669c044e3",
  "type": "com.nuon.workflow_step.lifecycle.v1",
  "source": "//nuon.co/ctl-api",
  "time": "2026-04-28T04:42:52Z",
  "subject": "orgeovwzku7qimrb9lxhrowckg/workflow_step/inw9zk2y7c2vcwf3kf30tgkg7e/iwsmd2x4kf6q7w7e0vh3vfz9hg/succeeded",
  "datacontenttype": "application/json",
  "nuonorgid": "orgeovwzku7qimrb9lxhrowckg",
  "nuonkind": "workflow_step",
  "nuontransition": "succeeded",
  "interests": [
    "resource:components",
    "op:components.deploy",
    "event:lifecycle.succeeded",
    "outcome:completion"
  ],
  "data": {
    "kind": "workflow_step",
    "transition": "succeeded",
    "org_id": "orgeovwzku7qimrb9lxhrowckg",
    "workflow": {
      "id": "inw9zk2y7c2vcwf3kf30tgkg7e",
      "type": "manual_deploy",
      "owner_id": "ins98e2wpzdxwoey393edtqj45",
      "owner_type": "installs"
    },
    "step": {
      "id": "iwsmd2x4kf6q7w7e0vh3vfz9hg",
      "name": "deploy api (apply)",
      "idx": 7,
      "target_type": "install_deploys",
      "target_id": "ind5a4j3mvk1q7yc2c0r3z9z8q",
      "component_id": "cmp05dfarnb3wt8wvcxanqovby",
      "execution_type": "system"
    },
    "outcome": {
      "status": "succeeded",
      "duration_ms": 12345
    },
    "links": {
      "org": "https://app.nuon.co/orgeovwzku7qimrb9lxhrowckg",
      "install": "https://app.nuon.co/orgeovwzku7qimrb9lxhrowckg/installs/ins98e2wpzdxwoey393edtqj45",
      "workflow": "https://app.nuon.co/orgeovwzku7qimrb9lxhrowckg/installs/ins98e2wpzdxwoey393edtqj45/workflows/inw9zk2y7c2vcwf3kf30tgkg7e",
      "component": "https://app.nuon.co/orgeovwzku7qimrb9lxhrowckg/installs/ins98e2wpzdxwoey393edtqj45/components/cmp05dfarnb3wt8wvcxanqovby"
    }
  }
}`

export const SamplePayload = () => {
  return (
    <Expand
      id="webhook-sample-payload"
      className="rounded-md border"
      headerClassName="p-4"
      heading={
        <Text weight="strong">
          View sample payload
        </Text>
      }
    >
      <div className="flex flex-col gap-3 p-4 border-t">
        <div className="flex justify-end">
          <ClickToCopyButton textToCopy={SAMPLE_PAYLOAD} title="Copy payload" />
        </div>
        <CodeBlock language="json">{SAMPLE_PAYLOAD}</CodeBlock>
      </div>
    </Expand>
  )
}

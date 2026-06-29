export default {
  title: 'Installs/InputValue',
}

import { InputValue } from './InputValue'

export const Scalar = () => <InputValue name="domain" value="nuon.run" />

export const Empty = () => <InputValue name="domain" value="" />

export const Missing = () => <InputValue name="domain" value={null} />

export const HelmValues = () => (
  <InputValue
    name="nuon_component_override_v1_helm_values_77686f616d69"
    value={'replicaCount: 5\nresources:\n  requests:\n    cpu: "150m"\n    memory: 64Mi\n'}
  />
)

export const TFVars = () => (
  <InputValue
    name="nuon_component_override_v1_tf_vars_6365727469666963617465"
    value={'domain_name = "whoami.example.com"\n'}
  />
)

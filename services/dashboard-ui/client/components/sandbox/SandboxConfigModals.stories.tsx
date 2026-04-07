export default {
  title: 'Sandbox/SandboxConfigModals',
}

import { ModalStory } from '@/components/__stories__/helpers'
import {
  SandboxEnvironmentVariablesModal,
  SandboxVariablesFilesModal,
} from './SandboxConfigModals'

export const EnvironmentVariables = () => (
  <ModalStory label="Open env vars modal">
    <SandboxEnvironmentVariablesModal
      envVars={{
        DATABASE_URL: 'postgres://localhost:5432/mydb',
        API_KEY: 'secret-key-123',
        NODE_ENV: 'production',
      }}
    />
  </ModalStory>
)

export const VariablesFiles = () => (
  <ModalStory label="Open variables files modal">
    <SandboxVariablesFilesModal
      variablesFiles={[
        'region = "us-west-2"\ninstance_type = "t3.medium"',
        'vpc_id = "vpc-12345"\nsubnet_ids = ["subnet-1", "subnet-2"]',
      ]}
    />
  </ModalStory>
)

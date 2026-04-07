export default {
  title: 'Components/Configs/KubernetesConfig',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { KubernetesManifestModal } from './KubernetesConfig'

export const ManifestModal = () => (
  <ModalStory>
    <KubernetesManifestModal
      manifest={`apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: my-registry/my-app:1.2.3
        ports:
        - containerPort: 8080`}
    />
  </ModalStory>
)

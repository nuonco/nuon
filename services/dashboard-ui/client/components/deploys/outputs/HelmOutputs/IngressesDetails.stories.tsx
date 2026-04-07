export default {
  title: 'Deploys/HelmOutputs/IngressesDetails',
}

import { IngressesDetails } from './IngressesDetails'

const mockIngresses = {
  default: {
    'my-app-ingress': {
      metadata: {
        creationTimestamp: '2024-01-15T10:30:00Z',
        generation: 1,
        resourceVersion: '33333',
        uid: 'ing-abc-123',
        annotations: {
          'external-dns.alpha.kubernetes.io/hostname': 'app.example.com',
          'alb.ingress.kubernetes.io/scheme': 'internet-facing',
          'alb.ingress.kubernetes.io/target-type': 'ip',
          'alb.ingress.kubernetes.io/certificate-arn': 'arn:aws:acm:us-east-1:123456789:certificate/abc-def',
          'alb.ingress.kubernetes.io/listen-ports': '[{"HTTP": 80}, {"HTTPS": 443}]',
          'alb.ingress.kubernetes.io/healthcheck-path': '/health',
          'alb.ingress.kubernetes.io/healthcheck-timeout-seconds': '10',
          'alb.ingress.kubernetes.io/healthy-threshold-count': '2',
          'alb.ingress.kubernetes.io/unhealthy-threshold-count': '3',
        },
      },
      status: {
        loadBalancer: {
          ingress: [{ hostname: 'k8s-abc123.us-east-1.elb.amazonaws.com' }],
        },
      },
    },
  },
}

export const Default = () => <IngressesDetails ingresses={mockIngresses} />

export const Empty = () => <IngressesDetails ingresses={{}} />

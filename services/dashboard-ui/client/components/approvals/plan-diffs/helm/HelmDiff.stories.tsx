export default {
  title: 'Approvals/PlanDiffs/HelmDiff',
}

import { HelmDiff } from './HelmDiff'

const mockPlan = {
  plan: `default, my-app, Deployment (apps/v1) to be changed
default, my-app-svc, Service (v1) to be added
default, my-app-cache, Deployment (apps/v1) to be destroyed
Plan: 1 to add, 1 to change, 1 to destroy`,
  op: 'upgrade',
  helm_content_diff: [
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'my-app',
      namespace: 'default',
      before: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: app
        image: my-app:1.2.0
        resources:
          limits:
            memory: 256Mi`,
      after: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: my-app:1.3.0
        resources:
          limits:
            memory: 512Mi`,
    },
    {
      api: 'v1',
      kind: 'Service',
      name: 'my-app-svc',
      namespace: 'default',
      before: '',
      after: `apiVersion: v1
kind: Service
metadata:
  name: my-app-svc
  namespace: default
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080`,
    },
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'my-app-cache',
      namespace: 'default',
      before: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-cache
  namespace: default
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: redis
        image: redis:6.2`,
      after: '',
    },
  ],
} as any

export const Default = () => <HelmDiff plan={mockPlan} />

const nginxIngressUpgradePlan = {
  plan: `ingress-nginx, nginx-ingress-controller, Deployment (apps/v1) to be changed
ingress-nginx, nginx-ingress-controller, ConfigMap (v1) to be changed
ingress-nginx, nginx-ingress-controller, Service (v1) to be changed
Plan: 0 to add, 3 to change, 0 to destroy`,
  op: 'upgrade',
  helm_content_diff: [
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'nginx-ingress-controller',
      namespace: 'ingress-nginx',
      before: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/version: 1.9.4
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: ingress-nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ingress-nginx
    spec:
      serviceAccountName: ingress-nginx
      containers:
      - name: controller
        image: registry.k8s.io/ingress-nginx/controller:v1.9.4
        args:
        - /nginx-ingress-controller
        - --configmap=ingress-nginx/nginx-ingress-controller
        - --publish-service=ingress-nginx/nginx-ingress-controller
        ports:
        - name: http
          containerPort: 80
        - name: https
          containerPort: 443
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10254
          initialDelaySeconds: 10
          periodSeconds: 10`,
      after: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/version: 1.10.1
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: ingress-nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ingress-nginx
    spec:
      serviceAccountName: ingress-nginx
      containers:
      - name: controller
        image: registry.k8s.io/ingress-nginx/controller:v1.10.1
        args:
        - /nginx-ingress-controller
        - --configmap=ingress-nginx/nginx-ingress-controller
        - --publish-service=ingress-nginx/nginx-ingress-controller
        - --enable-metrics=true
        ports:
        - name: http
          containerPort: 80
        - name: https
          containerPort: 443
        - name: metrics
          containerPort: 10254
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10254
          initialDelaySeconds: 10
          periodSeconds: 10`,
    },
    {
      api: 'v1',
      kind: 'ConfigMap',
      name: 'nginx-ingress-controller',
      namespace: 'ingress-nginx',
      before: `apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
data:
  proxy-body-size: 8m
  proxy-connect-timeout: "10"
  proxy-read-timeout: "60"
  use-forwarded-headers: "true"`,
      after: `apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
data:
  proxy-body-size: 16m
  proxy-connect-timeout: "10"
  proxy-read-timeout: "120"
  proxy-send-timeout: "120"
  use-forwarded-headers: "true"
  enable-real-ip: "true"
  max-worker-connections: "65536"`,
    },
    {
      api: 'v1',
      kind: 'Service',
      name: 'nginx-ingress-controller',
      namespace: 'ingress-nginx',
      before: `apiVersion: v1
kind: Service
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: LoadBalancer
  ports:
  - name: http
    port: 80
    targetPort: http
  - name: https
    port: 443
    targetPort: https
  selector:
    app.kubernetes.io/name: ingress-nginx`,
      after: `apiVersion: v1
kind: Service
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
    service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: "true"
    service.beta.kubernetes.io/aws-load-balancer-proxy-protocol: "*"
spec:
  type: LoadBalancer
  ports:
  - name: http
    port: 80
    targetPort: http
  - name: https
    port: 443
    targetPort: https
  selector:
    app.kubernetes.io/name: ingress-nginx`,
    },
  ],
} as any

export const NginxIngressUpgrade = () => (
  <HelmDiff plan={nginxIngressUpgradePlan} />
)

const certManagerInstallPlan = {
  plan: `cert-manager, cert-manager, Deployment (apps/v1) to be added
cert-manager, cert-manager, ServiceAccount (v1) to be added
cert-manager, cert-manager, ClusterRole (rbac.authorization.k8s.io/v1) to be added
cert-manager, cert-manager, ClusterRoleBinding (rbac.authorization.k8s.io/v1) to be added
cert-manager, cert-manager-webhook, Deployment (apps/v1) to be added
cert-manager, cert-manager-webhook, Service (v1) to be added
Plan: 6 to add, 0 to change, 0 to destroy`,
  op: 'install',
  helm_content_diff: [
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'cert-manager',
      namespace: 'cert-manager',
      before: '',
      after: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager
  namespace: cert-manager
  labels:
    app: cert-manager
    app.kubernetes.io/name: cert-manager
    app.kubernetes.io/version: v1.14.4
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: cert-manager
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cert-manager
    spec:
      serviceAccountName: cert-manager
      containers:
      - name: cert-manager-controller
        image: quay.io/jetstack/cert-manager-controller:v1.14.4
        args:
        - --v=2
        - --cluster-resource-namespace=$(POD_NAMESPACE)
        - --leader-election-namespace=kube-system
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        ports:
        - containerPort: 9402
          name: http-metrics
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 256Mi`,
    },
    {
      api: 'v1',
      kind: 'ServiceAccount',
      name: 'cert-manager',
      namespace: 'cert-manager',
      before: '',
      after: `apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager
  namespace: cert-manager
  labels:
    app: cert-manager
    app.kubernetes.io/name: cert-manager
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/cert-manager`,
    },
    {
      api: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRole',
      name: 'cert-manager',
      namespace: 'cert-manager',
      before: '',
      after: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cert-manager
  labels:
    app: cert-manager
rules:
- apiGroups: ["cert-manager.io"]
  resources: ["certificates", "certificaterequests", "issuers", "clusterissuers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["secrets", "events", "configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses", "ingresses/finalizers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]`,
    },
    {
      api: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRoleBinding',
      name: 'cert-manager',
      namespace: 'cert-manager',
      before: '',
      after: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cert-manager
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager
subjects:
- kind: ServiceAccount
  name: cert-manager
  namespace: cert-manager`,
    },
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'cert-manager-webhook',
      namespace: 'cert-manager',
      before: '',
      after: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager-webhook
  namespace: cert-manager
  labels:
    app: webhook
    app.kubernetes.io/name: webhook
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: webhook
  template:
    metadata:
      labels:
        app.kubernetes.io/name: webhook
    spec:
      serviceAccountName: cert-manager
      containers:
      - name: cert-manager-webhook
        image: quay.io/jetstack/cert-manager-webhook:v1.14.4
        args:
        - --v=2
        - --secure-port=10250
        - --dynamic-serving-ca-secret-namespace=cert-manager
        - --dynamic-serving-ca-secret-name=cert-manager-webhook-ca
        ports:
        - name: https
          containerPort: 10250
        - name: healthcheck
          containerPort: 6080
        resources:
          requests:
            cpu: 25m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 128Mi`,
    },
    {
      api: 'v1',
      kind: 'Service',
      name: 'cert-manager-webhook',
      namespace: 'cert-manager',
      before: '',
      after: `apiVersion: v1
kind: Service
metadata:
  name: cert-manager-webhook
  namespace: cert-manager
  labels:
    app: webhook
spec:
  type: ClusterIP
  ports:
  - name: https
    port: 443
    targetPort: https
  selector:
    app.kubernetes.io/name: webhook`,
    },
  ],
} as any

export const CertManagerInstall = () => (
  <HelmDiff plan={certManagerInstallPlan} />
)

const postgresOperatorUpgradePlan = {
  plan: `cnpg-system, cloudnative-pg, Deployment (apps/v1) to be changed
cnpg-system, cloudnative-pg, ConfigMap (v1) to be changed
cnpg-system, cloudnative-pg-backup, CronJob (batch/v1) to be added
Plan: 1 to add, 2 to change, 0 to destroy`,
  op: 'upgrade',
  helm_content_diff: [
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'cloudnative-pg',
      namespace: 'cnpg-system',
      before: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudnative-pg
  namespace: cnpg-system
  labels:
    app.kubernetes.io/name: cloudnative-pg
    app.kubernetes.io/version: 1.22.1
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: cloudnative-pg
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cloudnative-pg
    spec:
      serviceAccountName: cloudnative-pg
      containers:
      - name: manager
        image: ghcr.io/cloudnative-pg/cloudnative-pg:1.22.1
        command:
        - /manager
        args:
        - controller
        - --leader-elect
        env:
        - name: OPERATOR_IMAGE_NAME
          value: ghcr.io/cloudnative-pg/cloudnative-pg:1.22.1
        - name: OPERATOR_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        ports:
        - containerPort: 8080
          name: metrics
        - containerPort: 9443
          name: webhook-server
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi`,
      after: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudnative-pg
  namespace: cnpg-system
  labels:
    app.kubernetes.io/name: cloudnative-pg
    app.kubernetes.io/version: 1.23.0
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: cloudnative-pg
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cloudnative-pg
    spec:
      serviceAccountName: cloudnative-pg
      containers:
      - name: manager
        image: ghcr.io/cloudnative-pg/cloudnative-pg:1.23.0
        command:
        - /manager
        args:
        - controller
        - --leader-elect
        - --webhook-port=9443
        env:
        - name: OPERATOR_IMAGE_NAME
          value: ghcr.io/cloudnative-pg/cloudnative-pg:1.23.0
        - name: OPERATOR_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: MONITORING_QUERIES_CONFIGMAP
          value: cnpg-default-monitoring
        ports:
        - containerPort: 8080
          name: metrics
        - containerPort: 9443
          name: webhook-server
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 768Mi`,
    },
    {
      api: 'v1',
      kind: 'ConfigMap',
      name: 'cloudnative-pg',
      namespace: 'cnpg-system',
      before: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudnative-pg
  namespace: cnpg-system
data:
  INHERITED_ANNOTATIONS: ""
  INHERITED_LABELS: ""
  CREATE_ANY_SERVICE: "false"`,
      after: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudnative-pg
  namespace: cnpg-system
data:
  INHERITED_ANNOTATIONS: "monitoring.nuon.co/enabled"
  INHERITED_LABELS: "team,environment"
  CREATE_ANY_SERVICE: "true"
  ENABLE_INSTANCE_MANAGER_INPLACE_UPDATES: "true"`,
    },
    {
      api: 'batch/v1',
      kind: 'CronJob',
      name: 'cloudnative-pg-backup',
      namespace: 'cnpg-system',
      before: '',
      after: `apiVersion: batch/v1
kind: CronJob
metadata:
  name: cloudnative-pg-backup
  namespace: cnpg-system
  labels:
    app.kubernetes.io/name: cloudnative-pg
spec:
  schedule: "0 2 * * *"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: cloudnative-pg
          containers:
          - name: backup
            image: ghcr.io/cloudnative-pg/cloudnative-pg:1.23.0
            command:
            - /backup-tool
            args:
            - --all-namespaces
            - --backup-target=s3://cnpg-backups/daily
            - --retention=30d
            env:
            - name: AWS_DEFAULT_REGION
              value: us-west-2
            resources:
              requests:
                cpu: 50m
                memory: 64Mi
              limits:
                cpu: 200m
                memory: 256Mi
          restartPolicy: OnFailure`,
    },
  ],
} as any

export const PostgresOperatorUpgrade = () => (
  <HelmDiff plan={postgresOperatorUpgradePlan} />
)

const prometheusStackChangePlan = {
  plan: `monitoring, kube-prometheus-stack, Deployment (apps/v1) to be changed
monitoring, kube-prometheus-stack, ConfigMap (v1) to be changed
monitoring, kube-prometheus-stack-grafana, ConfigMap (v1) to be changed
monitoring, kube-prometheus-stack, ServiceMonitor (monitoring.coreos.com/v1) to be changed
Plan: 0 to add, 4 to change, 0 to destroy`,
  op: 'upgrade',
  helm_content_diff: [
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'kube-prometheus-stack',
      namespace: 'monitoring',
      before: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-prometheus-stack
  namespace: monitoring
  labels:
    app: prometheus
    chart: kube-prometheus-stack-56.6.2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      serviceAccountName: kube-prometheus-stack
      containers:
      - name: prometheus
        image: quay.io/prometheus/prometheus:v2.50.1
        args:
        - --config.file=/etc/prometheus/prometheus.yml
        - --storage.tsdb.path=/prometheus
        - --storage.tsdb.retention.time=15d
        - --web.enable-lifecycle
        ports:
        - containerPort: 9090
          name: web
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        - name: storage
          mountPath: /prometheus
        resources:
          requests:
            cpu: 200m
            memory: 512Mi
          limits:
            cpu: "1"
            memory: 2Gi
      volumes:
      - name: config
        configMap:
          name: kube-prometheus-stack
      - name: storage
        emptyDir: {}`,
      after: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-prometheus-stack
  namespace: monitoring
  labels:
    app: prometheus
    chart: kube-prometheus-stack-58.2.0
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      serviceAccountName: kube-prometheus-stack
      containers:
      - name: prometheus
        image: quay.io/prometheus/prometheus:v2.51.2
        args:
        - --config.file=/etc/prometheus/prometheus.yml
        - --storage.tsdb.path=/prometheus
        - --storage.tsdb.retention.time=30d
        - --web.enable-lifecycle
        - --web.enable-remote-write-receiver
        ports:
        - containerPort: 9090
          name: web
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        - name: storage
          mountPath: /prometheus
        resources:
          requests:
            cpu: 200m
            memory: 512Mi
          limits:
            cpu: "2"
            memory: 4Gi
      - name: thanos-sidecar
        image: quay.io/thanos/thanos:v0.34.1
        args:
        - sidecar
        - --tsdb.path=/prometheus
        - --prometheus.url=http://localhost:9090
        - --objstore.config-file=/etc/thanos/objstore.yml
        ports:
        - containerPort: 10902
          name: grpc
        - containerPort: 10901
          name: http
        volumeMounts:
        - name: storage
          mountPath: /prometheus
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 256Mi
      volumes:
      - name: config
        configMap:
          name: kube-prometheus-stack
      - name: storage
        emptyDir: {}`,
    },
    {
      api: 'v1',
      kind: 'ConfigMap',
      name: 'kube-prometheus-stack',
      namespace: 'monitoring',
      before: `apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-prometheus-stack
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 30s
      evaluation_interval: 30s
    scrape_configs:
    - job_name: kubernetes-pods
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
    - job_name: kubernetes-nodes
      kubernetes_sd_configs:
      - role: node
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt`,
      after: `apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-prometheus-stack
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    scrape_configs:
    - job_name: kubernetes-pods
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
    - job_name: kubernetes-nodes
      kubernetes_sd_configs:
      - role: node
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    - job_name: cnpg-clusters
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: [databases]
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_cnpg_io_cluster]
        action: keep
        regex: .+
      scrape_interval: 10s`,
    },
    {
      api: 'v1',
      kind: 'ConfigMap',
      name: 'kube-prometheus-stack-grafana',
      namespace: 'monitoring',
      before: `apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-prometheus-stack-grafana
  namespace: monitoring
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
    - name: Prometheus
      type: prometheus
      url: http://kube-prometheus-stack:9090
      isDefault: true
      access: proxy`,
      after: `apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-prometheus-stack-grafana
  namespace: monitoring
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
    - name: Prometheus
      type: prometheus
      url: http://kube-prometheus-stack:9090
      isDefault: true
      access: proxy
    - name: Thanos
      type: prometheus
      url: http://thanos-query:10902
      access: proxy
    - name: ClickHouse
      type: grafana-clickhouse-datasource
      url: http://clickhouse:8123
      access: proxy
      jsonData:
        defaultDatabase: otel`,
    },
    {
      api: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      name: 'kube-prometheus-stack',
      namespace: 'monitoring',
      before: `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-prometheus-stack
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  selector:
    matchLabels:
      app: prometheus
  endpoints:
  - port: web
    interval: 30s`,
      after: `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-prometheus-stack
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  selector:
    matchLabels:
      app: prometheus
  endpoints:
  - port: web
    interval: 15s
    metricRelabelings:
    - sourceLabels: [__name__]
      regex: "go_.*"
      action: drop
  - port: grpc
    interval: 30s`,
    },
  ],
} as any

export const PrometheusStackChange = () => (
  <HelmDiff plan={prometheusStackChangePlan} />
)

const redisClusterRollbackPlan = {
  plan: `redis, redis-cluster, Deployment (apps/v1) to be changed
redis, redis-cluster-sentinel, Service (v1) to be destroyed
redis, redis-cluster, ConfigMap (v1) to be changed
Plan: 0 to add, 2 to change, 1 to destroy`,
  op: 'rollback',
  helm_content_diff: [
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'redis-cluster',
      namespace: 'redis',
      before: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-cluster
  namespace: redis
  labels:
    app.kubernetes.io/name: redis
    app.kubernetes.io/version: 7.2.4
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: redis
  template:
    metadata:
      labels:
        app.kubernetes.io/name: redis
    spec:
      containers:
      - name: redis
        image: redis:7.2.4-alpine
        command: ["redis-server"]
        args: ["/etc/redis/redis.conf"]
        ports:
        - containerPort: 6379
          name: redis
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-credentials
              key: password
        - name: REDIS_MAXMEMORY_POLICY
          value: allkeys-lru
        - name: REDIS_SENTINEL_ENABLED
          value: "true"
        volumeMounts:
        - name: config
          mountPath: /etc/redis
        - name: data
          mountPath: /data
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 1Gi
      volumes:
      - name: config
        configMap:
          name: redis-cluster
      - name: data
        emptyDir: {}`,
      after: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-cluster
  namespace: redis
  labels:
    app.kubernetes.io/name: redis
    app.kubernetes.io/version: 7.2.3
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: redis
  template:
    metadata:
      labels:
        app.kubernetes.io/name: redis
    spec:
      containers:
      - name: redis
        image: redis:7.2.3-alpine
        command: ["redis-server"]
        args: ["/etc/redis/redis.conf"]
        ports:
        - containerPort: 6379
          name: redis
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-credentials
              key: password
        - name: REDIS_MAXMEMORY_POLICY
          value: allkeys-lru
        volumeMounts:
        - name: config
          mountPath: /etc/redis
        - name: data
          mountPath: /data
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 1Gi
      volumes:
      - name: config
        configMap:
          name: redis-cluster
      - name: data
        emptyDir: {}`,
    },
    {
      api: 'v1',
      kind: 'Service',
      name: 'redis-cluster-sentinel',
      namespace: 'redis',
      before: `apiVersion: v1
kind: Service
metadata:
  name: redis-cluster-sentinel
  namespace: redis
  labels:
    app.kubernetes.io/name: redis
    app.kubernetes.io/component: sentinel
spec:
  type: ClusterIP
  ports:
  - name: sentinel
    port: 26379
    targetPort: 26379
  - name: redis
    port: 6379
    targetPort: 6379
  selector:
    app.kubernetes.io/name: redis
    app.kubernetes.io/component: sentinel`,
      after: '',
    },
    {
      api: 'v1',
      kind: 'ConfigMap',
      name: 'redis-cluster',
      namespace: 'redis',
      before: `apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-cluster
  namespace: redis
data:
  redis.conf: |
    bind 0.0.0.0
    port 6379
    maxmemory 768mb
    maxmemory-policy allkeys-lru
    appendonly yes
    appendfsync everysec
    save 900 1
    save 300 10
    sentinel monitor mymaster redis-cluster-0.redis-cluster 6379 2
    sentinel down-after-milliseconds mymaster 5000
    sentinel failover-timeout mymaster 10000`,
      after: `apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-cluster
  namespace: redis
data:
  redis.conf: |
    bind 0.0.0.0
    port 6379
    maxmemory 768mb
    maxmemory-policy allkeys-lru
    appendonly yes
    appendfsync everysec
    save 900 1
    save 300 10`,
    },
  ],
} as any

export const RedisClusterRollback = () => (
  <HelmDiff plan={redisClusterRollbackPlan} />
)

const vmagentSingleRemovalPlan = {
  plan: `\u001b[0;33mobservability, vmagent, Deployment (apps) to be changed.\u001b[0m\nPlan: 0 to add, 1 to change, 0 to destroy`,
  op: 'upgrade',
  helm_content_diff: [
    {
      _version: '2',
      name: 'vmagent',
      namespace: 'observability',
      kind: 'Deployment',
      api: 'apps',
      type: 0,
      entries: [
        { type: 0, payload: 'apiVersion: apps/v1' },
        { type: 0, payload: 'kind: Deployment' },
        { type: 0, payload: 'metadata:' },
        { type: 0, payload: '  name: vmagent' },
        { type: 0, payload: '  namespace: observability' },
        { type: 0, payload: 'spec:' },
        { type: 0, payload: '  replicas: 1' },
        { type: 0, payload: '  template:' },
        { type: 0, payload: '    spec:' },
        { type: 0, payload: '      containers:' },
        { type: 0, payload: '      - args:' },
        { type: 0, payload: '        - --envflag.enable' },
        { type: 0, payload: '        - --httpListenAddr=:8429' },
        { type: 0, payload: '        - --loggerFormat=json' },
        { type: 0, payload: '        - --remoteWrite.bearerToken=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
        { type: 0, payload: '        - --remoteWrite.tmpDataPath=/tmpData' },
        { type: 0, payload: '        - --remoteWrite.url=https://metrics.example.com/api/v1/write' },
        { type: 1, payload: '        - --remoteWrite.bearerToken=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
        { type: 0, payload: '        env:' },
        { type: 0, payload: '        - name: POD_NAME' },
        { type: 0, payload: '          valueFrom:' },
        { type: 0, payload: '            fieldRef:' },
        { type: 0, payload: '              apiVersion: v1' },
        { type: 0, payload: '              fieldPath: metadata.name' },
        { type: 0, payload: '        image: victoriametrics/vmagent:v1.132.0' },
        { type: 0, payload: '        imagePullPolicy: IfNotPresent' },
        { type: 0, payload: '        resources:' },
        { type: 0, payload: '          limits:' },
        { type: 0, payload: '            cpu: 500m' },
        { type: 0, payload: '            memory: 512Mi' },
        { type: 0, payload: '          requests:' },
        { type: 0, payload: '            cpu: 200m' },
        { type: 0, payload: '            memory: 256Mi' },
      ],
    },
  ],
} as any

export const VmagentSingleRemoval = () => (
  <HelmDiff plan={vmagentSingleRemovalPlan} />
)

import type { FC } from 'react'
import {
  CaretLeft,
  CaretUp,
  CaretUpDown,
  CaretRight,
  DotsThreeVertical,
  Stack,
} from '@phosphor-icons/react/dist/ssr'
import {
  Button,
  Editor,
  DiffEditor,
  Menu,
  Text,
  Link,
  Tooltip,
  PageHeader,
  Dropdown,
} from '@/stratus/components'
import type { IPageProps } from '@/types'
import { nueQueryData } from '@/utils'

const DropdownMenu = () => (
  <Menu className="min-w-52">
    <Text variant="label" theme="muted">
      Section label
    </Text>
    <Button>
      Option <Stack />
    </Button>
    <Button>Option</Button>
    <Dropdown
      id="sub-dropdown"
      buttonText="Sub menu"
      position="beside"
      alignment="right"
      icon={<CaretRight />}
    >
      <Menu className="min-w-52">
        <Text variant="label" theme="muted">
          Section label
        </Text>
        <Button>Option</Button>
        <Button>Option</Button>
        <Dropdown
          id="sub-sub-dropdown"
          buttonText="Sub menu"
          position="beside"
          alignment="right"
          icon={<CaretRight />}
        >
          <Menu className="min-w-52">
            <Text variant="label" theme="muted">
              Section label
            </Text>
            <Button>Option</Button>
            <Button>Option</Button>
            <Dropdown
              id="sub-sub-sub-dropdown"
              buttonText="Sub menu"
              position="beside"
              alignment="right"
              icon={<CaretRight />}
            >
              <Menu className="min-w-52">
                <Text variant="label" theme="muted">
                  Section label
                </Text>
                <Button>Option</Button>
                <Button>Option</Button>
                <Button>Option</Button>
                <hr />
                <Text variant="label" theme="muted">
                  Section label
                </Text>
                <Button>Option</Button>
              </Menu>
            </Dropdown>
            <hr />
            <Text variant="label" theme="muted">
              Section label
            </Text>
            <Button>Option</Button>
          </Menu>
        </Dropdown>
        <hr />
        <Text variant="label" theme="muted">
          Section label
        </Text>
        <Button>Option</Button>
      </Menu>
    </Dropdown>

    <hr />

    <Text variant="label" theme="muted">
      Section label
    </Text>
    <Button>Option</Button>
    <Button>Option</Button>
    <Button>Option</Button>
    <hr />
    <Text variant="label" theme="muted">
      Section label
    </Text>
    <Link href="#">Option</Link>
  </Menu>
)

const StratusDasboard: FC<IPageProps<'org-id'>> = async ({ params }) => {
  const orgId = params?.['org-id']
  const { data, error } = await nueQueryData<Record<string, any>>({
    orgId,
    path: `orgs/current`,
  })

  return (
    <div className="flex flex-col gap-4 p-4 overflow-auto">
      <PageHeader>
        <Text variant="h1" weight="stronger">
          Page header
        </Text>
      </PageHeader>

      <Text variant="h1" weight="stronger">
        Editor
      </Text>
      <div className="flex flex-col gap-4">
        <DiffEditor
          diff={`
# Source: myapp/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  labels:
    app: myapp
    chart: myapp-1.2.0
-   version: 1.1.0
+   version: 1.2.0
spec:
- replicas: 2
+ replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
-       image: mycompany/myapp:v1.1.0
+       image: mycompany/myapp:v1.2.0
        ports:
        - containerPort: 8080
        resources:
          limits:
-           cpu: 500m
+           cpu: 1000m
-           memory: 512Mi
+           memory: 1Gi
          requests:
-           cpu: 200m
+           cpu: 500m
-           memory: 256Mi
+           memory: 512Mi
-       livenessProbe:
-         httpGet:
-           path: /healthz
-           port: 8080
-         initialDelaySeconds: 30
+       livenessProbe:
+         httpGet:
+           path: /health/live
+           port: 8080
+         initialDelaySeconds: 20
+       readinessProbe:
+         httpGet:
+           path: /health/ready
+           port: 8080
+         initialDelaySeconds: 10

# Source: myapp/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp
  labels:
    app: myapp
spec:
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
-   name: http
+   name: http-web
+   
+ ---
+ # Source: myapp/templates/configmap.yaml
+ apiVersion: v1
+ kind: ConfigMap
+ metadata:
+   name: myapp-config
+ data:
+   application.properties: |
+     app.environment=production
+     app.logging.level=INFO
+     app.feature.newFeature=true

# Source: myapp/templates/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  annotations:
-   nginx.ingress.kubernetes.io/rewrite-target: /
+   nginx.ingress.kubernetes.io/rewrite-target: /$1
+   cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
+ tls:
+ - hosts:
+   - myapp.example.com
+   secretName: myapp-tls
  rules:
- - host: app.example.com
+ - host: myapp.example.com
    http:
      paths:
-     - path: /
+     - path: /(.*)
        pathType: Prefix
        backend:
          service:
            name: myapp
            port:
              number: 80
        `}
        />
      </div>

      <Text variant="h1" weight="stronger">
        Menu
      </Text>
      <div className="flex gap-4">
        <Menu className="min-w-52 border rounded-lg">
          <Text variant="label" theme="muted">
            Settings
          </Text>
          <Button>
            <Stack /> Option
          </Button>
          <Button>Option</Button>
          <Button>Option</Button>

          <hr />

          <Text variant="label" theme="muted">
            Controls
          </Text>
          <Button>Option</Button>
          <Button>Option</Button>
          <hr />
          <Text variant="label" theme="muted">
            Remove
          </Text>
          <Link href="#">Option</Link>
        </Menu>
      </div>

      <Text variant="h1" weight="stronger">
        Dropdown
      </Text>
      <div className="flex gap-4">
        <Dropdown id="dropdown" buttonText="Dropdown below left">
          <DropdownMenu />
        </Dropdown>

        <Dropdown
          alignment="right"
          id="dropdown"
          buttonText="Dropdown below right"
        >
          <DropdownMenu />
        </Dropdown>

        <Dropdown
          id="dropdown"
          buttonText="Dropdown ablove left"
          position="above"
          icon={<CaretUp />}
        >
          <DropdownMenu />
        </Dropdown>

        <Dropdown
          id="dropdown"
          buttonText="Dropdown above right"
          alignment="right"
          position="above"
          icon={<CaretUp />}
        >
          <DropdownMenu />
        </Dropdown>
      </div>

      <div className="flex gap-4">
        <Dropdown
          alignment="right"
          id="dropdown"
          buttonText="Dropdown beside right"
          position="beside"
          icon={<CaretRight />}
        >
          <DropdownMenu />
        </Dropdown>

        <Dropdown
          id="dropdown"
          buttonText="Dropdown beside left"
          position="beside"
          icon={<CaretLeft />}
          iconAlignment="left"
        >
          <DropdownMenu />
        </Dropdown>

        <Dropdown
          id="dropdown"
          buttonText="Dropdown overlay"
          position="overlay"
          alignment="overlay"
          icon={<CaretUpDown />}
        >
          <DropdownMenu />
        </Dropdown>

        <Dropdown
          id="dropdown"
          buttonClassName="!p-2"
          buttonText={<DotsThreeVertical />}
          hideIcon
        >
          <DropdownMenu />
        </Dropdown>
      </div>

      <Text variant="h1" weight="stronger">
        Links
      </Text>
      <div className="flex flex-col gap-4">
        <Link href="#">default</Link>
        <Link href="#" variant="ghost">
          ghost
        </Link>
        <Link href="#" variant="nav" isActive>
          nav
        </Link>
        <Link href="#" variant="breadcrumb">
          breadcrumb
        </Link>
      </div>

      <Text variant="h1" weight="stronger">
        Tooltip
      </Text>
      <div className="flex gap-4">
        <div className="flex gap-4 items-start">
          <Tooltip
            tipContent={<Text variant="subtext">Something</Text>}
            position="top"
          >
            <Text>Top tooltip</Text>
          </Tooltip>

          <Tooltip
            tipContent={<Text variant="subtext">Something</Text>}
            position="bottom"
          >
            <Text>Bottom tooltip</Text>
          </Tooltip>

          <Tooltip
            tipContent={<Text variant="subtext">Something</Text>}
            position="left"
          >
            <Text>Left tooltip</Text>
          </Tooltip>

          <Tooltip
            tipContent={<Text variant="subtext">Something</Text>}
            position="right"
          >
            <Text>Right tooltip</Text>
          </Tooltip>

          <Text className="max-w-80 text-balance">
            Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
            eiusmod tempor incididunt ut
            <Tooltip
              tipContent={
                <Text className="w-max" variant="subtext">
                  Tip with icon
                </Text>
              }
              showIcon
            >
              labore
            </Tooltip>
            et dolore magna aliqua.
          </Text>
        </div>

        <div className="flex gap-4 items-start">
          <Tooltip
            tipContent={
              <div className="flex flex-col w-80">
                <Text variant="body" weight="stronger">
                  Complex title
                </Text>

                <Text className="text-pretty" variant="subtext">
                  Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed
                  do eiusmod tempor incididunt ut labore et dolore magna aliqua.
                </Text>
              </div>
            }
            position="bottom"
          >
            <Text>Complex tooltip</Text>
          </Tooltip>
        </div>
      </div>

      <Text variant="h1" weight="stronger">
        Buttons
      </Text>
      <div className="flex gap-4">
        <div className="flex gap-4">
          <Button variant="primary">Primary</Button>
          <Button>Secondary</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="danger">Danger</Button>
        </div>
      </div>

      <Text variant="h1" weight="stronger">
        Typography
      </Text>
      <div className="flex gap-4">
        <div className="flex flex-col">
          <Text variant="h1">
            H1: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="h2">
            h2: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="h3">
            h3: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="base">
            base: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="body">
            body: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="subtext">
            subtext: The quick brown fox jumps over the lazy dog.
          </Text>
          <Text variant="label">
            label: The quick brown fox jumps over the lazy dog.
          </Text>
        </div>
        <div className="flex flex-col">
          <Text family="mono" variant="h1">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="h2">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="base">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="body">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="subtext">
            The quick brown fox jumps over the lazy dog.
          </Text>
          <Text family="mono" variant="label">
            The quick brown fox jumps over the lazy dog.
          </Text>
        </div>
      </div>
    </div>
  )
}

export default StratusDasboard

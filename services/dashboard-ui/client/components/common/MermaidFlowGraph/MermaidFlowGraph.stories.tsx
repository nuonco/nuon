import { MermaidFlowGraph } from './MermaidFlowGraph'

export default {
  title: 'Common/MermaidFlowGraph',
}

const simpleGraph = `graph TD
  A[Start] --> B[Process]
  B --> C[End]`

export const Simple = () => <MermaidFlowGraph code={simpleGraph} />

const withEdgeLabels = `graph LR
  A[Request] -->|validate| B[Auth]
  B -->|success| C[Handler]
  B -->|failure| D[Error]`

export const EdgeLabels = () => <MermaidFlowGraph code={withEdgeLabels} />

const styledNodes = `graph TD
  A[OK] --> B[Warning]
  B --> C[Error]
  style A fill:#22c55e,stroke:#16a34a,color:#fff
  style B fill:#eab308,stroke:#ca8a04,color:#000
  style C fill:#ef4444,stroke:#dc2626,color:#fff`

export const StyledNodes = () => <MermaidFlowGraph code={styledNodes} />

const withSubgraphs = `graph TD
  subgraph Frontend
    A[React App] --> B[API Client]
  end
  subgraph Backend
    C[API Server] --> D[Database]
  end
  B --> C`

export const Subgraphs = () => <MermaidFlowGraph code={withSubgraphs} />

const leftToRight = `graph LR
  A[Input] --> B{Decision}
  B -->|yes| C[Action A]
  B -->|no| D[Action B]
  C --> E[Done]
  D --> E`

export const LeftToRight = () => <MermaidFlowGraph code={leftToRight} />

const edgeTypes = `graph TD
  A[Start] --> B[Solid arrow]
  B --- C[No arrow]
  C ==> D[Thick arrow]
  D -.-> E[Dashed arrow]`

export const EdgeTypes = () => <MermaidFlowGraph code={edgeTypes} />

const shapes = `graph LR
  A[Rectangle] --> B(Rounded)
  B --> C{Diamond}
  C --> D[(Cylinder)]
  D --> E((Circle))
  E --> F[[Subroutine]]`

export const Shapes = () => <MermaidFlowGraph code={shapes} />

const architectureExample = `graph TD
  subgraph Cloud Account
    R[Runner] --> K[Kubernetes]
    K --> H[Helm Charts]
    K --> M[Manifests]
    R --> T[Terraform]
    T --> V[VPC]
    T --> E[EKS Cluster]
    T --> I[IAM Roles]
  end
  subgraph Control Plane
    API[Control API] -->|deploy| R
    API -->|status| R
    D[Dashboard] --> API
  end
  D -->|manage| R
  style API fill:#3062D4,stroke:#1e50c0,color:#fff
  style D fill:#3062D4,stroke:#1e50c0,color:#fff
  style R fill:#059669,stroke:#047857,color:#fff`

export const Architecture = () => <MermaidFlowGraph code={architectureExample} />

const nestedSubgraphs = `graph TD
  subgraph VPC["Customer Cloud VPC (AWS)"]
    Runner["Nuon Runner"]
    RDS[("PostgreSQL RDS")]
    ACM["ACM Certificate"]
    ALB["Application Load Balancer"]
    Stack["CloudFormation Stack"]

    subgraph EKS["EKS Cluster"]
      Coder["Coder"]
      Logstream["Kubelogstream"]
      Observability["Grafana & Prometheus Observability"]
      DevEnv["Development Environment"]
    end
  end

  Stack -->|provisions| Runner
  Runner -->|provisions| RDS
  Runner -->|provisions| ACM
  Runner -->|provisions| ALB
  Runner -->|provisions| Coder
  Runner -->|provisions| Logstream
  Runner -->|provisions| Observability
  ACM -->|TLS| ALB
  ALB --> Coder
  RDS -->|DB| Coder
  Coder --> Observability
  Coder --> DevEnv
  Logstream --> DevEnv`

export const NestedSubgraphs = () => <MermaidFlowGraph code={nestedSubgraphs} />

const multipleTopLevel = `graph TD
  subgraph Nuon["Nuon Control Plane"]
    NuonAPI["Nuon API"]
  end

  subgraph Clients["Clients"]
    IDE["IDE with SSH"]
    Dashboard["Coder & Grafana Dashboards & Web IDE"]
  end

  subgraph VPC["Customer Cloud VPC (AWS)"]
    Runner["Nuon Runner"]
    ALB["Application Load Balancer"]

    subgraph EKS["EKS Cluster"]
      Coder["Coder"]
      Logstream["Kubelogstream"]
    end
  end

  NuonAPI -->|generates| Runner
  Runner -->|provisions| Coder
  Runner -->|provisions| Logstream
  Dashboard -->|HTTPS| ALB
  IDE -->|HTTPS| ALB
  ALB --> Coder`

export const MultipleTopLevelSubgraphs = () => <MermaidFlowGraph code={multipleTopLevel} />

const deeplyNested = `graph TD
  subgraph Cloud["Cloud Provider"]
    subgraph Region["us-west-2"]
      subgraph AZ1["Availability Zone A"]
        Web1["Web Server 1"]
        App1["App Server 1"]
      end
      subgraph AZ2["Availability Zone B"]
        Web2["Web Server 2"]
        App2["App Server 2"]
      end
      LB["Load Balancer"]
      DB[("Primary Database")]
    end
  end

  LB --> Web1
  LB --> Web2
  Web1 --> App1
  Web2 --> App2
  App1 --> DB
  App2 --> DB`

export const DeeplyNested = () => <MermaidFlowGraph code={deeplyNested} />

const edgesToSubgraphs = `graph TD
  subgraph Frontend["Frontend"]
    React["React App"]
    Next["Next.js SSR"]
  end

  subgraph Backend["Backend Services"]
    API["REST API"]
    GQL["GraphQL"]

    subgraph Workers["Background Workers"]
      Email["Email Worker"]
      Search["Search Indexer"]
    end
  end

  subgraph Data["Data Layer"]
    PG[("PostgreSQL")]
    Redis["Redis Cache"]
    ES["Elasticsearch"]
  end

  React --> API
  React --> GQL
  Next --> API
  API --> PG
  API --> Redis
  GQL --> PG
  Email --> PG
  Search --> ES
  API -->|enqueue| Email
  API -->|enqueue| Search`

export const CrossSubgraphEdges = () => <MermaidFlowGraph code={edgesToSubgraphs} />

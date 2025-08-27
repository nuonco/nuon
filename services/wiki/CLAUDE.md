# Wiki Service

The **Wiki** service is the internal documentation platform for Nuon, built with Astro Starlight to provide comprehensive documentation for team members, processes, and technical knowledge.

## Service Overview

This is an internal documentation website built with Astro and Starlight, serving as the central knowledge base for the Nuon team. It contains company processes, technical documentation, onboarding materials, and operational procedures.

## Architecture

- **Framework**: Astro with Starlight documentation theme
- **Language**: TypeScript
- **Styling**: Tailwind CSS with Starlight theming
- **Content**: Markdown and MDX files
- **Search**: Built-in search with Starlight
- **Navigation**: Automatic sidebar generation

## Relationship to Other Services

- **Internal Use**: Serves Nuon team members and employees
- **Knowledge Base**: Central repository for company knowledge
- **Onboarding**: New hire documentation and processes
- **Operations**: Runbooks and operational procedures
- **Development**: Technical documentation and guidelines

## Project Structure

### Core Files
- `astro.config.mjs` - Astro configuration with Starlight
- `package.json` - Dependencies and build scripts
- `tailwind.config.mjs` - Tailwind CSS configuration
- `Dockerfile` - Container build configuration

### Key Directories

#### `/src/` - Source Code
- `assets/` - Images and static assets
- `components/` - Custom Astro components
- `content/docs/` - Documentation content in Markdown
- `tailwind.css` - Custom CSS styles

#### `/src/content/docs/` - Documentation Content
Organized by functional areas:
- Company information and culture
- Engineering processes and standards
- Operations and runbooks
- Product documentation
- Team-specific information

#### Infrastructure
- `/infra/` - Terraform deployment configuration
- `/k8s/` - Kubernetes Helm chart
- `service.yml` - Service configuration

## Key Features

### Documentation Platform
- Markdown-based content authoring
- Automatic table of contents generation
- Code syntax highlighting
- Image optimization and handling

### Search & Navigation
- Built-in search functionality
- Automatic sidebar navigation
- Breadcrumb navigation
- Mobile-responsive design

### Starlight Features
- Dark/light mode support
- SEO optimization
- Accessibility features
- Fast static site generation

### Team Collaboration
- Git-based content workflow
- Easy content contribution process
- Version control for documentation
- Review process for changes

## Development

### Setup
```bash
cd services/wiki
npm install
npm run dev
```

### Key Scripts
- `npm run dev` - Development server
- `npm run build` - Production build
- `npm run test` - Astro validation
- `npm run lint` - Code linting

### Content Management
- Documentation written in Markdown/MDX
- Git-based workflow for contributions
- Automatic deployment on content changes
- Review process for sensitive information

## Content Organization

### Company Documentation
- Mission, values, and culture
- Organizational structure
- Policies and procedures
- Meeting notes and decisions

### Engineering Documentation
- Architecture decisions
- Development workflows
- Code standards and guidelines
- Infrastructure documentation

### Operations
- Incident response procedures
- Monitoring and alerting
- Deployment procedures
- Troubleshooting guides

### Team Information
- Individual team member pages
- Role descriptions and responsibilities
- Contact information
- Skills and expertise areas

## Deployment

### Kubernetes Deployment
- Containerized with Docker
- Deployed via Helm chart
- Internal access only
- SSL termination and authentication

### Infrastructure
- Terraform-managed infrastructure
- Internal DNS and routing
- Monitoring and logging integration
- Backup and disaster recovery

## Configuration

### Starlight Configuration
- Site metadata and branding
- Navigation structure
- Search configuration
- Theme customization

### Access Control
- Internal-only access
- Authentication integration
- Role-based content access
- Security policies

## Technologies Used

### Core Framework
- **Astro**: Modern static site generator
- **Starlight**: Documentation-focused Astro theme
- **TypeScript**: Type safety and developer experience

### Content & Styling
- **Markdown/MDX**: Content authoring format
- **Tailwind CSS**: Utility-first styling
- **Sharp**: Image optimization

### Development Tools
- **Astro Check**: Validation and type checking
- **ESLint**: Code quality
- **Docker**: Containerization

This service serves as the central knowledge repository for the Nuon team, providing easy access to company information, technical documentation, and operational procedures in a searchable, well-organized format.
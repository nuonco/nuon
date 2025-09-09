# Nuon CLI

The **Nuon CLI** is the public-facing command-line interface for developers using the Nuon platform. It provides a comprehensive set of commands for managing applications, components, deployments, and development workflows.

## Binary Overview

This is the primary CLI tool that developers use to interact with the Nuon platform. It's distributed as a standalone binary and provides an intuitive interface for all platform operations from local development to production deployment management.

## Architecture

- **Language**: Go
- **Framework**: Cobra CLI framework
- **Distribution**: Standalone binary compiled for multiple platforms
- **Authentication**: OAuth2 device flow with JWT tokens
- **API Integration**: Communicates with `ctl-api` service
- **Configuration**: Local config management with org/app context

## Relationship to Other Services

- **Primary API Consumer**: Communicates with `ctl-api` for all operations
- **Authentication**: Uses Auth0 for user authentication
- **Development Workflow**: Integrates with local development processes
- **Dashboard Complement**: CLI alternative to `dashboard-ui` web interface
- **Runner Coordination**: Manages deployments executed by `runner` binaries

## Project Structure

### Core Files
- `main.go` - CLI entry point
- `install.sh` - Installation script for distribution
- `Dockerfile` - Container build for CLI distribution
- `service.yml` - Service configuration

### Key Directories

#### `/cmd/` - Command Definitions
Cobra command definitions organized by functionality:
- `root.go` - Root command and global flags
- `apps.go` - Application management commands
- `components.go` - Component operations
- `builds.go` - Build management
- `installs.go` - Installation operations
- `actions.go` - Action workflow commands
- `secrets.go` - Secret management
- `orgs.go` - Organization operations
- `login.go` - Authentication commands
- `dev.go` - Development workflow commands
- `docs.go` - Documentation commands
- `version.go` - Version information

#### `/internal/` - Business Logic

##### Core Services
- `apps/` - Application management logic
- `components/` - Component operations
- `builds/` - Build management
- `installs/` - Installation operations
- `actions/` - Action workflow management
- `secrets/` - Secret operations
- `orgs/` - Organization management
- `releases/` - Release management

##### Development Workflow
- `dev/` - Development mode and workflow automation
- `auth/` - Authentication and token management
- `config/` - Configuration management
- `lookup/` - Resource lookup and resolution

##### User Interface
- `ui/` - CLI user interface components
  - `v2/` - Next-generation UI components
- Terminal output formatting
- Interactive prompts and spinners
- JSON and TOML output formats

## Key Features

### Application Management
- **App Creation**: `nuon apps create` with interactive setup
- **Configuration**: Manage app configs with validation
- **Sync**: Sync local directory with remote configuration
- **Validation**: Local validation before deployment

### Component Management
- **Component Operations**: Create, update, delete components
- **Build Management**: Trigger and monitor builds
- **Configuration**: Manage component configurations
- **Dependencies**: Handle component dependencies

### Installation & Deployment
- **Install Management**: Create and manage installations
- **Deployment**: Deploy components to installations
- **Monitoring**: Monitor deployment progress and status
- **Logs**: Stream deployment and application logs

### Development Workflow (`nuon dev`)
- **Local Development**: Enhanced development mode
- **Git Integration**: Git branch and change detection
- **Auto-sync**: Automatic configuration synchronization
- **Build Monitoring**: Monitor component builds
- **Deploy Polling**: Watch deployment status

### Organization Management
- **Multi-org Support**: Switch between organizations
- **Team Management**: Invite and manage team members
- **API Tokens**: Generate and manage API tokens
- **VCS Integration**: Connect GitHub and other VCS

### Authentication & Security
- **OAuth2 Flow**: Device code flow for secure login
- **Token Management**: Automatic token refresh
- **Multi-org Auth**: Organization-scoped authentication
- **API Key Support**: Alternative API key authentication

## Command Categories

### Core Commands
```bash
nuon apps          # Application management
nuon components    # Component operations  
nuon installs      # Installation management
nuon builds        # Build operations
nuon actions       # Action workflows
```

### Development Commands
```bash
nuon dev           # Development mode
nuon init          # Initialize new projects
nuon login         # Authentication
```

### Organization Commands
```bash
nuon orgs          # Organization management
nuon secrets       # Secret management
nuon docs          # Documentation
```

## Development

### Setup
```bash
cd bins/cli
go build -o nuon .
./nuon --help
```

### Code Quality
**IMPORTANT: Always run these commands after making code changes:**
```bash
# Format code to Go standards
go fmt ./...

# Check for common Go issues and lint problems
go vet ./...

# Build to verify compilation
go build -o nuon .
```

**When to run code quality checks:**
- After editing any Go files
- Before committing changes
- When adding new packages or dependencies
- When refactoring existing code

### Installation
Users can install via:
- Installation script: `curl -sSL install.nuon.co | bash`
- Package managers (Homebrew, etc.)
- Direct binary download
- Docker container

### Configuration
- Config stored in `~/.nuon/`
- Organization and app context management
- Token storage and refresh
- Local preferences

## Key Workflows

### Getting Started
1. **Login**: `nuon login` - Authenticate with Nuon platform
2. **Select Org**: `nuon orgs select` - Choose organization
3. **Create App**: `nuon apps create` - Create new application
4. **Add Components**: `nuon components create` - Add components

### Development Workflow
1. **Dev Mode**: `nuon dev` - Start development mode
2. **Make Changes**: Edit configurations locally
3. **Auto-sync**: CLI automatically syncs changes
4. **Monitor**: Watch builds and deployments
5. **Deploy**: Deploy to installations

### Deployment Workflow
1. **Create Install**: `nuon installs create` - Create installation
2. **Deploy**: `nuon installs deploy` - Deploy components
3. **Monitor**: `nuon installs logs` - Monitor deployment
4. **Manage**: Update inputs, manage components

## Integration Features

### Git Integration
- Detects Git repository and branch information
- Warns about uncommitted changes
- Integrates with VCS connections
- Branch-based development workflows

### Configuration Management
- TOML configuration file support
- Local validation and syntax checking
- Template rendering and variable substitution
- Configuration synchronization

### API Integration
- Automatic API client generation
- Error handling and retry logic
- Streaming logs and real-time updates
- Efficient polling and caching

## Distribution & Installation

### Binary Distribution
- Multi-platform builds (Linux, macOS, Windows)
- GitHub Releases with automatic builds
- Package manager integration
- Docker images for containerized usage

### Auto-update
- Version checking and update notifications
- Seamless binary updates
- Backward compatibility checks
- Migration assistance for breaking changes

## Technologies Used

### Core Technologies
- **Go**: Primary language with robust CLI libraries
- **Cobra**: CLI framework for command structure
- **Viper**: Configuration management
- **Survey**: Interactive prompts and forms

### Integration Libraries
- **HTTP Client**: API communication with retry logic
- **JWT**: Token handling and validation
- **Git**: Repository interaction and status
- **TOML**: Configuration file parsing

### UI/UX Libraries
- **Spinner**: Progress indicators
- **Table**: Structured data display
- **Colors**: Terminal color support
- **Prompt**: Interactive user input

This CLI serves as the primary interface for developers to interact with the Nuon platform, providing a powerful and intuitive command-line experience that complements the web dashboard and enables efficient development workflows.
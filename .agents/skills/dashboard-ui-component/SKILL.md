---
name: dashboard-ui-component
description: Use this skill when adding or using a UI component in the dashboard-ui client/ SPA.
model: sonnet
color: blue
---

This skill enforces checking existing components before creating new ones and reading TypeScript interfaces before using component props.

## Steps

1. Run `ls client/components/common/` and the relevant domain directory (e.g., `client/components/actions/`, `client/components/installs/`). Read the filenames.
2. If an existing component meets your needs, use it. Do NOT create a new component that duplicates an existing one.
3. Before writing JSX that uses a component, read its `interface I*` props definition in the component's source file. Do not guess prop names.
4. For Modal or Panel: always use `Modal` or `Panel` from `client/components/surfaces/`. Never use `ModalBase` or `PanelBase` directly.
5. For a new primitive (Button, Badge, Text, etc.): place the file flat in `client/components/common/MyComponent.tsx`.
6. For a new domain-specific component: place the file in `client/components/<domain>/MyComponent.tsx`.
7. Use a directory (`client/components/common/MyComponent/`) only when the component has internal sub-components that should not be exported directly.

## Anti-Patterns

- **Do not** create `LoadingSpinner.tsx` if `Spinner.tsx` already exists in `common/` — always check existing components first
- **Do not** pass props to a component without reading its interface — wrong props cause runtime errors that TypeScript may not catch at the call site if the type is broad
- **Do not** use `ModalBase` or `PanelBase` directly — always use the `Modal`/`Panel` wrappers from `surfaces/`
- **Do not** put a domain-specific component (e.g., `InstallCard`) into `client/components/common/`

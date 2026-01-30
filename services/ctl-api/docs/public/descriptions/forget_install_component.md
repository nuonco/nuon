# Forget Install Component

Permanently forget (soft delete) an install component from the system. This operation marks the install component as deleted while preserving the record for audit purposes.

## Use Cases

- Remove a component that is no longer needed from an install
- Clean up failed or orphaned install components
- Prepare for reinstalling a component from scratch

## Important Notes

- This is a **soft delete** operation - the record is marked as deleted but remains in the database
- The component will no longer appear in API responses or dashboard views
- Associated resources (terraform state, deploys, etc.) are preserved via soft delete
- This operation is **irreversible** via the API
- To restore, database-level operations would be required

## Prerequisites

- Install must exist and belong to the authenticated organization
- Component must exist for the specified install
- User must have appropriate permissions for the install's organization
- Component must be removed from the app configuration (sync required)

## Behavior

1. Validates install exists and belongs to org
2. Validates install component exists
3. Validates component is not in the app configuration
4. Soft deletes the install component record
5. Cascades soft delete to associated resources (via GORM associations)
6. Sends event loop signal for any cleanup workflows
7. Returns success response

## Validation

Before forgetting an install component, the system validates that the component no longer exists in the app configuration. If the component is still in the app config, the request will fail with a user-friendly error message.

**To resolve this error:**

1. Remove the component from your `nuon.yaml` file
2. Run `nuon apps sync` to update the app configuration
3. Retry the forget operation

## Related Endpoints

- `DELETE /v1/installs/{install_id}` - Delete entire install
- `POST /v1/installs/{install_id}/forget` - Forget entire install
- `POST /v1/installs/{install_id}/components/{component_id}/teardown` - Teardown component infrastructure

/* the stack version each install stack is actually running: its applied pointer, or
   for legacy stacks without one, the newest active version. A pointer needs a nonempty
   applied_config_id; a pointer without one matches neither branch and resolves to
   nothing rather than reviving a historical version. The legacy branch orders by
   created_at, like getActiveStackVersion: updated_at also moves when a version's
   phone-home token is re-minted, which says nothing about what was applied */
SELECT
    live.*
FROM (
    SELECT
        v.id,
        v.created_by_id,
        v.created_at,
        v.updated_at,
        v.deleted_at,
        v.org_id,
        stack.install_id,
        stack.id AS install_stack_id,
        v.app_config_id,
        v.status,
        v.contents,
        v.checksum,
        v.template_url,
        v.stack_name,
        v.phone_home_id,
        v.phone_home_url,
        v.phone_home_token_id,
        v.phone_home_token_revoked_at,
        v.phone_home_identity_name,
        v.aws_bucket_name,
        v.aws_bucket_key,
        v.aws_stack_parameters,
        v.custom_stacks_template_url,
        v.custom_stacks_aws_bucket_key,
        v.custom_stacks_output_map,
        v.custom_stacks_input_parameters_map,
        v.quick_link_url,
        v.quick_link_bucket_key,
        v.quick_link_uidef_bucket_key,
        v.terraform_contents,
        v.terraform_checksum,
        v.callback_ref,
        v.composite_error
    FROM
        install_stacks stack
        JOIN install_stack_versions v ON v.id = stack.app_config_ref ->> 'applied_config_by_id'
    WHERE
        stack.deleted_at = 0
        AND stack.app_config_ref ->> 'applied_config_by_type' = 'install_stack_versions'
        AND COALESCE(stack.app_config_ref ->> 'applied_config_id', '') <> ''
        AND v.deleted_at = 0
    UNION ALL
    (SELECT
        pick.id,
        pick.created_by_id,
        pick.created_at,
        pick.updated_at,
        pick.deleted_at,
        pick.org_id,
        pick.install_id,
        pick.install_stack_id,
        pick.app_config_id,
        pick.status,
        pick.contents,
        pick.checksum,
        pick.template_url,
        pick.stack_name,
        pick.phone_home_id,
        pick.phone_home_url,
        pick.phone_home_token_id,
        pick.phone_home_token_revoked_at,
        pick.phone_home_identity_name,
        pick.aws_bucket_name,
        pick.aws_bucket_key,
        pick.aws_stack_parameters,
        pick.custom_stacks_template_url,
        pick.custom_stacks_aws_bucket_key,
        pick.custom_stacks_output_map,
        pick.custom_stacks_input_parameters_map,
        pick.quick_link_url,
        pick.quick_link_bucket_key,
        pick.quick_link_uidef_bucket_key,
        pick.terraform_contents,
        pick.terraform_checksum,
        pick.callback_ref,
        pick.composite_error
    FROM (
        SELECT DISTINCT ON (stack.install_id, stack.id)
            v.id,
            v.created_by_id,
            v.created_at,
            v.updated_at,
            v.deleted_at,
            v.org_id,
            stack.install_id,
            stack.id AS install_stack_id,
            v.app_config_id,
            v.status,
            v.contents,
            v.checksum,
            v.template_url,
            v.stack_name,
            v.phone_home_id,
            v.phone_home_url,
            v.phone_home_token_id,
            v.phone_home_token_revoked_at,
            v.phone_home_identity_name,
            v.aws_bucket_name,
            v.aws_bucket_key,
            v.aws_stack_parameters,
            v.custom_stacks_template_url,
            v.custom_stacks_aws_bucket_key,
            v.custom_stacks_output_map,
            v.custom_stacks_input_parameters_map,
            v.quick_link_url,
            v.quick_link_bucket_key,
            v.quick_link_uidef_bucket_key,
            v.terraform_contents,
            v.terraform_checksum,
            v.callback_ref,
            v.composite_error
        FROM
            install_stacks stack
            JOIN install_stack_versions v ON v.install_stack_id = stack.id
        WHERE
            COALESCE(stack.app_config_ref ->> 'applied_config_by_id', '') = ''
            AND stack.deleted_at = 0
            AND v.deleted_at = 0
            AND v.status ->> 'status' = 'active'
        ORDER BY
            stack.install_id,
            stack.id,
            v.created_at DESC,
            v.updated_at DESC
    ) pick)
) live

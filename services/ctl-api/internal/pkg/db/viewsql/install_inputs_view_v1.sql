/*
install inputs fields + new comlumn `values_redacted` for consumption by public clients.
the values_redacted column maps

create an view that selects all of the regular inputs PLUS a column `values_redacted`. `values_redacted`
is a column that makes an hstore out of a subquery.

the subquery maps the name to a string value. the string value depends on wether the name/key is sensitive:
1. if NOT: we return the actual value of the respective key in the install_inputs value hstore
2. if YES: we return eight asterisks, aka a redaction.

*/
SELECT
    install_inputs.*,
    (
        SELECT
            hstore(
                array_agg(name),
                array_agg(
                    CASE
                        WHEN sensitive IS TRUE THEN '********'
                        WHEN sensitive IS FALSE THEN install_inputs.values -> app_inputs.name :: text
                        ELSE ''
                    END
                )
            )
        FROM
            app_inputs
        WHERE
            install_inputs.app_input_config_id = app_input_config_id
    ) AS values_redacted
FROM
    install_inputs

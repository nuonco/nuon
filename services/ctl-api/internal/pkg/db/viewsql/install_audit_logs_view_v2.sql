SELECT 
	install_id as install_id,
	'install_action_workflow_run' AS "type",
	created_at as time_stamp,
	row_to_json(iawr) AS log_line
FROM public.install_action_workflow_runs AS iawr

UNION ALL

SELECT 
	install_id AS install_id,
	'install_sandbox_run' AS "type",
	created_at as time_stamp,
	row_to_json(isr) AS log_line
FROM public.install_sandbox_runs AS isr

UNION ALL

SELECT 
	icmp.install_id AS install_id,
	'install_deploy' AS "type",
	idp.created_at as time_stamp,
	row_to_json(idp) AS log_line
FROM public.install_deploys AS idp
JOIN public.install_components AS icmp
ON idp.install_component_id = icmp.id

UNION ALL

SELECT 
	pr.install_id AS install_id,
	'policy_report' AS "type",
	pr.evaluated_at AS time_stamp,
	row_to_json(pr) AS log_line
FROM public.policy_reports AS pr
WHERE pr.install_id IS NOT NULL
  AND pr.deleted_at = 0

ORDER BY time_stamp ASC

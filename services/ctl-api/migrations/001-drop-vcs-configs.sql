DELETE FROM vcs_connections
WHERE id NOT IN
(
SELECT MIN(id)
FROM vcs_connections
GROUP BY org_id, github_install_id
)

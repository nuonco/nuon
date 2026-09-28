import httpx
from click.testing import CliRunner

from nuon_ext_install_config_history.cli import main
from nuon_ext_install_config_history.client import NuonAPIError, NuonClient, NuonContext
from nuon_ext_install_config_history.render import (
    attach_run_summaries,
    branch_id,
    run_summary,
)


def test_config_history_uses_install_branch_and_paginates():
    requests = []

    def handler(request: httpx.Request) -> httpx.Response:
        requests.append(request)
        offset = request.url.params.get("offset")
        if offset == "0":
            return httpx.Response(
                200,
                json=[{"id": "cfg-2"}],
                headers={"X-Nuon-Page-Next": "true"},
            )
        return httpx.Response(
            200,
            json=[{"id": "cfg-1"}],
            headers={"X-Nuon-Page-Next": "false"},
        )

    context = NuonContext("https://api.example.test", "token", "org-1", "ins-1")
    with NuonClient(context, transport=httpx.MockTransport(handler)) as client:
        configs = client.config_history("app-1", "branch-1")

    assert configs == [{"id": "cfg-2"}, {"id": "cfg-1"}]
    assert [request.url.path for request in requests] == [
        "/v1/apps/app-1/branches/branch-1/configs",
        "/v1/apps/app-1/branches/branch-1/configs",
    ]
    assert requests[0].headers["Authorization"] == "Bearer token"
    assert requests[0].headers["X-Nuon-Org-ID"] == "org-1"


def test_config_history_falls_back_for_legacy_install():
    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.path == "/v1/apps/app-1/configs"
        return httpx.Response(200, json=[], headers={"X-Nuon-Page-Next": "false"})

    context = NuonContext("https://api.example.test", "token", "org-1", "ins-1")
    with NuonClient(context, transport=httpx.MockTransport(handler)) as client:
        assert client.config_history("app-1", None) == []


def test_branch_id_accepts_api_null_string_shapes():
    assert branch_id({"app_branch_id": "branch-1"}) == "branch-1"
    assert branch_id({"app_branch_id": {"String": "branch-2"}}) == "branch-2"
    assert branch_id({"app_branch": {"id": "branch-3"}}) == "branch-3"


def test_components_are_scoped_to_branch():
    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.params["branch_id"] == "branch-1"
        return httpx.Response(200, json=[], headers={"X-Nuon-Page-Next": "false"})

    context = NuonContext("https://api.example.test", "token", "org-1", "ins-1")
    with NuonClient(context, transport=httpx.MockTransport(handler)) as client:
        assert client.components("app-1", "branch-1") == []


def test_pagination_has_a_safety_cap():
    def handler(_: httpx.Request) -> httpx.Response:
        return httpx.Response(
            200,
            json=[{"id": "cfg"}],
            headers={"X-Nuon-Page-Next": "true"},
        )

    context = NuonContext("https://api.example.test", "token", "org-1", "ins-1")
    with NuonClient(context, transport=httpx.MockTransport(handler)) as client:
        try:
            client._offset_pages("/configs", limit=1, max_pages=2)
        except NuonAPIError as exc:
            assert "exceeded 2 pages" in str(exc)
        else:
            raise AssertionError("expected pagination cap error")


def test_subcommand_help_does_not_require_auth():
    result = CliRunner().invoke(
        main,
        ["history", "--help"],
        env={"NUON_API_TOKEN": "", "NUON_ORG_ID": "", "NUON_INSTALL_ID": ""},
    )
    assert result.exit_code == 0
    assert "List the selected install" in result.output


def test_run_summary_prefers_pr_and_commit():
    workflow = {
        "name": "Manual run",
        "type": "app_branches_config_repo_update",
        "plan_only": False,
        "app_branch_runs": [
            {
                "app_config_id": "cfg-1",
                "head_sha": "abcdef1234567890",
                "pr_number": 42,
                "metadata": {"trigger": "pull_request", "pr_number": 42},
                "vcs_connection_commit": {
                    "sha": "abcdef1234567890",
                    "message": "Fix helm timeout\n\nmore detail",
                },
            }
        ],
    }
    assert run_summary(workflow) == "PR #42 · Fix helm timeout (abcdef1)"


def test_run_summary_uses_tag_and_preview():
    workflow = {
        "type": "app_branches_manual_update",
        "plan_only": True,
        "app_branch_runs": [
            {
                "app_config_id": "cfg-2",
                "run_type": "git-preview-run",
                "metadata": {"trigger": "tag", "tag": "v1.2.3"},
            }
        ],
    }
    assert run_summary(workflow) == "Tag v1.2.3 · preview"


def test_attach_run_summaries_matches_config_ids():
    configs = [{"id": "cfg-1"}, {"id": "cfg-orphan"}]
    workflows = [
        {
            "name": "Config update",
            "app_branch_runs": [{"app_config_id": "cfg-1", "metadata": {"trigger": "push"}}],
        }
    ]
    attach_run_summaries(configs, workflows)
    assert configs[0]["run_summary"] == "Config update"
    assert "run_summary" not in configs[1]

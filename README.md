# Links

## Public Sites

* [web](https://www.nuon.co)
* [twitter](https://twitter.com/nuoninc)
* [github](https://github.com/nuonco)
* [linkedin](https://www.linkedin.com/company/nuonco)

## Project Management

- [marketing Notion](https://www.notion.so/nuon/Marketing-e623d43b049d45f89787329a1bace156)
- [Product board](https://github.com/orgs/powertoolsdev/projects/16/views/2)
- [Sales board](https://github.com/orgs/powertoolsdev/projects/28)
- [Customer board](https://github.com/orgs/powertoolsdev/projects/29/views/1)

## Developer Resources

- [Wiki](https://github.com/powertoolsdev/mono/tree/main/wiki)
- [Operations](./wiki/operations.md)
- [admin panel prod](http://ctl.nuon.us-west-2.prod.nuon.cloud/docs/index.html)
- [admin panel stage](http://ctl.nuon.us-west-2.stage.nuon.cloud/docs/index.html)
- [AWS SSO](https://nuon.awsapps.com/start/#/)

## Docs / SDKs

* [docs](https://docs.nuon.co)
* [changelog](https://docs.nuon.co/changelog)
* [sdks](https://docs.nuon.co/sdks)
* [terraform](https://registry.terraform.io/providers/nuonco/nuon/latest/docs)

## Debugging

For debugging temporal:
- [temporal ui prod](http://temporal-ui.nuon.us-west-2.prod.nuon.cloud:8080/namespaces/orgs/workflows)
- [temporal ui stage](http://temporal-ui.nuon.us-west-2.stage.nuon.cloud:8080/namespaces/orgs/workflows)
- [temporal server internals](https://us5.datadoghq.com/dash/integration/671/temporal-server-overview)
- [temporal
  dashboard](https://us5.datadoghq.com/dashboard/8k6-7vv-y88/temporal?refresh_mode=sliding&view=spans&from_ts=1707889654953&to_ts=1707893254953&live=true)


For debugging workers/viewing logs:
- [workers-executors logs](https://us5.datadoghq.com/logs/livetail?query=service%3Aworkers-executors%20)
- [ctl-api
  logs](https://us5.datadoghq.com/logs/livetail?query=container_name%3Actl-api-api%20-%2Flivez%20-%2Freadyz%20-%22marking%20request%20as%20public%22%20-%2Ffavicon%20)
- [apis
  dashboard](https://us5.datadoghq.com/dashboard/4md-hrc-wdy/apis?refresh_mode=sliding&from_ts=1700600980740&to_ts=1700604580740&live=true)
- [orgs
  dashboard](https://us5.datadoghq.com/dashboard/6q2-svy-xpj?cols=host%2Cservice&refresh_mode=sliding&tpl_var_env%5B0%5D=%2A&view=spans&from_ts=1707250827640&to_ts=1707337227640&live=true)

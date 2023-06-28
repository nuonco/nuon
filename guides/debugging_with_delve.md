# How to: debug with delve and VS Code

Our golang services can be launched with the [delve](https://github.com/go-delve/delve) debugger, which runs a network service, and VS Code can attach "remotely" to that service to set breakpoints and step through the code.

Note that with the monorepo, debugging can be unusably slow if you launch VS Code from the monorepo root, so only open the directory containing the specific golang program you want to debug, for example `services/api` or `services/orgs-api` for best results.

Configure a VS Code launch configuration in `services/api/.vscode/launch.json` (or the directory for the service you are working with) similar to the following:

```json
{
  // Use IntelliSense to learn about possible attributes.
  // Hover to view descriptions of existing attributes.
  // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
  "version": "0.2.0",
  "configurations": [
    {
      "name": "delve localhost:2345",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "port": 2345,
      "host": "localhost",
      "apiVersion": 1
    }
  ]
}
```

Then, in a terminal shell, launch it under the delve debugger. Here's an example for the `orgs-api`

```bash
cd services/orgs-api
nuonctl service exec --name=orgs-api -- dlv debug --headless --listen 127.0.0.1:2345 . -- server
```

Delve should get prepared and stop the program just prior to `main()` (I think). It will not start the program until a remote debug client attaches to the debugger process. So don't worry if you don't see any startup log output yet.

In VS Code, activate the "Run and Debug" activity from the activity bar (ctrl+shft+d). You should see a menu of launch configurations including the one we defined in our `launch.json` above. Choose that one and click the play button.

You should now be good to set breakpoints and do typical graphical debugger investigation tasks.


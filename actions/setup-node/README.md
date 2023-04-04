# action-setup-node

Github action to set up node projects.

## Usage

Add this action to your typescript / node.js CI action to consolidate node.js setup, node module caching & installing dependencies.

``` yaml
unit-test:
  name: Unit test
  runs-on: ubuntu-latest
  steps:
    - name: Checkout repo
      uses: actions/checkout@v3

    - name: Set up Node.js
      uses: powertoolsdev/action-setup-node@main

    - name: Unit test
      run: npm run test:unit
```


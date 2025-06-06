# Contributing to PipeCD

If you're interested in contributing to PipeCD, this document will provide clear instructions on how to get involved.

> Note: Don't bother reading policies and flows, just want to manipulate the code? Jump to the [Development](#development) section.

The [Open Source Guides](https://opensource.guide/) website offers a collection of resources for individuals, communities, and companies who want to learn how to run and contribute to an open source project. Both contributors and newcomers to open source will find the following guides especially helpful:

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Building Welcoming Communities](https://opensource.guide/building-community/)

## Code of Conduct

PipeCD follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). Please read it to understand which actions are acceptable and which are not.

## Get Involved

There are various ways to contribute to PipeCD, and many of them don't involve writing code. Here are a few ideas to help you get started:

- Start by using PipeCD. Follow the [Quickstart](https://pipecd.dev/docs/quickstart/) guide. Does everything work as expected? If not, we're always looking for improvements. Let us know by [opening an issue](#issues).
- Browse through the [open issues](https://github.com/pipe-cd/pipecd/issues). Provide workarounds, ask for clarification, or suggest labels.
- If you find an issue you'd like to fix, [open a pull request](#pull-requests). Issues labeled as [_Good first issue_](https://github.com/pipe-cd/pipecd/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) are a good starting point.
- Read the [PipeCD docs](https://pipecd.dev/docs/). If you come across anything that is confusing or can be improved, click "Edit this page" on the right side of most docs to propose changes through the GitHub interface.
- Participate in [Discussions](https://github.com/pipe-cd/pipecd/discussions) and share your ideas.

Contributions are highly welcome. If you need help planning your contribution, please reach out to us on Twitter at [@pipecd_dev](https://twitter.com/pipecd_dev) and let us know you're seeking assistance.

### Join our Slack Channel

We have a `#pipecd` channel on [CNCF Slack](https://cloud-native.slack.com/) for discussions related to PipeCD development. You can also provide valuable help by assisting other users in the channel.

## Our Development Process

PipeCD uses [GitHub](https://github.com/pipe-cd/pipecd) as its source of truth. The core team will be working directly there. All changes will be public from the beginning.

All pull requests undergo checks by the continuous integration system, GitHub Actions. These checks include unit tests, lint tests, and more.

### Branch Organization

PipeCD has one primary branch `master`.

## Open a new issue

When [opening a new issue](https://github.com/pipe-cd/pipecd/issues/new/choose), please make sure to fill out the issue template. This step is crucial! Neglecting to do so may result in your issue not being promptly addressed. If this happens, feel free to open a new issue once you have gathered all the necessary information.

### Bugs

We use [GitHub Issues](https://github.com/pipe-cd/pipecd/issues) for our public bug reports. If you encounter a problem, take a look around to see if someone has already reported it. If you believe you have found a new, unreported bug, you can submit a [bug report](https://github.com/pipe-cd/pipecd/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug-report.md).

- **One issue, one bug:** Please report a single bug per issue.
- **Provide reproduction steps:** List all the steps necessary to reproduce the issue. The person reading your bug report should be able to follow these steps with minimal effort.

If you are only fixing a bug, you can submit a pull request right away, but we still recommend filing an issue to describe what you are fixing. This is helpful in case we do not accept that specific fix but still want to track the issue.

### Security Bugs

If you discover security-related bugs that may compromise the security of current users, please send a direct message to our maintainers on Slack or Twitter instead of opening a public issue.

You can find our DM contacts via [MAINTAINERS.md](./MAINTAINERS.md).

### Enhancement requests
If you would like to request an enhancement to existing features, you can file an issue with the [enhancement request template](https://github.com/pipe-cd/pipecd/issues/new?assignees=&labels=kind%2Fenhancement&projects=&template=enhancement.md).

### Feature requests

If you would like to request an entirely new feature, you can file an issue with the [feature request](https://github.com/pipe-cd/pipecd/issues/new?assignees=&labels=kind%2Ffeature&projects=&template=new-feature.md).

## Working on issues

### Good first issues

We maintain a list of [good first issues](https://github.com/pipe-cd/pipecd/labels/good%20first%20issue) to help you get started with the PipeCD codebase and familiarize yourself with our contribution process. It's an excellent place to begin.

Additionally, we recommend you not to work on multiple good first issues because it's for first contributers and currently we cannot make enough good first issues.

### Before you work on issues

If you want to work on any of these issues, simply leave a message saying "I'd like to work on this," and we will assign the issue to you and update its status as "claimed." We expect you to submit a pull request within seven days so that we can assign the issue to someone else if you are unavailable.

We recommend you to focus only one issue at once if you are newcomer contributer.

So you've decided to contribute back to the upstream by opening a pull request. You've put in a significant amount of time, and we appreciate your effort. We will do our best to work with you and review the pull request.

### Investigate an issue

Before you submitting a Pull Request, we recommend you to investigate the issue and comment what to do on it.
Then you can discuss how to solve the issue and reduce the communication on the Pull Request.

### Submitting a Pull Request

Are you working on your first Pull Request? You can learn how to do it from this free video series:

[**How to Contribute to an Open Source Project on GitHub**](https://egghead.io/courses/how-to-contribute-to-an-open-source-project-on-github)

When submitting a pull request, please ensure the following:

- **Issue assignment.** To avoid redundant work, make sure you are assigned to the issue.
- **Keep your PR small.** Small pull requests (~300 lines of diff) are much easier to review and are more likely to get merged. Make sure the PR addresses only one thing. If not, please split it.
- **Use descriptive titles.** It is recommended to follow the [commit message style](#commit-messages).
- **DCO.** If you haven't signed off already, check the [Contributor License Agreement](#contributor-license-agreement).
- **Run `make check`**. To ensure your change will pass the CI.

All pull requests should be opened against the `master` branch.

We have various integration systems that run automated tests to prevent mistakes. The maintainers will also review your code and fix obvious issues. These systems are in place to minimize your worries about the process. Your code contributions are more important than adhering to strict procedures, although completing the checklist will undoubtedly save everyone's time.

### Commit Messages

Commit messages should be simple and use easy words that indicate the focus of the commit and its impact on other developers. Summary in the present tense. Use capital case in the first character but do not use title case.

Example

```
Add imports to Terraform plan result
```

Don't stress too much about PR titles. The maintainers will help you get the title right.

### Licensing

By contributing to PipeCD, you agree that your contributions will be licensed under the Apache License Version 2. Include the following header at the top of your new file(s):

```go
// Copyright 2025 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
```

### Release Note and Breaking Changes

If your change introcudes a user-facing change, please update the following section in your PR description.

```md
**Does this PR introduce a user-facing change?**:
- **How are users affected by this change**:
- **Is this breaking change**:
- **How to migrate (if breaking change)**:
```

Note that if it's a new breaking change, make sure to complete the two latter questions.

## Development

PipeCD consists of several components and docs:

- **cmd/pipecd**: A centralized component that manages deployment data and provides a gRPC API for connecting pipeds, as well as web functionalities such as authentication. [README.md](./cmd/pipecd/README.md)
- **cmd/piped**: piped is an agent component that runs in your cluster. [README.md](./cmd/piped/README.md)
- **cmd/pipectl**: The command-line tool for PipeCD. [README.md](./cmd/pipectl/README.md)
- **cmd/launcher**: The command executor that enables the remote upgrade feature of the piped agent. [README.md](./cmd/launcher/README.md)
- **web**: The web application provided by the control plane. [README.md](./web/README.md)
- **docs**: Documentation and references. [README.md](./docs/README.md)

**You can find detailed development information in the README file of each directory.**

Note: While working with the PipeCD codebase, you may refer to [Makefile](./Makefile) for useful commands.

### Starting a Local Development Environment

#### Update dependencies

Run `make update/go-deps` and `make update/web-deps` to update the dependencies. Starting a local development environment might fail with errors if the dependencies are not up to date.

#### Starting a local registry

In order to start a local development environment, a registry needs to be running locally. 

Run `make kind-up` to start a local registry. 

This will create the kubernetes namespace `pipecd` if it does not exist and start a local registry in the namespace which can then be accessed by other components.

When cleaning up, run `make kind-down` to stop and delete the registery and the cluster.

#### Run PipeCD Control Plane

Run `make run/pipecd` to run PipeCD Control Plane using your local code changes. This will build and run PipeCD Control Plane.

Run `make stop/pipecd` to stop PipeCD Control Plane.

#### Port Forward

Run `kubectl port-forward -n pipecd svc/pipecd 8080` forward your local port to the `pipecd` pod port. 

#### Access the PipeCD UI

After port-forwarding, you can now access the PipeCD Control Plane console at `http://localhost:8080?project=quickstart`.

To login, you can use the configured static admin account as below:

- username: `hello-pipecd`
- password: `hello-pipecd`

#### Run Piped Agent

1. Make sure that PipeCD Control Plane is running and you can access the UI and login.

2. Access to Control Plane console, go to Piped list page - click the three vertical dots on the top right corner and then click on settings. After clicking on settings you will land on the Piped settings page. Next, add a new piped. 

Alternatively, you can go to `http://localhost:8080/settings/piped?project=quickstart`, please adjust the port and the project in the url if they are different from default. 

Then, copy generated Piped ID and base64 key for `piped-config.yaml`

3. Create the piped configuration file `piped-config.yaml`. This is an example configuration. Use the PipeD ID and base64 key created in step 2.
    ```yaml
    apiVersion: pipecd.dev/v1beta1
    kind: Piped
    spec:
      projectID: quickstart
      # FIXME: Replace here with your piped ID.
      pipedID: 7accd470-1786-49ee-ac09-3c4d4e31dc12
      # Base64 encoded string of the piped private key.
      # FIXME: Replace here with your piped base64 key.
      pipedKeyData: OTl4c2RqdjUxNTF2OW1sOGw5ampndXUyZjB2aGJ4dGw0bHVkamF4Mmc3a3l1enFqY20K
      # Write in a format like "host:443" because the communication is done via gRPC.
      # FIXME: Replace here with your piped address if you connect Piped to a control plane that does not run locally.
      apiAddress: localhost:8080
      repositories:
      - repoId: example
        remote: git@github.com:pipe-cd/examples.git
        branch: master
      syncInterval: 1m
      platformProviders:
      - name: example-kubernetes
        type: KUBERNETES
        config:
          # FIXME: Replace here with your kubeconfig absolute file path.
          kubeConfigPath: /path/to/.kube/config
    ```

4. Run `make run/piped CONFIG_FILE=piped-config.yaml` to start Piped agent.

### Online one-click setup for contributing

We are preparing Gitpod and Codespace to facilitate the setup process for contributing.

## Contributor License Agreement

For any code contribution, please carefully read the following documents:

- [License](https://github.com/pipe-cd/pipecd/blob/master/LICENSE)
- [Developer Certificate of Origin (DCO)](https://developercertificate.org/)

And signing off your commit with `git commit -s` (About commit sign-off please read [Github docs](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/managing-the-commit-signoff-policy-for-your-repository#about-commit-signoffs))

## What Happens Next?

The core PipeCD team will monitor the pull requests. Help us by keeping your pull requests consistent with the guidelines mentioned above.

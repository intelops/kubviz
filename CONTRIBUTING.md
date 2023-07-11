# Contribute to Ory KubViz

- [Introduction](#introduction)
- [FAQ](#faq)
- [How can I contribute?](#how-can-i-contribute)
- [Communication](#communication)
- [Contribute examples](#contribute-examples)
- [Contribute code](#contribute-code)
- [Contribute documentation](#contribute-documentation)
- [Disclosing vulnerabilities](#disclosing-vulnerabilities)
- [Code style](#code-style)
  - [Working with forks](#working-with-forks)
- [Conduct](#conduct)


## Introduction

_Please note_: We take KubViz's security and our users' trust very
seriously. If you believe you have found a security issue in KubViz, please
disclose it by contacting us at [security](https://intelops.ai/).

There are many ways in which you can contribute. The goal of this document is to
provide a high-level overview of how you can get involved in KubViz.

As a potential contributor, your changes and ideas are welcome at any hour of
the day or night, on weekdays, weekends, and holidays. Please do not ever
hesitate to ask a question or send a pull request.

If you are unsure, just ask or submit the issue or pull request anyways. You
won't be yelled at for giving it your best effort. The worst that can happen is
that you'll be politely asked to change something. We appreciate any sort of
contributions and don't want a wall of rules to get in the way of that.

That said, if you want to ensure that a pull request is likely to be merged,
talk to us! You can find out our thoughts and ensure that your contribution
won't clash with KubViz direction. A great way to do this is via
[KubViz Discussions](https://github.com/kube-tarian/kubviz/discussions) or the
[KubViz Chat](https://intelops.ai/).

## FAQ

- I am new to the community. Where can I find the
  [KubViz Community Code of Conduct?](https://github.com/intelops/kubviz/blob/main/CODE_OF_CONDUCT.md)

- I have a question. Where can I get
  [answers to questions regarding KubViz?](#communication)

- I would like to contribute but I am not sure how. Are there
  [easy ways to contribute?](#how-can-i-contribute)
  [Or good first issues?](https://github.com/intelops/kubviz/issues)

- I want to talk to other KubViz users.
  [How can I become a part of the community?](#communication)

- I would like to know what I am agreeing to when I contribute to KubViz.
  Does KubViz have
  [a Contributors License Agreement?](https://github.com/intelops/kubviz/blob/main/LICENSE.md)

- I would like updates about new versions of KubViz.
  [How are new releases announced?](https://intelops.ai/)

## How can I contribute?

If you want to start to contribute code right away, take a look at the
[list of good first issues](https://github.com/intelops/kubviz/issues).

There are many other ways you can contribute. Here are a few things you can do
to help out:

- **Give us a star.** It may not seem like much, but it really makes a
  difference. This is something that everyone can do to help out KubViz.
  Github stars help the project gain visibility and stand out.

- **Join the community.** Sometimes helping people can be as easy as listening
  to their problems and offering a different perspective. Join our Slack, have a
  look at discussions in the forum and take part in community events. More info
  on this in [Communication](#communication).

- **Answer discussions.** At all times, there are several unanswered discussions
  on GitHub. You can see an
  [overview here](https://github.com/intelops/kubviz/issues).
  If you think you know an answer or can provide some information that might
  help, please share it! Bonus: You get GitHub achievements for answered
  discussions.

- **Help with open issues.** We have a lot of open issues for KubViz and
  some of them may lack necessary information, some are duplicates of older
  issues. You can help out by guiding people through the process of filling out
  the issue template, asking for clarifying information or pointing them to
  existing issues that match their description of the problem.

- **Review documentation changes.** Most documentation just needs a review for
  proper spelling and grammar. If you think a document can be improved in any
  way, feel free to hit the `edit` button at the top of the page. More info on
  contributing to the documentation [here](#contribute-documentation).

- **Help with tests.** Pull requests may lack proper tests or test plans. These
  are needed for the change to be implemented safely.

## Communication

We use [Slack](https://intelops.ai/). You are welcome to drop in and ask
questions, discuss bugs and feature requests, talk to other users of kubviz, etc.

Check out [KubViz Discussions](https://github.com/kube-tarian/kubviz/discussions).
This is a great place for in-depth discussions and lots of code examples, logs
and similar data.

You can also join our community calls if you want to speak to the kubviz team
directly or ask some questions. You can find more info and participate in
[Slack](https://intelops.ai/) in the #community-call channel.

If you want to receive regular notifications about updates to kubviz,
consider joining the mailing list. We will _only_ send you vital information on
the projects that you are interested in.

Also, [follow us on Linkedin](https://www.linkedin.com/company/intelopsai/?originalSubdomain=in).

## Contribute examples

One of the most impactful ways to contribute is by adding examples. You can find
an overview of examples using kubviz services on the
[documentation examples page](https://github.com/intelops/kubviz). Source code for
examples can be found in most cases in the
[ory/examples](https://github.com/intelops/kubviz) repository.

_If you would like to contribute a new example, we would love to hear from you!_

Please [open an issue](https://github.com/intelops/kubviz/issues/new/choose) to
describe your example before you start working on it. We would love to provide
guidance to make for a pleasant contribution experience. Go through this
checklist to contribute an example:

1. Create a GitHub issue proposing a new example and make sure it's different
   from an existing one.
1. Fork the repo and create a feature branch off of `main` so that changes do
   not get mixed up.
1. Add a descriptive prefix to commits. This ensures a uniform commit history
   and helps structure the changelog. Please refer to this
   [list of prefixes for kubviz](https://github.com/intelops/kubviz/blob/main/.github/dependabot.yml)
   for an overview.
1. Create a `README.md` that explains how to use the example. (Use
   [the README template](https://github.com/intelops/kubviz/blob/main/README.md)).
1. Open a pull request and maintainers will review and merge your example.

## Contribute code

Unless you are fixing a known bug, we **strongly** recommend discussing it with
the core team via a GitHub issue or [in our chat](https://intelops.ai/)
before getting started to ensure your work is consistent with KubViz's
roadmap and architecture.

All contributions are made via pull requests. To make a pull request, you will
need a GitHub account; if you are unclear on this process, see GitHub's
documentation on [forking](https://help.github.com/articles/fork-a-repo) and
[pull requests](https://help.github.com/articles/using-pull-requests). Pull
requests should be targeted at the `main` branch. Before creating a pull
request, go through this checklist:

1. Create a feature branch off of `main` so that changes do not get mixed up.
1. [Rebase](http://git-scm.com/book/en/Git-Branching-Rebasing) your local
   changes against the `main` branch.
1. Run the full project test suite with the `go test -tags sqlite ./...` (or
   equivalent) command and confirm that it passes.
1. Run `make format`
1. Add a descriptive prefix to commits. This ensures a uniform commit history
   and helps structure the changelog. Please refer to this
   [list of prefixes for KubViz](https://github.com/intelops/kubviz/blob/main/.github/dependabot.yml)
   for an overview.

If a pull request is not ready to be reviewed yet
[it should be marked as a "Draft"](https://docs.github.com/en/github/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/changing-the-stage-of-a-pull-request).

Before your contributions can be reviewed you need to sign our
[Contributor License Agreement](https://github.com/intelops/kubviz/blob/main/LICENSE.md).

This agreement defines the terms under which your code is contributed to Ory.
More specifically it declares that you have the right to, and actually do, grant
us the rights to use your contribution. You can see the Apache 2.0 license under
which our projects are published
[here](https://github.com/intelops/kubviz/blob/main/LICENSE.md).

When pull requests fail the automated testing stages (for example unit or E2E
tests), authors are expected to update their pull requests to address the
failures until the tests pass.

Pull requests eligible for review

1. follow the repository's code formatting conventions;
2. include tests that prove that the change works as intended and does not add
   regressions;
3. document the changes in the code and/or the project's documentation;
4. pass the CI pipeline;
5. have signed our
   [Contributor License Agreement](https://github.com/intelops/kubviz/blob/main/LICENSE.md);
6. include a proper git commit message following the
   [Conventional Commit Specification](https://intelops.ai/).

If all of these items are checked, the pull request is ready to be reviewed and
you should change the status to "Ready for review" and
[request review from a maintainer](https://docs.github.com/en/github/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/requesting-a-pull-request-review).

Reviewers will approve the pull request once they are satisfied with the patch.

## Contribute documentation

Please provide documentation when changing, removing, or adding features. All
KubViz Documentation resides in the
[KubViz documentation repository](https://github.com/intelops/kubviz/tree/main). For further
instructions please head over to the KubViz Documentation
[README.md](https://github.com/intelops/kubviz/blob/main/README.md).

## Disclosing vulnerabilities

Please disclose vulnerabilities exclusively to
[security@intelops.ai](mailto:https://intelops.ai/). Do not use GitHub issues.

## Code style

Please run `make format` to format all source code following the KubViz standard.

### Working with forks

```bash
# First you clone the original repository
git clone git@github.com:intelops/kubviz.git

# Next you add a git remote that is your fork:
git remote add fork git@github.com:<YOUR-GITHUB-USERNAME-HERE>/intelops/kubviz.git

# Next you fetch the latest changes from origin for master:
git fetch origin
git checkout main
git pull --rebase

# Next you create a new feature branch off of master:
git checkout my-feature-branch

# Now you do your work and commit your changes:
git add -A
git commit -a -m "fix: this is the subject line" -m "This is the body line. Closes #123"

# And the last step is pushing this to your fork
git push -u fork my-feature-branch
```

Now go to the project's GitHub Pull Request page and click "New pull request"

## Conduct

Whether you are a regular contributor or a newcomer, we care about making this
community a safe place for you and we've got your back.

[KubViz Community Code of Conduct](https://github.com/intelops/kubviz/blob/main/CODE_OF_CONDUCT.md)

We welcome discussion about creating a welcoming, safe, and productive
environment for the community. If you have any questions, feedback, or concerns
[please let us know](https://intelops.ai/).
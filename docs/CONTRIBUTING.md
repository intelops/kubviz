# Contribution Guidelines

- [Introduction](#introduction)
- [Reporting Issues](#reporting-issues)
- [Feature Requests](#feature-requests)
- [Contribute code](#contribute-code)

## Introduction

The goal of this document is to provide an overview of how you can get involved in KubViz.

As a potential contributor, your changes and ideas are welcome at any hour of the day or night, on weekdays, weekends, and holidays. Please do not ever hesitate to ask a question or send a pull request.

If you are unsure, just ask or submit the issue or pull request anyways. You won't be yelled at for giving it your best effort.

## Reporting Issues

If you find a bug while working with the KubViz, please [open an issue on GitHub](https://github.com/intelops/kubviz/issues) and let us know what went wrong. We will try to fix it as quickly as we can.

## Feature Requests

You are more than welcome to open issues in this project to [suggest new features](https://github.com/intelops/kubviz/issues).


## Contribute code

### Dependencies

This project is written in Golang.

You need 

<a href="https://go.dev/doc/install" target="_blank">`Go 1.16+`</a>

<a href="https://docs.docker.com/engine/install/" target="_blank">`Docker`</a>

<a href="https://docs.docker.com/compose/install/standalone/" target="_blank">`Docker Compose`</a>

All contributions are made via pull requests. To make a pull request, you will need a GitHub account; if you are unclear on this process, see GitHub's documentation on [forking](https://help.github.com/articles/fork-a-repo) and [pull requests](https://help.github.com/articles/using-pull-requests). Pull requests should be targeted at the main branch. Before creating a pull request, go through this checklist:

Clone Kubviz and run it in Docker

```bash
git clone https://github.com/intelops/kubviz.git

cd kubviz
```
1. Create a feature branch of `main` so that changes do not get mixed up.

2. Rebase your local changes against the `main` branch.

3. Run the full project test it.

4. Add a descriptive prefix to commits. This ensures a uniform commit history and helps structure the changelog.

If a pull request is not ready to be reviewed yet it should be marked as a "Draft".

### Working with forks

```bash
# First you clone the original repository
git clone git@github.com:intelops/kubviz.git

# Next you add a git remote that is your fork:
git remote add fork git@github.com:<YOUR-GITHUB-USERNAME-HERE>/kubviz.git

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

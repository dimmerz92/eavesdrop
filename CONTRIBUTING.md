# Contributing
Thank you for your interest in contributing! Because I am a team of one, I have a couple contribution guidelines that make it easier for me to triage any incoming suggestions.

## Issues
- Issues are the best place to propose a new feature.
- Search the issues before proposing a feature to see if it is already under discussion. Referencing existing issues is a good way to increase the priority of your own.
- I don't have an issue template yet, but the more detailed your description of the issue, the more quickly I'll be able to evaluate it.
- See an issue that you also have? Give it a reaction (and comment, if you have something to add).

## Creating a Development Environment
### Pre-requisites
To develop for Eavesdrop, you'll need the following:

- Go 1.24.1 or later

### Installing Packages
To install Eavesdrops's required packages, run the following command:

```sh
go mod tidy
```

## Pull Requests
### Technical Requirements
1. All PRs must be made against the `dev` branch, except documentation PRs which can be made against `master`.
1. Please avoid sending the `tmp` directory or any non gotest test files along with your PR.
1. Please include test cases.
1. I will squash all PRs, so you're welcome to submit with as many commits as you like; they will be evaluated as a single, standalone change.

### Review Guidelines
1. Open PRs represent issues that I am actively working on merging. If I think a proposal needs more discussion, or that the existing code would require a lot of back-and-forth to merge, I might close it and suggest you make an issue.
1. Smaller PRs are easier and quicker to review. If I feel that the scope of your changes is too large, I will close the PR and try to suggest ways that the change could be broken down.
1. Please do not PR new features unless you have already made an issue proposing the feature, and it has been accepted. This helps me triage the features I can support before you put a lot of work into them.
1. Correspondingly, it is fine to directly PR bugfixes for behavior that Eavesdrop already guarantees, but please check if there's an issue first, and if you're not sure whether this *is* a bug, make an issue where I can hash it out..
1. Refactors that do not make functional changes will be automatically closed, unless explicitly solicited. Imagine someone came into your house unannounced, rearranged a bunch of furniture, and left. wft!
1. Typo fixes in the documentation (not the code comments) are welcome, but formatting or debatable grammar changes will be automatically closed.

## Misc
1. If you think I closed something incorrectly, feel free to (politely) tell me why! I'm human (as far as you're aware...) and make mistakes.

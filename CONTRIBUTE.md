# How to Contribute

1. **Contributor**: An issue has to be created with a problem description and possible solutions using the github template. The better the problem and solution is described, the easier for the **Curators and Others** to understand it and the faster the process. In case of a bug, steps to reproduce will help a lot.
2. **Curators and Others**: The curators will engage in a discussion about the problem and the possible solution. Others can join the discussion to bring other solutions and insights at any point.
3. **Curators**: After the discussion mentioned above, it will be determined if the proposed solution will be implemented or not. Appropriate tags will be applied to the issue.
4. **Contributor**: The contributor will work as follows:
    - assign the issue to himself
    - create a fork
    - git clone the fork on your machine
    - enable [signing your work](SIGNYOURWORK.md)
    - create a PR. use WIP to mark unfinished work e.g. __WIP: Fixing a bug (fixes #1)__
    - after development has finished remove the WIP if applied
5. **Curators**: The curators will conduct a full code review
6. **Curators**: After at least 2 curators have approved the PR, it will be merged to master
7. **Curators**: A release will follow after that at some point

PR's should have the following requirements:

- Tests are required (where applicable, terms may vary)
  - Unit
  - Component
  - Integration
- High code coverage
- Coding style (go fmt)
- Linting we use [golangci-lint](https://github.com/golangci/golangci-lint)
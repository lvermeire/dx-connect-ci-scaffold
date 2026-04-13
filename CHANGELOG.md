# Changelog

## [1.0.1](https://github.com/lvermeire/dx-connect-ci-scaffold/compare/v1.0.0...v1.0.1) (2026-04-13)


### Bug Fixes

* fix ci output tests ([#21](https://github.com/lvermeire/dx-connect-ci-scaffold/issues/21)) ([7587e2a](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/7587e2a7c665f4470edbf5dda47284b062bcdd6e))

## [1.0.0](https://github.com/lvermeire/dx-connect-ci-scaffold/compare/v0.2.0...v1.0.0) (2026-04-09)


### ⚠ BREAKING CHANGES

* promote to stable v1.0.0

### Features

* promote to stable v1.0.0 ([66a7c2d](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/66a7c2da088ecf01a8c087b848281e3ee72b9007))

## [0.2.0](https://github.com/lvermeire/dx-connect-ci-scaffold/compare/v0.1.0...v0.2.0) (2026-04-09)


### Features

* add web Dockerfile, nginx config, and docker-compose ([2b8d254](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/2b8d25464921f5455725a7a5e8c06c34a01160cd))
* **api:** add health and items HTTP handlers ([7a96568](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/7a96568fab98bc72e3ebb3998f25a82f1ee31c09))
* **api:** add in-memory ItemStore ([e2d9bdd](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/e2d9bdd4f81b4ed43db0a96453d18e245cab2783))
* **api:** add multi-stage Dockerfile with distroless runtime ([c03fb66](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/c03fb66ca72dc6f8f1295ec1319f9c631c6055db))
* **api:** wire router and server entrypoint ([6a8c281](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/6a8c2810f954713f3208190bc9f419c0c055cbbb))
* **web:** add ItemList component with tests ([c2bdc5c](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/c2bdc5c606d3332224051d9b233404a2d2c12895))
* **web:** scaffold Vue 3 app with Vite and ESLint ([e03eb49](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/e03eb49c3c52175a03da7c45d0b3e30ec1cdcb1f))


### Bug Fixes

* **api:** bump Dockerfile base image to golang:1.26.2-alpine ([3c2513e](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/3c2513e923d005bf8a2c131996b13f211b156891))
* **api:** document goimports unavailability in golangci-lint v2 ([df35973](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/df359731e66f25a0ec89d42525783786a5f0b290))
* **api:** pin go directive to 1.23 per spec ([d303ff3](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/d303ff3da8c8266078fbfc352d2649b74e49ddc2))
* **api:** split error branches and use correct Content-Type in CreateItem ([75783a5](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/75783a549c9dbcf12f18f5b088feec41a101ad21))
* **ci:** go-task/setup-task, golangci-lint v2.11.4, Go 1.26.2 ([22b150d](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/22b150d478c3598b7ddebd5bc1f42313e931ad6a))
* **ci:** use Taskfile tasks and fix broken action versions ([5d1a05e](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/5d1a05eadd4f000cd20d60059cb913a8ea88069e))
* update Go module path to match GitHub owner (lvermeire) ([1db627e](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/1db627ed97a453c5ff2b42d82765aa894e396aa3))


### Documentation

* add CI/CD scaffold design spec ([70c285c](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/70c285ca23a0254dfeaaa270d736705ef63292a0))
* add Plan A — Foundation implementation plan ([2e2ce0d](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/2e2ce0d9db62d18f1ae42b4c3a80ef9c343c09c3))
* add Plan B — CI/CD pipelines implementation plan ([32ad139](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/32ad1396aef36bab2d7576da0bb2900fc597badc))
* fix module path in CLAUDE.md ([41b9241](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/41b9241d282c852852fa4315a3d20f7df975ce6d))
* update CLAUDE.md for Plan B / cicd-pipelines worktree ([9d307f6](https://github.com/lvermeire/dx-connect-ci-scaffold/commit/9d307f63e57ea3c821bb7953ed6cb0ff829f18e4))

## Changelog

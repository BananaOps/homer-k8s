# Changelog

## [0.6.2](https://github.com/BananaOps/homer-k8s/compare/v0.6.1...v0.6.2) (2026-11-04)


### 🐛 Bug Fixes

* refresh dependencies, fix YAML keys and EC2 discovery, repair CI and release automation ([#180](https://github.com/BananaOps/homer-k8s/issues/180)) ([8c619a7](https://github.com/BananaOps/homer-k8s/commit/8c619a729bc954800b24bb135ec15b68787586ca))


### 📚 Documentation

* refresh README ([#184](https://github.com/BananaOps/homer-k8s/issues/184)) ([9935b68](https://github.com/BananaOps/homer-k8s/commit/9935b6814e34a3d7a0c9ef2a82e769740ea37314))


### 👷 CI/CD

* bump charts to 0.6.1 ([56a0b98](https://github.com/BananaOps/homer-k8s/commit/56a0b98d66cb14f2ee4eb3cfa8194710c1b3723f))
* pin GitHub Actions to their latest release by commit SHA ([#183](https://github.com/BananaOps/homer-k8s/issues/183)) ([3c4d86b](https://github.com/BananaOps/homer-k8s/commit/3c4d86bc37ffae68c3941332996a2eec3dc9d98e))

## [0.6.1](https://github.com/BananaOps/homer-k8s/compare/v0.6.0...v0.6.1) (2026-11-04)


### Bug Fixes

* **chart:** grant homerpages to the controller ClusterRole ([#178](https://github.com/BananaOps/homer-k8s/issues/178)) ([c2c060c](https://github.com/BananaOps/homer-k8s/commit/c2c060cb5e9632fe95d1a2937c106483fb83049e))

## [0.6.0](https://github.com/BananaOps/homer-k8s/compare/v0.5.1...v0.6.0) (2026-11-04)


### Features

* add new crd for page ([#170](https://github.com/BananaOps/homer-k8s/issues/170)) ([01a4dd2](https://github.com/BananaOps/homer-k8s/commit/01a4dd2a7f4aefb8a7da7845258559ae40f98fcd))
* more fields in config items ([#165](https://github.com/BananaOps/homer-k8s/issues/165)) ([0fbef5c](https://github.com/BananaOps/homer-k8s/commit/0fbef5ca6cb5b97e1aade244e8a9e2d34b13c1f2))
* pass through extra Homer keys and add customCss to the chart ([#177](https://github.com/BananaOps/homer-k8s/issues/177)) ([549ad24](https://github.com/BananaOps/homer-k8s/commit/549ad24d6230ae56bed99e541949cbfea6892f2e))


### Bug Fixes

* security issue G703 (CWE-22): Path traversal via taint analysis ([392062a](https://github.com/BananaOps/homer-k8s/commit/392062a93f3a03c5668f70fadb7fe93ceb7d9523))

## [0.5.1](https://github.com/BananaOps/homer-k8s/compare/v0.5.0...v0.5.1) (2025-05-06)


### Bug Fixes

* correct path ko-build ([#120](https://github.com/BananaOps/homer-k8s/issues/120)) ([e01e8f5](https://github.com/BananaOps/homer-k8s/commit/e01e8f58d459dd70de1cbd49c80625dfbc635117))

## [0.5.0](https://github.com/BananaOps/homer-k8s/compare/v0.4.0...v0.5.0) (2025-05-06)


### Features

* bump golang and run mod tidy ([#118](https://github.com/BananaOps/homer-k8s/issues/118)) ([e4966fd](https://github.com/BananaOps/homer-k8s/commit/e4966fdc340bf11cfafe394a5225d2c5c629f0b2))

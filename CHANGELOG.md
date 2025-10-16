# Changelog

## [0.1.5](https://github.com/kbrockhoff/terraform-provider-context/compare/v0.1.4...v0.1.5) (2025-10-16)


### Bug Fixes

* add --provider-name flag to tfplugindocs commands ([abaa907](https://github.com/kbrockhoff/terraform-provider-context/commit/abaa907146cbe9b0772390c79e42d4d3f2e4c6a7))
* add custom tfplugindocs template to fix docs generation ([e10aa6d](https://github.com/kbrockhoff/terraform-provider-context/commit/e10aa6d0fd0c9ed36681d7331e90a17cbe7ef4a5))
* add security-events permission and fix codecov action parameter ([a86666c](https://github.com/kbrockhoff/terraform-provider-context/commit/a86666ceb296fa986b22df5bf934703e4e81cdb7))
* resolve tfplugindocs path expansion and skip data-sources example directory ([6f8ddcc](https://github.com/kbrockhoff/terraform-provider-context/commit/6f8ddcc07037f62f21d9579239d9dd971a26d535))
* update security scanner action and add provider-name flag to docs generation ([e126232](https://github.com/kbrockhoff/terraform-provider-context/commit/e1262324b7aaf2e8acbaa9a28e1dd0221eaf63f4))

## [0.1.4](https://github.com/kbrockhoff/terraform-provider-context/compare/v0.1.3...v0.1.4) (2025-10-11)


### Bug Fixes

* use source file glob in release.extra_files ([fbdd2c7](https://github.com/kbrockhoff/terraform-provider-context/commit/fbdd2c70cedb907fb05768f7b679a66c17f6df94))

## [0.1.3](https://github.com/kbrockhoff/terraform-provider-context/compare/v0.1.2...v0.1.3) (2025-10-11)


### Bug Fixes

* add manifest.json to release extra_files ([24d2c9b](https://github.com/kbrockhoff/terraform-provider-context/commit/24d2c9bd7df7414174c0dcc4f489ee7d43625400))
* reference renamed manifest from dist directory in release ([070a3fc](https://github.com/kbrockhoff/terraform-provider-context/commit/070a3fc1f31bde569d3af4d588264e91298d70c2))

## [0.1.2](https://github.com/kbrockhoff/terraform-provider-context/compare/v0.1.1...v0.1.2) (2025-10-11)


### Bug Fixes

* update goreleaser config to v2 standards ([5546b3c](https://github.com/kbrockhoff/terraform-provider-context/commit/5546b3cabbf5057bd79164f87ca0dda34966b54c))

## [0.1.1](https://github.com/kbrockhoff/terraform-provider-context/compare/v0.1.0...v0.1.1) (2025-10-11)


### Bug Fixes

* integrate release-please with goreleaser workflow ([90b9e22](https://github.com/kbrockhoff/terraform-provider-context/commit/90b9e224d46a0742e3c32dd0968668e58994fc1b))

## 0.1.0 (2025-09-27)


### Features

* implement core Terraform provider functionality ([2c680f1](https://github.com/kbrockhoff/terraform-provider-context/commit/2c680f1d959ec2f3a88df77a1c2afcc702ac613e))
* make TF_CLI_CONFIG_FILE work both locally and in GitHub Actions ([1aaa421](https://github.com/kbrockhoff/terraform-provider-context/commit/1aaa421d6f15a249c6b2e522b70edaa4a56ba73a))
* update tag names to match terraform-external-context and clean up environment tags ([0cdaacd](https://github.com/kbrockhoff/terraform-provider-context/commit/0cdaacdecc858dd73f390b1231b0e5fa0457e86d))


### Bug Fixes

* configure goreleaser and release-please for Terraform Registry publishing ([86fe222](https://github.com/kbrockhoff/terraform-provider-context/commit/86fe2222d515a62e93c240c31ffb2c511222732f))
* resolve gemini identified bugs ([c888d2e](https://github.com/kbrockhoff/terraform-provider-context/commit/c888d2e2e000806117dbd15974393a73e3326481))
* resolve github copilot comments ([7a041be](https://github.com/kbrockhoff/terraform-provider-context/commit/7a041bec454f11a6721806b38a7e19e399df5520))
* update golangci-lint configuration for v2.4.0 compatibility ([8aaa904](https://github.com/kbrockhoff/terraform-provider-context/commit/8aaa9043f4f3baf26dbd8f5a7712d7d768bd5e5e))


### Performance Improvements

* precompile regular expressions in cloud provider implementations ([3fe2cc1](https://github.com/kbrockhoff/terraform-provider-context/commit/3fe2cc176c022ec7f9172d52fe55b5f91248e89e))

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

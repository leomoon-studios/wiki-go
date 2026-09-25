# Changelog

## [v1.9.2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.9.2) (2026-09-21)

- ci: link authors in release history ([3a4c07d](https://github.com/leomoon-studios/wiki-go/commit/3a4c07d842e31144ad1aef58ba54dfd43274da2f))
- security: canonicalize document request paths ([b62aa4b](https://github.com/leomoon-studios/wiki-go/commit/b62aa4b759da2ec9d86c42a1f26d4b15af524ebe))
- security: enforce document access rules on editor APIs ([78843c0](https://github.com/leomoon-studios/wiki-go/commit/78843c0454ee345bd649df1fc6dec32cd0a8d857))
- test: cover document access containment ([e86940b](https://github.com/leomoon-studios/wiki-go/commit/e86940bba408fc116cb9a78bb1e57c489b701366))
- security: constrain attachment rename paths ([81cd12e](https://github.com/leomoon-studios/wiki-go/commit/81cd12e248314a52605f0f39718c8a76272eac29))
- security: enforce safe attachment renames ([8adbcb7](https://github.com/leomoon-studios/wiki-go/commit/8adbcb70647e22654b08116a4d8b12cf76cb2413))
- security: resolve document paths safely ([92fcec4](https://github.com/leomoon-studios/wiki-go/commit/92fcec4cabd3b4b057664e6f160949cccaa90aad))
- security: contain document editor paths ([94f7e73](https://github.com/leomoon-studios/wiki-go/commit/94f7e73ef8df7c58091d72d77a5e26e9fdad98b8))
- test: cover editor path containment ([20b0d50](https://github.com/leomoon-studios/wiki-go/commit/20b0d500d256e44fb042820bc008bdb8cdbf8e64))
- security: validate metadata request destinations ([25498eb](https://github.com/leomoon-studios/wiki-go/commit/25498ebf5c012b870967e38eec2a1e5133fa1a83))
- security: enforce metadata request policy ([c587c70](https://github.com/leomoon-studios/wiki-go/commit/c587c70852fb94b1a75a12f631b029c4d6d9434e))
- test: cover metadata endpoint security ([1c7f473](https://github.com/leomoon-studios/wiki-go/commit/1c7f473196bcbe040790942171a2c5cd72260adb))
- security: unify attachment path boundaries ([bba3e64](https://github.com/leomoon-studios/wiki-go/commit/bba3e64ea09fe55114c4065aa9de2cd143d6267f))
- security: contain comment file paths ([98c0adc](https://github.com/leomoon-studios/wiki-go/commit/98c0adccb8a8b30fad45a5654161774a5fa18e77))
- security: enforce comment storage boundary ([224ef2a](https://github.com/leomoon-studios/wiki-go/commit/224ef2a11facddef6b3a3c1f0a69156d1e574a33))
- test: cover comment deletion traversal ([bbae067](https://github.com/leomoon-studios/wiki-go/commit/bbae067d11b0250a025e37af282d32676003a518))
- security: neutralize log control characters ([2ed3666](https://github.com/leomoon-studios/wiki-go/commit/2ed36665be0880f7a55d2bf4ffa659b56a51ba8e))
- test: cover login log injection ([f93c8f9](https://github.com/leomoon-studios/wiki-go/commit/f93c8f9887817e104c18ba6a3dd9d2fc8ca52dd0))
- security: audit sensitive logging ([af27b23](https://github.com/leomoon-studios/wiki-go/commit/af27b2353bfeda0c77c7cb5a6254ed1223451012))
- test: cover page backslash traversal ([177ee6d](https://github.com/leomoon-studios/wiki-go/commit/177ee6d9626fa3ad23e0eb2890fc8f758d9af1ee))

## [v1.9.1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.9.1) (2026-09-13)

- Restore external Markdown images ([18d7fe3](https://github.com/leomoon-studios/wiki-go/commit/18d7fe32022287aad3244ac82fc3637e4379dbd1))
- fix: render Mermaid inside details blocks ([1f403d0](https://github.com/leomoon-studios/wiki-go/commit/1f403d0dc1bb6dccf2054fcb17255132b98c1d64))

## [v1.9.0](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.9.0) (2026-09-05)

- Improve side panel expand/collapse behavior (#181) ([2677ab6](https://github.com/leomoon-studios/wiki-go/commit/2677ab6b3cf8a1c67f2c78e8f7356498be702ab0))
- PR 181 follow up ([c8d7538](https://github.com/leomoon-studios/wiki-go/commit/c8d75384a8cf333ff98b2379abbf4dbe63cdef73))
- Updates rest-api-examples.http file ([6e57ef2](https://github.com/leomoon-studios/wiki-go/commit/6e57ef2c587be9a8f08a7699a43cd9c16285625f))
- Security (markdown): render comments without unsafe HTML ([d3b0132](https://github.com/leomoon-studios/wiki-go/commit/d3b0132fa6391f1f9f984f346fd087de9b245d27))
- Security (markdown): add trusted nodes for Wiki-Go extensions ([8eb0961](https://github.com/leomoon-studios/wiki-go/commit/8eb0961bc9ce5a05fc426f91f454b993057019fe))
- Security (markdown): safely render direction and Mermaid blocks ([71ac014](https://github.com/leomoon-studios/wiki-go/commit/71ac014d3d100b69b3f7e25c92965b68bd811049))
- Security (markdown): validate and safely render media extensions ([73d415a](https://github.com/leomoon-studios/wiki-go/commit/73d415ac1b7b48b4c5e3852cc0ff318a7f193c45))
- Security (markdown): safely render formatting and navigation extensions ([1a5faaf](https://github.com/leomoon-studios/wiki-go/commit/1a5faaf2e5a9479e5bfa10166f600b17e025559f))
- Security (markdown): harden kanban and links rendering ([e707a62](https://github.com/leomoon-studios/wiki-go/commit/e707a6204d72ba6bc0fea0b87dc586ccc2f86d7d))
- Fixes github style admonition regression ([a9b71d5](https://github.com/leomoon-studios/wiki-go/commit/a9b71d54ce9b8c552775c5bd34f9687fce16e4f3))
- Security (markdown): disable raw HTML in document renderers ([f893388](https://github.com/leomoon-studios/wiki-go/commit/f893388f5a0392169f3c889f65fb822f8c066c30))
- Fixes shortcode regression ([713edfa](https://github.com/leomoon-studios/wiki-go/commit/713edfa9b6211de4b5d9719eabb8f3a698b48fa0))
- Security (html): restrict trusted HTML conversion points ([b7a790a](https://github.com/leomoon-studios/wiki-go/commit/b7a790af344106d371987d0c243166e4761a7a83))
- Security (svg): replace regex sanitization with an XML allowlist ([f16f2ab](https://github.com/leomoon-studios/wiki-go/commit/f16f2abc7ba8133ce6d3bd7299268ec2a2723649))
- Security (csp): enforce a restrictive page security policy ([c8db525](https://github.com/leomoon-studios/wiki-go/commit/c8db525cc9515440eb95d48683266dca43fa45a6))
- Fixes regression on favicons not loading in link pages ([f698857](https://github.com/leomoon-studios/wiki-go/commit/f6988571bc12477fd8d0b2cbf55cdbab545a9d98))
- Security: harden sessions and trusted proxy handling ([2449c3b](https://github.com/leomoon-studios/wiki-go/commit/2449c3b7468418c7eae7875815ffeec761158b02))
- Fixes bad links indicator regression ([e1f1228](https://github.com/leomoon-studios/wiki-go/commit/e1f1228a261bbcb8bd2839a5de48011b1b99fe44))
- Updates SECURITY.md ([6cb2ee5](https://github.com/leomoon-studios/wiki-go/commit/6cb2ee55dcc3aa889ac8f177c43ae05b61e60d7d))
- Prepare for the chapter links feature ([5c5e605](https://github.com/leomoon-studios/wiki-go/commit/5c5e6050bc2edd48ea9e0187262228bde1758c0f))
- Implement chapter links ([55d38d2](https://github.com/leomoon-studios/wiki-go/commit/55d38d281c768eec2b21ed70a9d3ea69d07614b5))
- Add show/retract button to chapter links side panel ([55b7e6d](https://github.com/leomoon-studios/wiki-go/commit/55b7e6d0ae7bf5a5f81c703680b1c3bd73ffe76d))
- Fix rearrangement of side panel contents ([7e8054c](https://github.com/leomoon-studios/wiki-go/commit/7e8054c42c03b5d5b26bc863c5e052ba3e6f94ce))
- Make page title bold in the side panel ([919bd33](https://github.com/leomoon-studios/wiki-go/commit/919bd331f6215003edac13c7619a74f910802cf0))
- Refactor code block mark detection ([ec3445d](https://github.com/leomoon-studios/wiki-go/commit/ec3445dddd348e1e5e0d27cf730f3e558d740b7b))
- Use retracted instead of collapsed for chapter links panel ([915aca9](https://github.com/leomoon-studios/wiki-go/commit/915aca92078e855ee13b634de9f57b86a3382048))
- Fix side panel on mobile ([def3ed7](https://github.com/leomoon-studios/wiki-go/commit/def3ed719a4328150f410d2bb6d0ff67730cabf4))
- Make the side panel tab movable on mobile ([02c856d](https://github.com/leomoon-studios/wiki-go/commit/02c856df7ce0dc494f214185aca2517fece30d26))
- Merge PR #186 chapter links into dev ([bd0b128](https://github.com/leomoon-studios/wiki-go/commit/bd0b1284bded01d7a095ef5060c032168c7c3f16))
- Exposes trusted chapter heading metadata ([42d5033](https://github.com/leomoon-studios/wiki-go/commit/42d503390b4e0664be6659e4f567465a807a4801))
- Safely render chapter links from typed headings ([e458284](https://github.com/leomoon-studios/wiki-go/commit/e458284e44ded06f2cbf679183e6c9af08f40e98))
- Makes chapter links state CSP compatible ([19c2872](https://github.com/leomoon-studios/wiki-go/commit/19c2872b05581b5f92d808b90fade7da88c4fba8))
- Preserve inline TOC behavior with chapter links and refine styling ([4d317a5](https://github.com/leomoon-studios/wiki-go/commit/4d317a5876b1a7ef468bce9f0f462f3319dd5b3d))
- Localizes and hardens the panel controls ([80aaf9c](https://github.com/leomoon-studios/wiki-go/commit/80aaf9cd56afb7ea3ddd5be9c524b65e434cec1f))
- Add permanent security and regression coverage ([f6c574e](https://github.com/leomoon-studios/wiki-go/commit/f6c574e1289c56d473389c3ae2192961cd8a5556))
- Merge secure chapter links integration into dev ([ef2d3e5](https://github.com/leomoon-studios/wiki-go/commit/ef2d3e51177d1590a29c7754205645e1f940e8ce))
- Adds close chapter-links on mobile on outside tap ([d41e277](https://github.com/leomoon-studios/wiki-go/commit/d41e2779f017d98c8a5becc027f7b9dacde8a88b))
- Merge pull request #187 from leomoon-studios/dev ([4fd20d9](https://github.com/leomoon-studios/wiki-go/commit/4fd20d91be8bf48f0442a6d663b15e01b469feca))

## [v1.8.13](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.13) (2026-08-20)

- Fixes root_dir when absolute path is used ([0882888](https://github.com/leomoon-studios/wiki-go/commit/0882888bbbf17c2b30abbb854e5c3b4f2611edf4))

## [v1.8.12](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.12) (2026-07-14)

- SECURITY: Updates to go 1.26.5 to update crypto package security issue ([842049d](https://github.com/leomoon-studios/wiki-go/commit/842049d0cf097f0cd6e4f7b7c66888e56078d9b0))

## [v1.8.11](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.11) (2026-06-28)

- Fixes timezone UI papercut ([60e80de](https://github.com/leomoon-studios/wiki-go/commit/60e80de6f731828efafad337fe9b6ec03c8ba78c))
- Adds proper logging #177 ([07ff417](https://github.com/leomoon-studios/wiki-go/commit/07ff4170393e4698486b9859d3ad9276221c3320))
- SECURITY: fixes directory listing leaking access-controlled child document names ([57c8efe](https://github.com/leomoon-studios/wiki-go/commit/57c8efe083d7666526056d482c52ab600411fafc))
- Adds dual-pane scroll sync #176 ([52244fa](https://github.com/leomoon-studios/wiki-go/commit/52244faf27f97a5fddbbad7f9f8d2f3d2367c841))
- Add [[wikilink]] support (#179) ([9707f25](https://github.com/leomoon-studios/wiki-go/commit/9707f25a33f7c88a4d06ffd327fb8f218822300a))
- Adds early exit for wikilink processor ([c59e174](https://github.com/leomoon-studios/wiki-go/commit/c59e1740f9fc9fb2d8417fb19fa42b23eabfd8bc))
- Adds slug index cache with mtime invalidation for wikilinks ([1d02d10](https://github.com/leomoon-studios/wiki-go/commit/1d02d10ae5a9c381ff9b3f81f66f0a532a4444f2))
- Removes redundant regex match for wikilinks and test improvements ([4e67805](https://github.com/leomoon-studios/wiki-go/commit/4e67805dd8ce6f6d9b7a5d70cb94b9142afd098a))
- Fixes scroll position reset when opening/closing sidebar on mobile ([400667b](https://github.com/leomoon-studios/wiki-go/commit/400667b4f0f9ade3de71961c83294da288f3da2f))
- Updates workflows to fix build issues ([c4f7cd5](https://github.com/leomoon-studios/wiki-go/commit/c4f7cd5cba17c1d2a2a291a7cb84f95038acb611))

## [v1.8.10](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.10) (2026-05-26)

- Adds paypal to FUNDING.yml ([3528e04](https://github.com/leomoon-studios/wiki-go/commit/3528e04168c6212bc974963616d4f334775d4606))
- Adds tab scrolling to settings #173 ([64d9164](https://github.com/leomoon-studios/wiki-go/commit/64d9164c3866d5ef7611dbd54365d8843d5600c8))
- Fixes image title on image links not working #174 ([8ddef91](https://github.com/leomoon-studios/wiki-go/commit/8ddef91cd3faf294049c1c41572e85be4a847b7d))
- Adds one more missing translation key. Translation done by ChatGPT #172 ([6daeffe](https://github.com/leomoon-studios/wiki-go/commit/6daeffeedfce9ac59fcd9457e6917bcf2b357a76))

## [v1.8.9](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.9) (2026-05-13)

- Links to Notion import tool doc in the demo site docs ([c071bfc](https://github.com/leomoon-studios/wiki-go/commit/c071bfc37961865703c1ca85ff5d29d71cbc87dd))
- Fixes the wrong link back to github for Notion import tool ([8c479d8](https://github.com/leomoon-studios/wiki-go/commit/8c479d854aeae7576e5691df05a3aa51764fe0aa))
- Document missing keeploggedin boolean (#165) ([e18646c](https://github.com/leomoon-studios/wiki-go/commit/e18646c8febc2f7bd880f64b70de1797e95af12d))
- Reduces docker image size. Thanks @Izumiko #163 ([0b9af9c](https://github.com/leomoon-studios/wiki-go/commit/0b9af9cb929174ccdde2145c098c48c87469ee79))
- Allow all users to change their own password (#164) ([600e977](https://github.com/leomoon-studios/wiki-go/commit/600e977bfef7c7b31c8319e4fd568f5f70f95600))
- fix(goldext): format :::stats recent::: edit times in configured wiki timezone (#169) ([f32b484](https://github.com/leomoon-studios/wiki-go/commit/f32b4849bd68717295a8d12210bf891d5194a573))
- Optimizations for user password change ([ca29f37](https://github.com/leomoon-studios/wiki-go/commit/ca29f3788bb5e290fbed00a21c75bf36420ca559))
- Optimizations for stats recent plugin ([7b5595a](https://github.com/leomoon-studios/wiki-go/commit/7b5595a7165c553a6eaa7c6481369e515db5b8dc))

## [v1.8.8](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.8) (2026-04-25)

- Changes breadcrumbs "Home" to an icon ([b8bd2e5](https://github.com/leomoon-studios/wiki-go/commit/b8bd2e53c2ddd17b439a4599284ba631b69528ff))
- Adds more missing translations #155 ([fc41178](https://github.com/leomoon-studios/wiki-go/commit/fc41178ec3cde9b46ecf54e8044313f49546f77e))
- docs: add GROWTH.md community visibility playbook ([67b164f](https://github.com/leomoon-studios/wiki-go/commit/67b164f492be05841ed7960acdea7704a3de1ebb))
- Merge pull request #158 from Gingiris/master ([65c3c93](https://github.com/leomoon-studios/wiki-go/commit/65c3c9344f2a8725fb3819a8fdb724af2050d599))
- notion to wikigo (#160) ([e15acbf](https://github.com/leomoon-studios/wiki-go/commit/e15acbfa8edda55516a8cc42245402d64dbff031))
- Adding setting Wiki.AlwaysOpenChildrenInSidebar (#161) ([1fb3455](https://github.com/leomoon-studios/wiki-go/commit/1fb34552ff4d84bc418c5c3f03ad7c41a79b0547))
- Some optimizations for AlwaysOpenChildrenInSidebar setting ([4db6d64](https://github.com/leomoon-studios/wiki-go/commit/4db6d647de8418c492be142bc865a954242bd737))
- Adds more missing translation keys with translations done by Gemini #157 ([3b945ed](https://github.com/leomoon-studios/wiki-go/commit/3b945ed59ae0dd4467394fd0cc05de88670ffa0c))
- Updates go golang v1.26.2 and modules ([ca696a4](https://github.com/leomoon-studios/wiki-go/commit/ca696a4e83a0029c042166c79d7b3dfa5cd42e09))
- Links to Notion import tool doc in main README ([3e7fa7a](https://github.com/leomoon-studios/wiki-go/commit/3e7fa7a26b05203ca4a9cc0c2883966bac4b67c9))

## [v1.8.7](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.7) (2026-03-13)

- SECURITY: locks down search based on permission ([e1005d0](https://github.com/leomoon-studios/wiki-go/commit/e1005d09f7cb577197ca8899036c0a00281bb344))
- Adds logging in state to login button with translations done by ChatGPT ([cbad250](https://github.com/leomoon-studios/wiki-go/commit/cbad25034dc22fd714fe9d76cac47b8f6f46982f))
- Some translation fixes #151 ([f640aa6](https://github.com/leomoon-studios/wiki-go/commit/f640aa6f0082f8030f5022159165442f71b52fae))

## [v1.8.6](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.6) (2026-03-08)

- Add split edit|preview mode in editor ([051df00](https://github.com/leomoon-studios/wiki-go/commit/051df00cc1a2913989a66917ce3844adc8d7464b))
- Merge pull request #147 from trapexit/split ([af0abd2](https://github.com/leomoon-studios/wiki-go/commit/af0abd2eac56e4f4418181708bbd0725093ae270))
- Fix: adds box-sizing border-box to prevent editor preview padding overflow ([6f766a2](https://github.com/leomoon-studios/wiki-go/commit/6f766a223af2afe4899223f6738fc3034b5eb4b2))
- Adds better strikethrough than the default CodeMirror ([4514545](https://github.com/leomoon-studios/wiki-go/commit/451454553f2f079e7520b4929f056250e40d59e0))
- adds ctrl+shift+s shortcut for split view and fixes some issues with keyboard shortcut handling ([31452fc](https://github.com/leomoon-studios/wiki-go/commit/31452fc7f5d41bfd173f437cccb7678349925ef3))
- Updates go version and dependencies ([5965ab2](https://github.com/leomoon-studios/wiki-go/commit/5965ab232117d117c523c00b520c31d3915a29f7))
- Fixes valid link detection with achors #153 ([d7e24af](https://github.com/leomoon-studios/wiki-go/commit/d7e24af19920505725a6573439b6ae135c29e440))

## [v1.8.5](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.5) (2026-02-10)

- feat: Add custom.js for user customization ([b2737af](https://github.com/leomoon-studios/wiki-go/commit/b2737afe4bbf1bf64af64053967e812af36744ae))
- Merge pull request #141 from fullfox/feature/custom-js ([ffba9b0](https://github.com/leomoon-studios/wiki-go/commit/ffba9b033d79aec3bdbff0029bf4d5777f9d90a6))
- Adds Open Graph meta tags to head section in base template ([db446f7](https://github.com/leomoon-studios/wiki-go/commit/db446f7baf51ec011ecc2b90f7af427563d48913))
- Merge pull request #143 from acc1729/open-graph ([db7b169](https://github.com/leomoon-studios/wiki-go/commit/db7b16933e3e8091c9eff0d44290b7200af42f5b))

## [v1.8.4](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.4) (2026-01-15)

- Updates configuration on config.yaml and README ([1ccd84d](https://github.com/leomoon-studios/wiki-go/commit/1ccd84dbf13f1244a17f62cbb5f76ad939612eca))
- Improves kanban documentation ([fdb9544](https://github.com/leomoon-studios/wiki-go/commit/fdb9544f4fd797cdc10d964d02e966743aeb6da6))
- Adds dynamic :::year::: shortcode for current year in site notice ([dd75ac1](https://github.com/leomoon-studios/wiki-go/commit/dd75ac1912c73f6a7719d07cef1fce07eb214364))
- Changes timezone to dropdown in settings ([b7a342f](https://github.com/leomoon-studios/wiki-go/commit/b7a342ff2ed705f772f126c25916918639470e60))
- feat: Highlight local links in red if their page does not exist ([fa9b91c](https://github.com/leomoon-studios/wiki-go/commit/fa9b91cd1f0230dbe0cf937be80cbc008615fd6a))
- fix: Use --danger-color CSS var instead of hardcoded red ([c8cf27d](https://github.com/leomoon-studios/wiki-go/commit/c8cf27df80fdde228990a10eac2f0624bbd33d02))
- fix: Remove generated binary from repo ([dfd4b91](https://github.com/leomoon-studios/wiki-go/commit/dfd4b91396ec9a9ba01452fd9e4005e2f96b5051))
- Merge pull request #134 from acc1729/bad-links ([c9894ec](https://github.com/leomoon-studios/wiki-go/commit/c9894ecd3a62e78750a91e0b59c6cc774821f8a3))
- Improves bad links detection for both document links and attachments by making the link red ([67e3a19](https://github.com/leomoon-studios/wiki-go/commit/67e3a194622c170def3b974e9aa67118ae2e32d2))

## [v1.8.3](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.3) (2026-01-07)

- Adds better bug and feature request templates ([43a4c52](https://github.com/leomoon-studios/wiki-go/commit/43a4c5299dc97514b0c917c2245fc5d0881db232))
- Fixes sitemap filtering entries based on access rules ([fec8c56](https://github.com/leomoon-studios/wiki-go/commit/fec8c561a05042aa786437f4480b8b2f6ba21750))

## [v1.8.2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.2) (2026-01-03)

- Adds one more bidirectional text fix for stats.css ([7cc607d](https://github.com/leomoon-studios/wiki-go/commit/7cc607dc8084dbe26d6bdf29025005fe964517e8))
- made password hash strength (bcrypt cost factor / rounds) configurable in config.yaml ([569bb4f](https://github.com/leomoon-studios/wiki-go/commit/569bb4fe9c1b3f73d75ceb7091a999b119434701))
- Adds github flavored infobox exmamples to website syntax guide ([770fb3f](https://github.com/leomoon-studios/wiki-go/commit/770fb3f3bbc5fd5b23b45a9c3f65edffa1ddfd71))
- Merge pull request #129 from SheevaPlug/master ([4af8e02](https://github.com/leomoon-studios/wiki-go/commit/4af8e026b1900b37636676e8234e50c171cfe952))
- Fixes rendering issue with code blocks inside blockquotes #132 details, emoji, highlight, stats, subscript, superscript, typography had this issue and it is fixed now. Also fixed editor highlight when inside codeblock. ([6033c55](https://github.com/leomoon-studios/wiki-go/commit/6033c5587c6c9cd0389e694f88ac3ee8c4a78bb0))

## [v1.8.1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.1) (2025-12-25)

- Adds more users/docs with different access roles to demo-site-files ([976a8c9](https://github.com/leomoon-studios/wiki-go/commit/976a8c90c0a590ab8e721091f53ef3cbe6199094))
- added cli option and environment variable for config file path (think docker secret) ([170caa8](https://github.com/leomoon-studios/wiki-go/commit/170caa8c3d9ec51dce4963a0d4c20a820309138b))
- added Makefile ([a59d191](https://github.com/leomoon-studios/wiki-go/commit/a59d191ffeb2a01d4f1faa47be2a1c5c2b51176f))
- git ignore emacs backup files ([e0c062d](https://github.com/leomoon-studios/wiki-go/commit/e0c062dec79034186e3768e42c7cb8f4a8f38829))
- Updates README and rest api documents ([61a66ab](https://github.com/leomoon-studios/wiki-go/commit/61a66ab0653647483c72e99de502da4e1d4a8c26))
- Updates demo-site-files to include access rule examples ([82e976a](https://github.com/leomoon-studios/wiki-go/commit/82e976a6e8ee69ba7bb487d74da29d5e08387127))
- Adds suppport for .sfd files as text-based attachments ([fcf049f](https://github.com/leomoon-studios/wiki-go/commit/fcf049f2e84b036746705138fe88db9b6f0310ff))
- Fixes a couple of issues with bidirectional text ([1baae63](https://github.com/leomoon-studios/wiki-go/commit/1baae6339d1bf1fb2e4af1fac53a11fac277c204))
- Merge pull request #124 from SheevaPlug/master ([d4eee76](https://github.com/leomoon-studios/wiki-go/commit/d4eee7673844194dc5957633b6f95eed311a3e81))
- Improves docs ([7f01778](https://github.com/leomoon-studios/wiki-go/commit/7f01778af7880e424b48b4f71d74a443ca3068ac))
- Fixes a corrupt config when config is generated from scratch #128 ([5c49315](https://github.com/leomoon-studios/wiki-go/commit/5c493155db87d95d040175dcca87f1e5508ea083))

## [v1.8.0](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.8.0) (2025-12-18)

- Adds backend for group-based access control ([67830fa](https://github.com/leomoon-studios/wiki-go/commit/67830faa1758184f7a2329fa13eb05427846e701))
- Adds frontend for access rules management ([d072ce6](https://github.com/leomoon-studios/wiki-go/commit/d072ce60b035b9f254c0d6797f7834fe9808863d))
- Fontend bugfixes ([fcd646b](https://github.com/leomoon-studios/wiki-go/commit/fcd646bdf6acafc91bf8466bc89a3b7e46f7c896))
- CSS improvements for user management tab ([dbadc8e](https://github.com/leomoon-studios/wiki-go/commit/dbadc8ed1fdb57a449366399141a9da0d15a5d19))
- Adds better dialog scrolls ([2caa8f9](https://github.com/leomoon-studios/wiki-go/commit/2caa8f903a1d43ea11ab21ec0be7fbcf1b86276e))
- Adds shadow to dialog header when scrolling ([c658012](https://github.com/leomoon-studios/wiki-go/commit/c658012f064aebf9956ecbd247213e7545f2c24a))
- Improves version history on mobile in horizontal orientation ([57bcfbc](https://github.com/leomoon-studios/wiki-go/commit/57bcfbcbba47c7f0e2096abb3659748ef1a832a9))
- Fixes close button not working on add rule dialog ([c597184](https://github.com/leomoon-studios/wiki-go/commit/c597184ad9e1cdab1a328edcdf9ff930d45aaa7d))
- Fixes confirmation dialog buttons alignment issue ([c1dab48](https://github.com/leomoon-studios/wiki-go/commit/c1dab48f5f1e9886525c682eb149b50ed48d95a2))
- Changes access-rules-manager.js to use a existing dialogs instead of alerts ([f4c7009](https://github.com/leomoon-studios/wiki-go/commit/f4c70093d37a11167819178223bd6deb2f96b727))
- Adds translation keys to en.json ([ea91861](https://github.com/leomoon-studios/wiki-go/commit/ea91861f5697ac6ccc45bc8208721c8a1c97513a))
- Adds CSS improvements ([4ed1529](https://github.com/leomoon-studios/wiki-go/commit/4ed15297a6fbbb16cabbb2e6d8e36312f4c281de))
- Minor bugfix with access-rules-manager.js ([d426d2c](https://github.com/leomoon-studios/wiki-go/commit/d426d2c531f76e117e18ec414a0b5463240043f1))
- Adds translations by ChatGPT for the new access control feature ([655f4f2](https://github.com/leomoon-studios/wiki-go/commit/655f4f20cdd91d21f5874565f8eb58613cdff9db))
- Adds backup feature in settings dialog ([afd9a53](https://github.com/leomoon-studios/wiki-go/commit/afd9a53cabf67a63812ad36c6bbe699cdf7f4a30))
- Adds translations by ChatGPT for the new backup feature ([330cd0b](https://github.com/leomoon-studios/wiki-go/commit/330cd0b4138d1ba04ed6833b9dd29490b3c98ffc))
- Minor CSS fixes ([d00da70](https://github.com/leomoon-studios/wiki-go/commit/d00da70568db0c7383089a18a0fe89dab0934e62))
- Changes add access rule to show document names instead of folder names ([e631d90](https://github.com/leomoon-studios/wiki-go/commit/e631d904f2f47d39518fde6faa59c0ed8778fcb7))
- Changes homepage to only allow "This document only" in access rules ([c462884](https://github.com/leomoon-studios/wiki-go/commit/c462884160ec84081d9d4af0cdcbc2d801625793))
- Fixes attachments not woking in access rules ([847fdef](https://github.com/leomoon-studios/wiki-go/commit/847fdef762bddf7fdda5b6a6c5e466edd1eddc50))
- Adds better styling for radio buttons ([72df42f](https://github.com/leomoon-studios/wiki-go/commit/72df42f2dfc8c8e6b5a32e48f9f2fdda6067f391))
- Updates iframe attributes to fix allowfullscreen warning ([a07180c](https://github.com/leomoon-studios/wiki-go/commit/a07180c5f1f4dcb913bb1fc5a0ee8a8abc051016))
- Adds various resouce loading optimizations ([e2f6b9b](https://github.com/leomoon-studios/wiki-go/commit/e2f6b9b059192ff22234d02390d04a1c86fcd20f))
- Prevents sticky hover on touch devices for editor buttons ([d646fe6](https://github.com/leomoon-studios/wiki-go/commit/d646fe6ee994a04d5b35e58aee53c6b54dd10336))
- Adds attachments dialog improvements ([a19c582](https://github.com/leomoon-studios/wiki-go/commit/a19c582aa02cd57505dd16acf5d5d8e39dcaaff7))
- Adds translations for attachment bulk deletion by ChatGPT ([0dba37e](https://github.com/leomoon-studios/wiki-go/commit/0dba37e51d67cbbb52c963d90dee0768d22ef835))
- Fixes 1password extension injection breaking syntax highlighting ([2f041dd](https://github.com/leomoon-studios/wiki-go/commit/2f041dd5b3d103f2a721b942474957ee9447b98b))
- Adds icons to access rule dialog options ([a4d3c79](https://github.com/leomoon-studios/wiki-go/commit/a4d3c7932cf60bb58cb01bf72d1d03827b4192a6))
- Updates security docs with the new features ([a619088](https://github.com/leomoon-studios/wiki-go/commit/a61908813bcc78ae60b03f55d1dd7adcf0f7a8e6))

## [v1.7.8](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.8) (2025-12-01)

- Adds scroll margin to prevent headings from being hidden behind the fixed top bar ([b975d53](https://github.com/leomoon-studios/wiki-go/commit/b975d536656ed34a2e4dfbc5f4789aa40ad36097))
- Hides sidebar on mobile up to 950px width ([75eb4c9](https://github.com/leomoon-studios/wiki-go/commit/75eb4c92ab7e3fad95ea93d4fc2e5875fa6e6b6e))
- Adds scrolling to new document dialog on mobile landscape ([d4b501e](https://github.com/leomoon-studios/wiki-go/commit/d4b501e5b0377ad591a2611e05c96fc35c8ee306))
- Adds github flavored infoboxes #109 ([3d57668](https://github.com/leomoon-studios/wiki-go/commit/3d57668562383e8eb346cb17a0650fa840f7c555))

## [v1.7.7](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.7) (2025-11-27)

- Adds PWA support ([2dafe99](https://github.com/leomoon-studios/wiki-go/commit/2dafe998750a1cb427e66d3d438c8c8ac83f36b0))
- Updates golang to 1.25.4 ([79c6944](https://github.com/leomoon-studios/wiki-go/commit/79c69440081d0ba6548707615ee2a8d63a0b51b8))

## [v1.7.6](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.6) (2025-11-24)

- Improves settings dialog tabs style on mobile ([e822f23](https://github.com/leomoon-studios/wiki-go/commit/e822f2340995e26ef665bc339d5297ff68c9bfc3))
- Adds bg blur and bg scroll for dialogs ([f852210](https://github.com/leomoon-studios/wiki-go/commit/f852210bf720af84aff25577d5673681434d8280))
- Updates Mermaid to version 11.12.1 ([39d8af6](https://github.com/leomoon-studios/wiki-go/commit/39d8af6d7ee59d3923f83e66a412269abf75ff11))
- Adds link management feature to demo site ([657e7b8](https://github.com/leomoon-studios/wiki-go/commit/657e7b87d1fb375570b54a81bcef183235d49691))
- Adds some missing translations for link management by ChatGPT ([1376dca](https://github.com/leomoon-studios/wiki-go/commit/1376dcac51d9138c6cd87557323f65aafbed6528))
- Adds en translation keys for 404 page ([80ac5a3](https://github.com/leomoon-studios/wiki-go/commit/80ac5a37759a0fe3ac7624e3e27bb80656c87d17))
- Adds translations for 404 by ChatGPT ([d8a37c6](https://github.com/leomoon-studios/wiki-go/commit/d8a37c668f3886dadd3f89fc608cde75b08002c1))
- Adds missing translation keys for confirmation-dialog to en.json ([5cd080c](https://github.com/leomoon-studios/wiki-go/commit/5cd080c45ef6a89cec82609796d2b26d45410775))
- Adds missing translations by ChatGPT ([36376e2](https://github.com/leomoon-studios/wiki-go/commit/36376e258b0e6188e5b2900b61393e5e07ffa05a))
- Improves the sidebar navigation drag on mobile ([f6bdead](https://github.com/leomoon-studios/wiki-go/commit/f6bdeada301b877405e7c7c53895d8fb68637132))
- Adds persistent login sessions #103 ([e5538f2](https://github.com/leomoon-studios/wiki-go/commit/e5538f2787ec7ce42c8c4110782ae75b4eee6623))

## [v1.7.5](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.5) (2025-11-03)

- Fixes editor performance issues ([09b032a](https://github.com/leomoon-studios/wiki-go/commit/09b032a9701e243ed0fa3b93b8d21036b0fa30d8))

## [v1.7.4](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.4) (2025-10-24)

- Fixes blank document rendering issue #104 ([f1727e5](https://github.com/leomoon-studios/wiki-go/commit/f1727e51cb6586728599eef9791322ed259cfad5))

## [v1.7.3](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.3) (2025-10-17)

- Improves breadcrumbs-container css ([55e59ce](https://github.com/leomoon-studios/wiki-go/commit/55e59ce87004d51f25a3314817d6e4e61befcbea))
- removes unnecessary padding for highlighted code blcoks ([0b42a25](https://github.com/leomoon-studios/wiki-go/commit/0b42a2555c8a5d50fb23aacbf0d9253b07327ecb))
- Disable browser back gesture to improve sidebar navigation ([d660b37](https://github.com/leomoon-studios/wiki-go/commit/d660b379a78e6d8a2c7d327750e85d26dbb06cd6))
- Rewrote sidebar-navigation ([d2ce7b8](https://github.com/leomoon-studios/wiki-go/commit/d2ce7b8eb09cccabaeaed4b051430ee962aa42b1))

## [v1.7.2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.2) (2025-09-26)

- Changes demosite folders to double digit to fix sorting ([3e97496](https://github.com/leomoon-studios/wiki-go/commit/3e97496a66fa6af25d1d75ab73524ed056367b7f))
- Fixes permission issue with Docker volumes ([732a89d](https://github.com/leomoon-studios/wiki-go/commit/732a89d158035210743e181df071b6eb37c8d14e))
- Merge pull request #100 from NicatorBa/master ([38c9626](https://github.com/leomoon-studios/wiki-go/commit/38c9626b0772c05805f2a6be631e4cabd4f0baee))
- Cleanup ([6a75378](https://github.com/leomoon-studios/wiki-go/commit/6a75378ba16df031a0ead098bbe4af7fec9ba2de))
- Fixes password change banner hiding breadcrumbs buttons #102 ([5efcdb7](https://github.com/leomoon-studios/wiki-go/commit/5efcdb71840177ae74e6779ddf595016f95b0128))
- Changes recent links from 30 days to 1 day ([cc98c6c](https://github.com/leomoon-studios/wiki-go/commit/cc98c6c79f6df19aeb9ab69e872d5272c26cc574))

## [v1.7.1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.1) (2025-09-13)

- Updates to go 1.25.1 and fixes docker file go version ([7235adf](https://github.com/leomoon-studios/wiki-go/commit/7235adf80ba4e46cfa63f0db72ccfb6b699b534d))

## [v1.7.0](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.7.0) (2025-09-13)

- upgrades wiki-go to use Go 1.25 and updates dependencies ([8b315fb](https://github.com/leomoon-studios/wiki-go/commit/8b315fbd55e9c0bb805efdb406f2637dbc75fbb7))
- Adds link document type with translations done by ChatGPT ([ca0b84f](https://github.com/leomoon-studios/wiki-go/commit/ca0b84fb9c7e4da78e1ed7366e9eb9806b8a223f))
- Adds info about links management feature ([0116996](https://github.com/leomoon-studios/wiki-go/commit/01169963c09b72da0602629f3f64a0334a2d716f))

## [v1.6.2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.6.2) (2025-08-17)

- updates local dns name in deploy script ([7a5075f](https://github.com/leomoon-studios/wiki-go/commit/7a5075f9fd752c2279f6360af0bd1c826d8c6d4d))
- Removes unnecessary margin-bottom that causes scrollbar to appear ([d951306](https://github.com/leomoon-studios/wiki-go/commit/d9513061e4dcfdbbcd6ee74dd395948b81f19a65))
- Changes new doc path to always be a child of current path #93 ([3d4863d](https://github.com/leomoon-studios/wiki-go/commit/3d4863d08510eaf1da2eef6d76b472ede4cf5b0d))
- Adds auto comlpletion to Document Path when creating a new document #94 ([0fd1ef1](https://github.com/leomoon-studios/wiki-go/commit/0fd1ef1d11a8aa0f0fa2f7bf7b4cb1fad53fe8e5))

## [v1.6.1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.6.1) (2025-08-04)

- Fixes release-binaries workflow to include commits since last stable tag ([110d2c4](https://github.com/leomoon-studios/wiki-go/commit/110d2c4857d9782746f5658d25911dcc30838d81))
- adds more editor buttons #90 ([c91ffb0](https://github.com/leomoon-studios/wiki-go/commit/c91ffb01a213ccba9e7cce649c0df7c846c8c04f))
- minor css improvement and adds deploy script ([b062163](https://github.com/leomoon-studios/wiki-go/commit/b062163c81b5ee508396564899695586a0a0cd9e))
- fixes shortcut in documentation ([c227043](https://github.com/leomoon-studios/wiki-go/commit/c2270434ffa7eabc7c3462bfd11b494f469321d0))
- fixes usage document ([2a34710](https://github.com/leomoon-studios/wiki-go/commit/2a347108ef0bd124c56401852e898facba8812d7))
- Fixes config not saving with double quotes when saving to config.yaml #91 And updates docs ([6df3c2c](https://github.com/leomoon-studios/wiki-go/commit/6df3c2c79f1ab281d4bd154d5d7b6f04560b2daa))

## [v1.6.0](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.6.0) (2025-08-02)

- Improves release binaries dev workflow ([89b1271](https://github.com/leomoon-studios/wiki-go/commit/89b1271a6fc23ee80724774153f696b9bcb0e173))
- top navigation improvement on desktop ([a5615db](https://github.com/leomoon-studios/wiki-go/commit/a5615db134acaffa01737b4ee39320185da5e817))
- Fixes hamburger and breadcrumbs alignment on mobile ([1b14843](https://github.com/leomoon-studios/wiki-go/commit/1b14843e91d27d15a5d0af61581d342b0616d44b))

## [v1.6.0-rc4](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.6.0-rc4) (2025-07-31)

- bugfix - mac hoktkey support #85 ([23b59cf](https://github.com/leomoon-studios/wiki-go/commit/23b59cf58c26b4a6a187819928c8715733ca6d0e))

## [v1.6.0-rc3](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.6.0-rc3) (2025-07-17)

- Moves all hotkeys to one place ([65cd928](https://github.com/leomoon-studios/wiki-go/commit/65cd928b974d8f34e666b8e36bce6e799d92f225))
- Adds mac hotkey support #85 ([b5eb673](https://github.com/leomoon-studios/wiki-go/commit/b5eb6739bd7332ebf94b6d6d64fdfa96dc0a23cf))
- Adds dynamic tooltip keyboard shortcuts #85 ([f4db4fe](https://github.com/leomoon-studios/wiki-go/commit/f4db4fe861017ea8559ea05957ccbe882b4516e6))
- Improves task item's "saved" notification #84 ([f4e990a](https://github.com/leomoon-studios/wiki-go/commit/f4e990a4f4deef68d6576df436ca1f4464466e20))
- Improves column "saved" notification #84 ([624c798](https://github.com/leomoon-studios/wiki-go/commit/624c798425d8807123aaf3b60a42a8d9f89fff04))
- Changes task and column save status to only show errors #84 ([5dc4b5e](https://github.com/leomoon-studios/wiki-go/commit/5dc4b5e921d61bbe528f0e8f0d26ef07ebeb879f))
- Removes drag handle icon and minor css changes ([e12a8c0](https://github.com/leomoon-studios/wiki-go/commit/e12a8c0331c858a9e2e3bc252ab73934e29c69a4))

## [v1.6.0-rc2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.6.0-rc2) (2025-07-09)

- Minor changes to the demo site ([cc583cc](https://github.com/leomoon-studios/wiki-go/commit/cc583cce2ba76bf05fce6000e822ccf1518b8f23))
- Moves back kanban examples to get started ([322984f](https://github.com/leomoon-studios/wiki-go/commit/322984f5abf2a7acc066121f6884cafcaa82582b))
- Removes github release from release-docker-dev.yml ([62e9ef9](https://github.com/leomoon-studios/wiki-go/commit/62e9ef9cad295cb5b61ae022b3fc4382d5c2ee9a))
- More consistent naming for workflows ([33ccda5](https://github.com/leomoon-studios/wiki-go/commit/33ccda559226239e4c0ebeea5d6605462e8152b0))
- Fixes task acion icons visibility #78 ([b77f89c](https://github.com/leomoon-studios/wiki-go/commit/b77f89c2dd688178e51875daf74ba7c6107e3e3b))
- Adds strikethrough text support to kanban tasks ([66ae7dd](https://github.com/leomoon-studios/wiki-go/commit/66ae7ddaaddecaaeb22ae0f7786da9e3bf126126))
- Fixes renaming task with image not displaying correctly ([4286183](https://github.com/leomoon-studios/wiki-go/commit/42861839818f940146a9c20776be7050928c2d23))
- Fixes double slashes in local file paths for images ([f14e905](https://github.com/leomoon-studios/wiki-go/commit/f14e9058c6cd552b64327edf25c8a1429bc64d0c))
- Fixes dragging tasks with images displaying very large ([190708d](https://github.com/leomoon-studios/wiki-go/commit/190708da81c6aa28cf2a84d30ac6f95926848948))
- Updates mermaid to 11.8.1 ([f610ea6](https://github.com/leomoon-studios/wiki-go/commit/f610ea6f74464b5f56c66fc430c5528393536f86))

## [v1.6.0-rc1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.6.0-rc1) (2025-07-05)

- Rewrite of details plugin ([9364cc4](https://github.com/leomoon-studios/wiki-go/commit/9364cc4f5d6b5a242f32172d676e95d7e01824fd))
- Updates README and documentation for the demo site ([922e819](https://github.com/leomoon-studios/wiki-go/commit/922e819f08101f8dbb412714f7ffe5f0665a6ed3))
- Updates config file for demo site ([ec1ff4f](https://github.com/leomoon-studios/wiki-go/commit/ec1ff4fa5506db8f370c3160eacb26a76609d0ab))
- Adds development release workflow ([c6b608b](https://github.com/leomoon-studios/wiki-go/commit/c6b608b4d2bb5e1c89d1448dc6da731035d1d288))
- Fix editor not showing in mobile landscape view #72 ([6c2be94](https://github.com/leomoon-studios/wiki-go/commit/6c2be946fc418bc457ee694c1e5618e6802f10cc))
- Adds indicator for referenced files in files tab #56 ([669715f](https://github.com/leomoon-studios/wiki-go/commit/669715f64403bd97786aa34eebf49677288bcfc2))
- Fixes sitemap.xml generation to use correct http/https #74 ([241b546](https://github.com/leomoon-studios/wiki-go/commit/241b546e16db6be2a62f837b90e303ee8569e31a))
- Basic frontmatter with kanban support ([66303d0](https://github.com/leomoon-studios/wiki-go/commit/66303d017a6af9f2a71b979796a1367c61908fca))
- Adds initial kanban live editing ([d20b743](https://github.com/leomoon-studios/wiki-go/commit/d20b743257a735624cf1f930f2761fc4009bccea))
- Adds kanban task edit/delete buttons ([6f98720](https://github.com/leomoon-studios/wiki-go/commit/6f98720053325fb65198fbba0cf2151e0ca49ed0))
- Adds logic to only include kanban css and js for kanban documents ([01fa080](https://github.com/leomoon-studios/wiki-go/commit/01fa0806b0c2f8b64b3be0fdb2bac24f77004fa8))
- Improves kanban styles ([65ce6ab](https://github.com/leomoon-studios/wiki-go/commit/65ce6ab7fdfafdce2d908b571612005ba8eb24f8))
- Bugfix to preserve markdown formatting when adding or renaming tasks ([2f254c0](https://github.com/leomoon-studios/wiki-go/commit/2f254c03431848070c69b4ea99cbe596d074d28d))
- Adds add/rename board functionality ([1560135](https://github.com/leomoon-studios/wiki-go/commit/1560135a6eb0d77379f258a6a37fe180357b1849))
- Adds add/rename board functionality 2 ([78d1d13](https://github.com/leomoon-studios/wiki-go/commit/78d1d135263f700ff85c8f73f3c29b3abea30427))
- Bugfix to save toggle state of newly added tasks This rewrites many parts of the kanban-live.js file to ensure that the toggle state of newly added tasks is saved correctly. ([aa869d4](https://github.com/leomoon-studios/wiki-go/commit/aa869d414b3c1f89af9e74b84e7202462d277c88))
- Adds advanced markdown support to kanban ([ffe1abd](https://github.com/leomoon-studios/wiki-go/commit/ffe1abd0ac57bc0991f84b6f27a54c8b3f65058d))
- Fixes content after kanban not rendering and getting deleted ([2e259e5](https://github.com/leomoon-studios/wiki-go/commit/2e259e5949c8580ea47d65e7300aeb15d13cd051))
- Fix - stops task dragging when task is being renamed ([44ead6e](https://github.com/leomoon-studios/wiki-go/commit/44ead6ec5bcc6b17c8fe1d607a89e3c8ffb813b7))
- Refactors kanban-live.js into multiple files ([6d01908](https://github.com/leomoon-studios/wiki-go/commit/6d01908c214ca776cdaf3180554ad43267133d44))
- Separates tasklist permissions into a shared file ([a5e3da0](https://github.com/leomoon-studios/wiki-go/commit/a5e3da0ed58985195adcc48296d1766baeb8a587))
- Fixes task id mismatch ([aeeec62](https://github.com/leomoon-studios/wiki-go/commit/aeeec62a2835c7cb897c5e9f0a64f4bde37cd38d))
- Adds highlight overlay mode to the editor ([f48a1e6](https://github.com/leomoon-studios/wiki-go/commit/f48a1e6e860984a31c6d9c835d9231195a47a6d3))
- Adds frontmatter highlighting to the editor ([e61b655](https://github.com/leomoon-studios/wiki-go/commit/e61b655f813cd860ce8f9b22c703d1332a6a38cd))
- Two fixes for kanban tasks - Adds highlight support - Fixes renaming a task with markdown content ([f7bd40b](https://github.com/leomoon-studios/wiki-go/commit/f7bd40bc42bfd4d153dbf1d25f0d38d958004c6a))
- Hide task action buttons when editing ([6022372](https://github.com/leomoon-studios/wiki-go/commit/60223724d9370d0423346d4093ed5f53f57a42c5))
- Adds multiple kanban support ([3607b74](https://github.com/leomoon-studios/wiki-go/commit/3607b74a837108aabef0ca285f4eadb7e87ab117))
- Adds kanban translations using ChatGPT ([03a6a51](https://github.com/leomoon-studios/wiki-go/commit/03a6a51533d40fe0f96d032853bf979c660497af))
- Bugfix: adds support for markdown between kanban boards ([41abde6](https://github.com/leomoon-studios/wiki-go/commit/41abde6ade1b0f7d9b170e24d9228fce59438f26))
- Minor style bugfix with kanban column rename ([8798a65](https://github.com/leomoon-studios/wiki-go/commit/8798a655975b88c9e72ce36b5ab611aacb6c2f8d))
- Adds support for kanban column deletion ([e062d15](https://github.com/leomoon-studios/wiki-go/commit/e062d15d8ea5ed5ffe01f86e1ba283930a0250e1))
- Adds missing translations for kanban column deletion using ChatGPT ([d5b39c0](https://github.com/leomoon-studios/wiki-go/commit/d5b39c0d955fcbd84f6bf6920ea170f39a43630a))
- Adds document type selection to new document dialog Markdown and Kanban for now Translations done by ChatGPT ([73bc4f9](https://github.com/leomoon-studios/wiki-go/commit/73bc4f97a248598abfd666f012459c049aa7a5fd))
- Adds support for multiple columns with same name in one kanban ([477cb28](https://github.com/leomoon-studios/wiki-go/commit/477cb283351d995fce9161a8b11cb0484d593bb5))
- Updates documentation ([e2e0975](https://github.com/leomoon-studios/wiki-go/commit/e2e097598e2ab381092ab8862bbdc03857656a66))
- Seperates add board dialog from base template ([f8353bb](https://github.com/leomoon-studios/wiki-go/commit/f8353bbda48d2d413acdb50f5877108400e2f88f))
- Fix - corrects Kanban terminology - columns are not boards ([8a1f23d](https://github.com/leomoon-studios/wiki-go/commit/8a1f23db7641b2d291f28418996b7a5bbfd65a54))
- Fixes some translations not working becuase they were hardcoded ([650df15](https://github.com/leomoon-studios/wiki-go/commit/650df155fef9a3f750a5e0781dc6ee88fd6a1b74))
- Refactors editor.js to use modular architecture ([b117461](https://github.com/leomoon-studios/wiki-go/commit/b117461f6b35d55e19564ebbf189b23262787b0e))
- Fixes file view button renaming instead ([ae9499a](https://github.com/leomoon-studios/wiki-go/commit/ae9499a67ebdf65b829460d2b5d38e975292091d))
- Fixes kanban pages not having docPath in the markdown preprocessor This is need to render the images/attachments correctly ([c4c0113](https://github.com/leomoon-studios/wiki-go/commit/c4c0113ab0e9210f85e4cea9d7a34dea2aed233f))
- Minor css fix for highlight in edit mode ([7747036](https://github.com/leomoon-studios/wiki-go/commit/7747036bf36c76fc4f18399f1fde3d32c38a0cda))
- Adds dockerhub dev release workflow ([739fc10](https://github.com/leomoon-studios/wiki-go/commit/739fc10bf6d21f186ed66cca31465771c977c067))
- Merge pull request #75 from leomoon-studios/dev ([832aa98](https://github.com/leomoon-studios/wiki-go/commit/832aa98e49730bfcb7e0957db4304e72569bcee3))

## [v1.5.6](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.5.6) (2025-06-03)

- Multiple improvements - Fixes print changing the dark theme to light - Changes mermaid theme when site theme is changed - Removes extra line for footnotes when printing - Fixes TOC not working in "Full width content" - Fixes search in mobile not being full width - CSS changes to make sure editor preview matches content renderer - Cleans up editor.css - Hide initial Mermaid raw text and show once diagram is rendered ([166073f](https://github.com/leomoon-studios/wiki-go/commit/166073f4087b640beed4daa343e308a2095eea26))

## [v1.5.5](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.5.5) (2025-06-01)

- Updates README ([472dd8f](https://github.com/leomoon-studios/wiki-go/commit/472dd8fa6dd9a9fb9717af53a7fad60d4335ca8c))
- CSS cleanup and minor visual improvements ([74f0893](https://github.com/leomoon-studios/wiki-go/commit/74f089308d40ecaa370b88cfa0bc1ab3c98d5a24))
- Updates to go 1.24.3 ([cae8b73](https://github.com/leomoon-studios/wiki-go/commit/cae8b73c171834523d1d58c3a4859b4e849272d4))
- Adds better script tag sanitization This will allow script tags inside codeblocks but not in the main document. ([81dd4e8](https://github.com/leomoon-studios/wiki-go/commit/81dd4e81d5563d099625dc17ee4c21ccf918d976))
- Removes extra border on footnotes ([b5c84ef](https://github.com/leomoon-studios/wiki-go/commit/b5c84eff3a409f9571b6c93b7dd85e4c0f083b0d))
- Changes golang version ([763021a](https://github.com/leomoon-studios/wiki-go/commit/763021ab19a2c5d3830f9c2a861c3728563cb376))

## [v1.5.4](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.5.4) (2025-05-26)

- Creates a new unsaved changes dialog to improve user experience #59 ([3b980da](https://github.com/leomoon-studios/wiki-go/commit/3b980da1e26b677d7918269b3a3aeddb69fe193b))
- Cleans up translations. Adds improved translations by ChatGTP-4.1. ([cf779ad](https://github.com/leomoon-studios/wiki-go/commit/cf779ad46a5e0243b515c722066d7761f0e69c7b))
- Adds timestamp to "dev" version string so static files are not cached when testing This will be overriden when building a release with the version number ([c7f094a](https://github.com/leomoon-studios/wiki-go/commit/c7f094a82bdf352c0859e51ea2d6a3c0fd208b50))
- Fixes search results layout when full width is enabled #61 ([7ffbb00](https://github.com/leomoon-studios/wiki-go/commit/7ffbb00846f77b7894efcfafaab35d5593ae4f1f))
- Adds ctrl+shift+f hotkey to focus the search box #62 ([ad4c8d4](https://github.com/leomoon-studios/wiki-go/commit/ad4c8d43908a8b98a662592a4ac265eadd2f2aa2))

## [v1.5.3](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.5.3) (2025-05-24)

- Typography improvement and wiki-go binary cleanup #57 for 1/2, 1/4, and 3/4 they must be in () for it to be processed. ([957590f](https://github.com/leomoon-studios/wiki-go/commit/957590fa8b8f1461016172580787eef18d9c74a2))
- Fixes undoing when there's no history erases content #58 ([4371f83](https://github.com/leomoon-studios/wiki-go/commit/4371f83cfa4fbbd6e018bcc39962b25a3d332aef))
- minor cleanup ([9db08c7](https://github.com/leomoon-studios/wiki-go/commit/9db08c743402bd49048875708a5b70e0b7619530))
- Fixes unsaved changes issue with translation done by ChatGPT #59 ([41c6e21](https://github.com/leomoon-studios/wiki-go/commit/41c6e21318a10e5aaaea6b3902a70600585801f6))
- Adds auto disable debug mode when releasing #60 ([d19dfa9](https://github.com/leomoon-studios/wiki-go/commit/d19dfa96f3706f776e15445e00eb499447c332ad))
- Moves auto disable debug mode after checkout #60 ([491e72e](https://github.com/leomoon-studios/wiki-go/commit/491e72e9d298b0e8917e3f6d49d84768ed1e886d))

## [v1.5.2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.5.2) (2025-05-22)

- Adds ctrl+shift+p to toggle preview and minor cleanups ([aa47cd9](https://github.com/leomoon-studios/wiki-go/commit/aa47cd95fb3f39bc395cd0e29011138d66a9b2a3))
- Fixes rendering attachments with spaces in the filename #54 ([0e35368](https://github.com/leomoon-studios/wiki-go/commit/0e3536812415531fd691ab3062aa0cbcad7f7236))

## [v1.5.1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.5.1) (2025-05-20)

- Fixed some example stats markdown content - recenter to recent ([fcf082e](https://github.com/leomoon-studios/wiki-go/commit/fcf082e1028167ab618c19a0d908b104396de4c6))
- Merge pull request #48 from chrismalek/master ([2f3f703](https://github.com/leomoon-studios/wiki-go/commit/2f3f70329888ef012596d06f3fa7a8e9c9ced18f))
- Fixes stats shortcodes evaluating when inside inline code blocks #53 ([0ead998](https://github.com/leomoon-studios/wiki-go/commit/0ead9983b56f9d81f202c47a7f02e7b97a4a9d50))
- Fixes stored XSS attacks by disabling unsafe HTML in Goldmark #52 ([62e6bf1](https://github.com/leomoon-studios/wiki-go/commit/62e6bf1c313dfa069d402ca42ebf96090901c0c0))
- Fixes stored XSS attacks by reverting previous patch and sanitizing script tags #52 ([461d8bd](https://github.com/leomoon-studios/wiki-go/commit/461d8bdaa1c0b73be6209aa2452e9e64cfff1feb))
- Adds api/files/rename endpoint example ([404c169](https://github.com/leomoon-studios/wiki-go/commit/404c169fc79ca2766cca6180c6ffb61c99d65592))
- Adds two types of sitemap #46 ([f6246c0](https://github.com/leomoon-studios/wiki-go/commit/f6246c041fea3ffa3aaed104105885ff7f09594e))
- Minor css cleanup ([1b2bcca](https://github.com/leomoon-studios/wiki-go/commit/1b2bcca81fedded18cb1e82d33acbc22f321a42c))
- Adds full width content option with translation by ChatGPT ([65c9228](https://github.com/leomoon-studios/wiki-go/commit/65c92284f4e1ce92eb0ab9066af1152ab3d014d3))
- Adds word wrap toggle button with `Alt+Z` shortcut ([a3709cc](https://github.com/leomoon-studios/wiki-go/commit/a3709cce862858531df3abf6e91e936fd7322e7d))
- Adds the missing toc shortcode button to editor toolbar ([f06fff8](https://github.com/leomoon-studios/wiki-go/commit/f06fff8272fb0cb81047c160991ec5d9a265f092))
- CSS changes to make three renderers look consistent - Page renderer - Editor preview - Version preview ([8318912](https://github.com/leomoon-studios/wiki-go/commit/8318912a14db7e89efcf7c393667ab406edb6817))

## [v1.5.0](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.5.0) (2025-05-19)

- Adds issue templates ([b1592a5](https://github.com/leomoon-studios/wiki-go/commit/b1592a57ff9f39013a1922e70e9aad01e2b2c31b))
- Bugfix: fixes SuperscriptPreprocessor malforming footnote parsing #44 ([5868d96](https://github.com/leomoon-studios/wiki-go/commit/5868d962385c0dc81e7dbfa9754e7af0a1cd3e0c))
- Adds logo offset if it exists ([dee58c5](https://github.com/leomoon-studios/wiki-go/commit/dee58c58d439f20d760a36debcc5999a8f30f881))
- Updates README.md ([d4547bf](https://github.com/leomoon-studios/wiki-go/commit/d4547bf03ff9363fdeabf635360ae86e50471111))
- Adds demo site files ([6d7d369](https://github.com/leomoon-studios/wiki-go/commit/6d7d3690f113a3c0ced1a9ead46934fb8d139bfe))
- Adds Document Sorting and Naming to README #45 ([71612ed](https://github.com/leomoon-studios/wiki-go/commit/71612ed245bef87f46ce74670d97d86a042fe2a0))
- Adds "Enable link embedding from clipboard" setting #12 ([2c9b4e0](https://github.com/leomoon-studios/wiki-go/commit/2c9b4e090a0d28f39ebcaa7cdcbd08b4cecea3db))
- Adds translation for "Enable link embedding from clipboard" using ChatGPT #12 ([de36e6f](https://github.com/leomoon-studios/wiki-go/commit/de36e6f3726821f94989f6abb86d7b269825a121))
- Adds Document Sorting and Naming to README #47 ([0ee3546](https://github.com/leomoon-studios/wiki-go/commit/0ee354633defccdc3678106c7b1e36d1cd542127))
- Re-orders the new document dialog ([86809bf](https://github.com/leomoon-studios/wiki-go/commit/86809bf46de583642c6a9cd182c504981ef74a65))
- Adds settings.hide_attachments option #41 ([b6058f5](https://github.com/leomoon-studios/wiki-go/commit/b6058f5039e5245d385ada10b50b0cb506092fbc))
- Adds translations for settings.hide_attachments using ChatGPT ([c346480](https://github.com/leomoon-studios/wiki-go/commit/c346480be1aeaceaf4ccc6d8808872fa08a77600))
- Initial release of markdown table editor using mte-kernel 2.1.1 #43 ([b8d6100](https://github.com/leomoon-studios/wiki-go/commit/b8d6100ae4e0844258c41fcd05946fc606ee00a6))
- Adds more keyboard shortcuts #50 ([e91b1d0](https://github.com/leomoon-studios/wiki-go/commit/e91b1d09f4507b13709fe79bf1d3180a0a79337b))
- Bugfix: fixes column move left breaking table if it's the first column ([7bbbef6](https://github.com/leomoon-studios/wiki-go/commit/7bbbef69e6b74a21c8c5728a3e7fa4db24992f88))
- Bugfix: fixes direction blocks not being processed correctly ([753c604](https://github.com/leomoon-studios/wiki-go/commit/753c6043c474fb61b9f2fc631f0701a5968b3b18))
- Adds renaming to attachments dialog with translation by ChatGPT #16 ([8d3e9e0](https://github.com/leomoon-studios/wiki-go/commit/8d3e9e0c6eb572a0fee3fb94c3984bed07c0d414))

## [v1.4.5](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.4.5) (2025-05-13)

- Adds 404 not found template page ([8fd6205](https://github.com/leomoon-studios/wiki-go/commit/8fd6205837e402057ab1df6db433276952153509))
- Adds SECURITY.md ([f82d719](https://github.com/leomoon-studios/wiki-go/commit/f82d7190d9954b18f7cc452590e5d7621b8200c0))
- Improves config file generation/validation This ensures that the config file has all the required fields if missing ([203da29](https://github.com/leomoon-studios/wiki-go/commit/203da290a28aedb47173c516d7d4b128283d3667))
- Adds rate limiting/ban for logins based on IP ([63ef359](https://github.com/leomoon-studios/wiki-go/commit/63ef3593df1eda90aa3b4ab8925482bfe1023c78))
- Adds Security Settings ([5b67520](https://github.com/leomoon-studios/wiki-go/commit/5b67520fd319e356e18b9b9ecdb6bfdc302ec45d))
- Adds translations for security settings and login ban using ChatGPT ([6044813](https://github.com/leomoon-studios/wiki-go/commit/6044813c16087a01363e0af299fa9e4f763fc105))
- Updates README.md, SECURITY.md ([2dd106e](https://github.com/leomoon-studios/wiki-go/commit/2dd106e82dd0a003f4e99347c2467d8e01bfe219))
- Minor cleanup ([7de9085](https://github.com/leomoon-studios/wiki-go/commit/7de9085c59e186c9c184d505e2a01e7e99bc9730))
- Adds refreshSidebar after successful import ([f8e5aff](https://github.com/leomoon-studios/wiki-go/commit/f8e5affc650407c101aea62fb215782538094537))
- Improves settings security tab and more cleanups ([ec83eab](https://github.com/leomoon-studios/wiki-go/commit/ec83eabbe1d6e261e898dc9afeb380b0c56f156b))
- Splits styles.css into multiple files ([5a6f21a](https://github.com/leomoon-studios/wiki-go/commit/5a6f21a45b5ecbf94606a4bf8db3add8e6e9a563))
- Makes sidebar header and footer fixed ([5949a4a](https://github.com/leomoon-studios/wiki-go/commit/5949a4a9c9cfb455ff546620aaf1c8a8c392726f))
- Adds scroll to active document in sidebar #40 ([454addf](https://github.com/leomoon-studios/wiki-go/commit/454addf2a97dfac91418bf1108591576981d0109))
- Improves login ban settings ([fd46518](https://github.com/leomoon-studios/wiki-go/commit/fd4651838a0bf56a3c0b515b874f76e322bf3644))
- Adds login redirect to original page when wiki is private #42 ([f520edd](https://github.com/leomoon-studios/wiki-go/commit/f520edd072511bcff06fd17026770e66afe53683))

## [v1.4.4](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.4.4) (2025-05-08)

- Improves folder nesting indicator Stops it from flashing and has open/close indicator ([c007e5e](https://github.com/leomoon-studios/wiki-go/commit/c007e5e0ad7a265d066dabcfbbf18ef462527e92))
- Adds versioning to all cached files #34 This would change the URL of cahced files on a new release which would force those files to be downloaded again. ([18664a5](https://github.com/leomoon-studios/wiki-go/commit/18664a587ade1cb92b9a7fe6552cf5f646f81b14))

## [v1.4.3](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.4.3) (2025-05-07)

- Refactors editor.js - lots of cleanup - moves cm styles to css files ([3e6ddc0](https://github.com/leomoon-studios/wiki-go/commit/3e6ddc07c3971f8c70417d9166246c42466a357e))
- Updates all modules and vendor ([0c1f9d1](https://github.com/leomoon-studios/wiki-go/commit/0c1f9d1f49e063ef7ea47b9b5879c6848164806c))
- Subscript processor bugfix ([3acf9ea](https://github.com/leomoon-studios/wiki-go/commit/3acf9eadc2884f020e7471288f2180c0260dbf52))
- Adds auth for api endpoints when wiki is set to privatge - ServeFileHandler - ListFilesHandler - GetCommentsHandler ([ad8b04f](https://github.com/leomoon-studios/wiki-go/commit/ad8b04fc3f0c201b08c69b5a3726ed5d794d1244))
- Adds folder nesting indicator ([c004397](https://github.com/leomoon-studios/wiki-go/commit/c004397df6d2a4b7af34451f0b60b07e3c19f804))
- Adds better slug generation using gosimple/slug ([7495810](https://github.com/leomoon-studios/wiki-go/commit/7495810aef652dd1e108d4f36d23e92f21767d39))
- Adds webp file attachment support ([734afc8](https://github.com/leomoon-studios/wiki-go/commit/734afc8592a68637533a281674dc1aa0ebca9fdb))
- Adds custom.css for user customization ([47bd1f5](https://github.com/leomoon-studios/wiki-go/commit/47bd1f50e7d2954f25fee063d0f8e7ebf3182f74))

## [v1.4.2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.4.2) (2025-05-05)

- Adds image upload from clipboard #15 ([e87d8a8](https://github.com/leomoon-studios/wiki-go/commit/e87d8a8e93554018d9b8f674120736f4948f30c3))
- Adds auto link embedding from clipboard #12 ([87a816a](https://github.com/leomoon-studios/wiki-go/commit/87a816a534512fb9907260f6885680b9937555a2))
- Disables custom TaskListPreprocessor to use the built-in one ([9b97edc](https://github.com/leomoon-studios/wiki-go/commit/9b97edcfb4ddde5c601fad54367aaa75f48ae39f))
- Adds live checkbox editing when logged in as admin or editor #24 ([bc22ea5](https://github.com/leomoon-studios/wiki-go/commit/bc22ea5a94a39dfed4b44835de6778c53c8a0a09))

## [v1.4.1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.4.1) (2025-05-05)

- Adds preview reset for version history dialog ([7450bd5](https://github.com/leomoon-studios/wiki-go/commit/7450bd543b45bafa83998780184da198039d7b49))
- Fixes migrate not letting config.yaml creation for new setups ([d31ff3c](https://github.com/leomoon-studios/wiki-go/commit/d31ff3c5e44dbc43900f336c34d67f2ff6d01004))

## [v1.4.0](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.4.0) (2025-05-04)

- Changes from admin/viewer to user roles: admin/editor/viewer with migration script ([e62ac90](https://github.com/leomoon-studios/wiki-go/commit/e62ac908fc4a7356a725fb53daddef3e70a13bea))
- Adds rest api examples ([7586621](https://github.com/leomoon-studios/wiki-go/commit/7586621203b16c7664cf7d959d347956ddb2b76d))
- Improves document handling in editor.go ([3ee2b49](https://github.com/leomoon-studios/wiki-go/commit/3ee2b494b75bbfc63eaa29563f28aa633049485b))
- Adds import feature as zip file Import zip package must have folder structure and only md files. ([a71db6b](https://github.com/leomoon-studios/wiki-go/commit/a71db6bb2b1d9354621bcc901806b3e7cc19f198))
- Adds missing translations using ChatGPT ([23faf2b](https://github.com/leomoon-studios/wiki-go/commit/23faf2b9e9ea17d590b3d6dec6e0533d41036288))
- Cleans up all json api endpoints to be more consistent ([93cc457](https://github.com/leomoon-studios/wiki-go/commit/93cc457c755efeb5e9b6b17a8eb6b6132777590a))
- Adds moving of comments when document is moved/renamed ([8f90020](https://github.com/leomoon-studios/wiki-go/commit/8f9002094fd3eaa5f7a612bca0a12e9c37f3d1e4))
- Bugfix - fixes moving to root not working Documents with category now can move to root ([65acfac](https://github.com/leomoon-studios/wiki-go/commit/65acfacebcf1d61791a348825fdfa00890e36110))
- Minor logging cleanup ([f2799ab](https://github.com/leomoon-studios/wiki-go/commit/f2799ab6ffb423e8cd6697a13041c080ce06dfc0))

## [v1.3.19](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.19) (2025-04-28)

- Fixes move/rename button visibility ([1919dc6](https://github.com/leomoon-studios/wiki-go/commit/1919dc60951a7c4f7ffb80c314ef53da6aff99d8))
- Adds the missing styles for hyperlinks in document history preview ([bf756ba](https://github.com/leomoon-studios/wiki-go/commit/bf756baa4d55137799b2795d9da6d76e56b717b7))
- Fixes anchor picker last item being clipped and stops it from changing size ([7f4fbfd](https://github.com/leomoon-studios/wiki-go/commit/7f4fbfd2066b2837692b718512647ef252a68d07))

## [v1.3.18](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.18) (2025-04-27)

- Adds keep me logged in option ([22ad5c1](https://github.com/leomoon-studios/wiki-go/commit/22ad5c19efacc2e08bc28341aa478d903325cb29))
- Adds translation variables to login page that were missing ([a6d038c](https://github.com/leomoon-studios/wiki-go/commit/a6d038c898529ec18cd3c1bf420e2efcc3fbc7a0))
- Adds translations for "keep me logged in" using ChatGPT ([b79747c](https://github.com/leomoon-studios/wiki-go/commit/b79747c63804ec7e9b08d897121b17b51a08495d))

## [v1.3.17](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.17) (2025-04-27)

- Improves custom favicon handling #14 ([7d09d0f](https://github.com/leomoon-studios/wiki-go/commit/7d09d0f1191100febde8fa26ffcea345c8fa6c6f))
- Adds native ssl support while keeping reverse proxy support ([4e6c29d](https://github.com/leomoon-studios/wiki-go/commit/4e6c29da9ba4d0d18e9ce1f59875b1d5ac96a476))
- Adds labels to docker image #19 ([92784c2](https://github.com/leomoon-studios/wiki-go/commit/92784c23e07b7b95b183582e1e33e386ed05f5b3))
- improves readme and adds nginx reverse proxy example ([78768a8](https://github.com/leomoon-studios/wiki-go/commit/78768a8cb5fa9780086b54a730ff1d541116afeb))

## [v1.3.16](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.16) (2025-04-25)

- Adds hover styles to heading anchors - hidden by default, visible on hover ([bb15711](https://github.com/leomoon-studios/wiki-go/commit/bb1571127bd373e94a21a0a3d5b1d1e453385a6d))
- Changes anchor button to first list docs then anchors of the selected doc ([514f24e](https://github.com/leomoon-studios/wiki-go/commit/514f24ebd52d4480b3923bb323baaba05f92f9f4))
- Adds cache-control for static files ([af1a572](https://github.com/leomoon-studios/wiki-go/commit/af1a57241656afeb1a9039098e90008132cc40ea))
- Some css cleanup ([bededbf](https://github.com/leomoon-studios/wiki-go/commit/bededbf37b7dc99cbfb07910966b4d136b187494))

## [v1.3.15](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.15) (2025-04-22)

- Adds headinganchor preprocessor ([a44649f](https://github.com/leomoon-studios/wiki-go/commit/a44649fce4e4934fd676645895e6a65fa3189c10))
- Removes homepage from documents list ([591ab4c](https://github.com/leomoon-studios/wiki-go/commit/591ab4c0164f994080a028eee307e3ebd7890066))
- Improves js heading anchor id generation ([216cc17](https://github.com/leomoon-studios/wiki-go/commit/216cc17cc49a6dd3043456b9a2931489616564fc))
- Adds translations for anchor picker. Translation done by ChatGPT. ([cc14772](https://github.com/leomoon-studios/wiki-go/commit/cc14772fa0f9b4b33ef5ee74c61293ff1916ace2))
- Adds updating of anchor links list when the popup is opened ([4f3698d](https://github.com/leomoon-studios/wiki-go/commit/4f3698dbcddd0ac82bfe001642dce70889ebb0c6))
- Adds disable_file_upload_checking support used for local deployment ([14de6da](https://github.com/leomoon-studios/wiki-go/commit/14de6daeebdbe09e175d0b7baeb5d888afdacc59))
- Adds Folder Structure to README ([41ffcdf](https://github.com/leomoon-studios/wiki-go/commit/41ffcdff8781b7bff09e1cd5b3c4ed3006f6cd23))

## [v1.3.14](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.14) (2025-04-20)

- Improves CSS for recent edits stats ([6554729](https://github.com/leomoon-studios/wiki-go/commit/655472934f8ae9933e817599582124feee009c39))
- Adds translations for document picker. Translations done by ChatGPT. ([a248aff](https://github.com/leomoon-studios/wiki-go/commit/a248aff36832d639b3226e02393524844b56457e))
- Adds bidirectional text support to document picker search input ([96646fa](https://github.com/leomoon-studios/wiki-go/commit/96646fa36bc2e8811411a84b97a4fd1a9f860171))
- Changes comments timestamp from Unix to human readable ([7318111](https://github.com/leomoon-studios/wiki-go/commit/731811187a844f6b305a629f01b77eca089f8134))

## [v1.3.13](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.13) (2025-04-20)

- Adds document linking to the editor ([f25a944](https://github.com/leomoon-studios/wiki-go/commit/f25a944f222c2da9608eb541ba1223166fc58e82))
- Refactors all popups to use the same positioning logic: positionPicker ([1ca7f43](https://github.com/leomoon-studios/wiki-go/commit/1ca7f43079e4973103d7314c69f79b271ef8717e))
- Updates go to 1.24.2 and its dependencies ([c6a9771](https://github.com/leomoon-studios/wiki-go/commit/c6a977199cf07abe9e284d6ba2df5a525c9636ef))
- Adds vendor dependencies to codebase for offline building ([acbe190](https://github.com/leomoon-studios/wiki-go/commit/acbe1905932a323f71cc26c5bbb428e3627acc8c))

## [v1.3.12](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.12) (2025-04-17)

- Fixes print adding extra blank page ([0c473ef](https://github.com/leomoon-studios/wiki-go/commit/0c473ef92250704f060009c216076b3c9ebceed6))
- Updates README with new preview.png and removes extra screenshots ([b84bd01](https://github.com/leomoon-studios/wiki-go/commit/b84bd011f6808a9cb6fbbb2ba2e639bdd29baf78))
- adds language support for Czech, Danish, Finnish, Hebrew, Dutch, Norwegian, Polish, Swedish, and Turkish ([583c964](https://github.com/leomoon-studios/wiki-go/commit/583c964051df52c21749ccc13ab7f6d3b71c3c43))
- Splits settings dialog from two tabs to three tabs to make sure it fits on mobile devices. ([9b8c331](https://github.com/leomoon-studios/wiki-go/commit/9b8c33113a01023f43f712f06a4dbe16996825a1))
- Adds missing styles for blockquotes in version history preview ([9a01a8c](https://github.com/leomoon-studios/wiki-go/commit/9a01a8c712d767aff6ddf3dae601440be6fc5aca))

## [v1.3.11](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.11) (2025-04-13)

- Improves README with demo site info ([289bdd3](https://github.com/leomoon-studios/wiki-go/commit/289bdd3f9393df9a45f380bdb39581362a0deac7))
- Fixes toolbar showing login with category and subcategory pages ([94765d4](https://github.com/leomoon-studios/wiki-go/commit/94765d4fac6465f4b80d13613633cdd498d24f17))

## [v1.3.10](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.10) (2025-04-12)

- Adds version display to footer when user is admin ([10168ec](https://github.com/leomoon-studios/wiki-go/commit/10168ecb8835329a0ebe38dae3c90467d5281adc))
- Updates README with demo site info ([afb7c4c](https://github.com/leomoon-studios/wiki-go/commit/afb7c4c1f7a6e02570fb1718fd8153be8e1e4a36))

## [v1.3.9](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.9) (2025-04-12)

- Minor css fix for version in footer ([385803c](https://github.com/leomoon-studios/wiki-go/commit/385803c804b45117a6dcb5f2f409ce6c6a9ef70b))
- Updates README.md and changes workflow names ([7d78cce](https://github.com/leomoon-studios/wiki-go/commit/7d78ccebf9de3f50331e2fe5a77e07fb3126d908))
- Adds deleting of comments when a document is deleted ([00479a4](https://github.com/leomoon-studios/wiki-go/commit/00479a48f9e9028d384261f329ff12a0ad91f0bf))
- Adds server-side auth to the base template to make toolbar loading faster ([594f895](https://github.com/leomoon-studios/wiki-go/commit/594f8958ccc1b97106f7dae4a2609648784998cd))
- Fixes editor delay when user is not logged in ([acbd302](https://github.com/leomoon-studios/wiki-go/commit/acbd3029286f95441b87f7f7ce8ddf14c2e6bba2))
- Adds page reload when user logs out to close comments section ([b37f8d6](https://github.com/leomoon-studios/wiki-go/commit/b37f8d6016e299148e0fb6cc607f620f6700f475))

## [v1.3.8](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.8) (2025-04-05)

- Fix text clipping on the right side in edit mode ([cee7b8c](https://github.com/leomoon-studios/wiki-go/commit/cee7b8cc7d826158a71934d71d05a0ad12a95ae7))
- Workflows cleanup: - Seperates binaries from docker ([98c7586](https://github.com/leomoon-studios/wiki-go/commit/98c75862a3f7da955ac56063df1a348ff2795acf))
- Updates gitignore to: - Only track README.md and SCREENSHOTS.md - Ignore all other .md files generated by wiki-go or while testing ([0da68fb](https://github.com/leomoon-studios/wiki-go/commit/0da68fb653e45755a9eec0e455fc52e6ebdfe6b6))
- Adds version number to footer ([6d0b2f4](https://github.com/leomoon-studios/wiki-go/commit/6d0b2f405294cc9339f6aa70159f2ce508887bd3))
- Fixes sidebar on mobile when password warning is active ([f53e283](https://github.com/leomoon-studios/wiki-go/commit/f53e283ee3ef8355b5329d40411e2b843ddc5ba8))
- Fixes docker build version displaying as "dev" ([a669343](https://github.com/leomoon-studios/wiki-go/commit/a66934384c2d5f39e41ec9c8cd48e4b59a165167))

## [v1.3.7](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.7) (2025-04-03)

- Updates README and adds SCREENSHOTS ([ac52b58](https://github.com/leomoon-studios/wiki-go/commit/ac52b5812221d05ff48f7c82c3d89c8efb35299a))
- Adds release body and minifying to release.yml ([a5980ee](https://github.com/leomoon-studios/wiki-go/commit/a5980ee3f10b4dab0befcff4520ffbe19865624d))
- Adds missing css and js minify to docker image and improves find for both ([cddf63a](https://github.com/leomoon-studios/wiki-go/commit/cddf63a3ff56bd503f9bf2a65dd75977c1bb4412))

## [v1.3.6](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.6) (2025-04-02)

- Adds commit logs to release notes ([6c55721](https://github.com/leomoon-studios/wiki-go/commit/6c5572125ea29f0e74e9ebef7eaf505ac76ca66a))
- Adds the missing body tag for the release notes ([f9c108b](https://github.com/leomoon-studios/wiki-go/commit/f9c108bd0f51c436c91fc2fbc164d17300855c0c))
- Changes emoji loading in editor from AJAX to preload ([a0ae673](https://github.com/leomoon-studios/wiki-go/commit/a0ae67335ed379d01184696f0e33b7a079ca38c4))

## [v1.3.5](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.5) (2025-04-01)

- Adds more emojis ([e234ff5](https://github.com/leomoon-studios/wiki-go/commit/e234ff5348b6d47627ebc8a8b3104a357ea36c0d))
- Fixes search close button not centered vertically ([f356b72](https://github.com/leomoon-studios/wiki-go/commit/f356b72be851ad8663670c4b36b01fd2929cb626))
- Improves file upload dialog workflow. No more confirmation boxes. ([b1e11b8](https://github.com/leomoon-studios/wiki-go/commit/b1e11b8d0206258af1bcd870e66ed77d24aecf36))
- combines theme-init.js and theme-manager.js ([eeef89e](https://github.com/leomoon-studios/wiki-go/commit/eeef89e6fc91923c355949f7634a9d9fd88991da))
- Fixes version history not working with home page ([dff2a37](https://github.com/leomoon-studios/wiki-go/commit/dff2a37e11c964c9ee804c343dcfc674f3ef7b66))
- Minor css fix ([2506eb3](https://github.com/leomoon-studios/wiki-go/commit/2506eb3a264f0b0329e58ed7343ae0dc18c5ca1b))

## [v1.3.4](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.4) (2025-03-31)

- Refactors file utilities from base.js ([be39445](https://github.com/leomoon-studios/wiki-go/commit/be39445afe20e40c991c6ec786b85ff7f4cc8656))
- Refactors theme manager from base.js ([ef001e6](https://github.com/leomoon-studios/wiki-go/commit/ef001e6faf9cc54107e357d6c1bfc5e4257fe2f9))
- Refactors version history from base.js ([d740709](https://github.com/leomoon-studios/wiki-go/commit/d74070916062f64c6f02405e42c2e1e6ba059553))
- Refactors dialog system from base.js ([9a18256](https://github.com/leomoon-studios/wiki-go/commit/9a18256c9326e2a32d7d8cf4e59f6216ce019b6b))
- Refactors sidebar navigation from base.js ([f830aee](https://github.com/leomoon-studios/wiki-go/commit/f830aeeddfb97b5c77c19ca8e51334052a130a48))
- Refactors authentication from base.js ([0097485](https://github.com/leomoon-studios/wiki-go/commit/00974854db0e56a3fea87b6a7acf122c8c3096f6))
- Fixes esc handling for dialogs and edit mode ([6d9e812](https://github.com/leomoon-studios/wiki-go/commit/6d9e812a1ed72f4296a40a0f55380bea7bfa5a69))
- Refactors document management from base.js ([d719c1b](https://github.com/leomoon-studios/wiki-go/commit/d719c1ba0405435d28d89310f0ca5ecd7a2ac4ce))
- Refactors copy button from base.js ([deabcbb](https://github.com/leomoon-studios/wiki-go/commit/deabcbb2cf8e1bca50c82d038ea211f4ef34ca4c))
- Refactors settings from base.js ([3811901](https://github.com/leomoon-studios/wiki-go/commit/381190111ece5e4ca63d863b464617e96e350121))
- Refactors file upload from base.js ([67ba27b](https://github.com/leomoon-studios/wiki-go/commit/67ba27b3d2dcabb418a5976a32dd39eba8bf73bd))
- Refactors move document from base.js ([6769d08](https://github.com/leomoon-studios/wiki-go/commit/6769d08df877a0ed5883229d6807da19bb4fb2e9))
- Refactors keyboard shortcuts from base.js ([cc8d719](https://github.com/leomoon-studios/wiki-go/commit/cc8d719c5bcb515e76fd9835773a9da5421adce8))
- Rebases the remaining base.js ([a07a16c](https://github.com/leomoon-studios/wiki-go/commit/a07a16c05d7e86232258c314a366d3e218557676))
- Refactors emojis to load from emojis.json file for both editor.js and emoji.go ([ef71284](https://github.com/leomoon-studios/wiki-go/commit/ef712840ae98313061efc1a884227c3ec77a56f8))

## [v1.3.3](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.3) (2025-03-29)

- Removes unneccesary maxUploadSize fetch at page load ([f9b2e5b](https://github.com/leomoon-studios/wiki-go/commit/f9b2e5b51d83fedefeed6a484f81821c295e21d8))
- Hides comments when printing ([5ade04b](https://github.com/leomoon-studios/wiki-go/commit/5ade04bbe99e61e23e1a325f81fa7e5af2e2a86e))
- removes easymde references since we're not using it anymore ([8de1dc4](https://github.com/leomoon-studios/wiki-go/commit/8de1dc4f6a22594eed41419bc25572368eb8ff76))

## [v1.3.2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.2) (2025-03-28)

- Adds macos arm64 build to release workflow ([a9ec426](https://github.com/leomoon-studios/wiki-go/commit/a9ec426082769f59c9d81053c350caef4eca039e))
- Fixes syntax highlighting not working ([c69fb77](https://github.com/leomoon-studios/wiki-go/commit/c69fb77ae15e478b1e841b9d03f71676b3fdc132))

## [v1.3.1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.1) (2025-03-27)

- Multiple improvements - Updates README.md - More css fixes - Toolbar buttons to be vertically centered - Toolbar buttons to hide text on 975px and below - Comment header not to wrap on mobile - Versions content table to keep max width ([d46e369](https://github.com/leomoon-studios/wiki-go/commit/d46e369a93f8bcf3c5f99a873a12fad7af1011ec))

## [v1.3.0](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.3.0) (2025-03-26)

- Adds FUNDING.yml ([242191b](https://github.com/leomoon-studios/wiki-go/commit/242191b535d66218af90bec48f5db0d234756306))
- Adds custom editor using codemirror 5.65.18 This makes adding new features easier ([3b23a77](https://github.com/leomoon-studios/wiki-go/commit/3b23a770bcf1242ebc368be36ac72a8c8116c392))
- CSS cleanup ([0c547e9](https://github.com/leomoon-studios/wiki-go/commit/0c547e91da86ca167b4264c095790a7421bee197))
- fixes emojis not rendering for subdirectory listing ([81141a4](https://github.com/leomoon-studios/wiki-go/commit/81141a4fa4fdb645a4bd6ffdf927697f1d3728df))
- improves css for search results title and close button ([e1a729d](https://github.com/leomoon-studios/wiki-go/commit/e1a729d4b55e3c94af9b9fff1b67e86448b488f0))
- Merge pull request #2 from leomoon-studios/codemirror-test2 ([9049efa](https://github.com/leomoon-studios/wiki-go/commit/9049efa968136c0c1d2d4493144f683ba885608b))

## [v1.2.2](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.2.2) (2025-03-22)

- upgrade prism to 1.30.0 ([48cba4e](https://github.com/leomoon-studios/wiki-go/commit/48cba4edda5b306fd4369fe4b6e540220be21705))
- Adds emoji support ([fe07272](https://github.com/leomoon-studios/wiki-go/commit/fe07272ea7e8cc21575e546655be8e457cc6a569))
- fixes "strings.Title is deprecated" warning ([39bbe86](https://github.com/leomoon-studios/wiki-go/commit/39bbe866a045ad298886f1168738d33995d4c815))
- updates README.md and home.go with supported markdown emojis ([d3f2360](https://github.com/leomoon-studios/wiki-go/commit/d3f2360640580b8cdc9f076e1a2f930221210aa5))
- Adds bidirectional text support to comments and fixes toolbar button css ([78c4137](https://github.com/leomoon-studios/wiki-go/commit/78c4137a0b765111c30c5ab046520cc77916d76f))
- Adds support for log and csv file types as attachments ([b745a58](https://github.com/leomoon-studios/wiki-go/commit/b745a585cd4e1503c5ed3ac6cc0a03000908d5f0))
- Merge pull request #1 from leomoon-studios/dev ([4f290d0](https://github.com/leomoon-studios/wiki-go/commit/4f290d09ad5c793a1a76b009d0073f3e5e964ae4))

## [v1.2.1](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.2.1) (2025-03-19)

- adds screenshot to README ([7e6941a](https://github.com/leomoon-studios/wiki-go/commit/7e6941a2c2c1aa5db695266d5ed967cc03cedfc9))
- minor fixes ([7f8cd17](https://github.com/leomoon-studios/wiki-go/commit/7f8cd1794caa76644e2d43b5a9cfda05d6abe71a))
- refactors mermaid there were rendering issues caused by goldmark ([588c876](https://github.com/leomoon-studios/wiki-go/commit/588c876945b426a80518a3322e6fb87ba57c00be))
- fixes comments css interfering with codeblocks syntax highlighting also improves comments css ([cc50176](https://github.com/leomoon-studios/wiki-go/commit/cc50176fdc53c5bc9eace9b2ae086a6e4670e5c6))
- Merge pull request #2 from leomoon-studios/dev ([d4bd1e7](https://github.com/leomoon-studios/wiki-go/commit/d4bd1e762a88ce1aac2382b687bde19303fc5e07))

## [v1.2.0](https://github.com/leomoon-studios/wiki-go/releases/tag/v1.2.0) (2025-03-18)

- initial public release. needs lots of css cleanup. ([f895aee](https://github.com/leomoon-studios/wiki-go/commit/f895aeec4c9a6f385f387e588d9da857f98c153c))

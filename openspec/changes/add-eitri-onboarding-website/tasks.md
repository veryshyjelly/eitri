# Tasks

Sections 1–6 record the original website work and its remaining verification. Sections 7–11 are the approved flexibility/advanced-documentation expansion and start unchecked. Capturing or validating this plan does not complete implementation. Preserve historical completion evidence, record new verification separately, and leave unavailable native-platform checks explicitly open.

## 1. Static website foundation

- [x] 1.1 Create the isolated `website/` structure, static page templates, shared content modules, and deterministic generation script; verify generation produces homepage and guide HTML in `website/dist/` without a runtime backend.
- [x] 1.2 Add shared navigation, semantic page landmarks, skip link, page titles/descriptions, favicon, and social preview metadata; verify both routes load directly and internal links resolve in a local static server.
- [x] 1.3 Add local preview/build instructions to the existing root `AGENTS.md` and ignore generated output; verify the documented commands work and no extra README or unrelated Markdown files are created.

## 2. Shared visual and accessible experience

- [x] 2.1 Implement responsive layouts, readable typography, restrained warm accents, and dark/light color tokens; verify contrast in both themes and usable layouts at 320 pixels and desktop widths.
- [x] 2.2 Implement the early theme preference read and labeled theme toggle; verify dark defaults even with a light system preference, saved choices apply before rendering across both pages, and blocked/invalid storage does not break the page.
- [x] 2.3 Implement keyboard focus styling, accessible disclosure/status behavior, contained code overflow, and reduced-motion handling; verify the controls by keyboard and with reduced motion enabled.

## 3. Homepage narrative and demonstration

- [x] 3.1 Write the homepage purpose, two primary actions, three workflow benefits, optional advanced section, FAQ, and final setup action; verify their order and wording against `website-introduction` scenarios and current Eitri behavior.
- [x] 3.2 Inspect the existing tutorial assets and prepare a readable workflow video, poster, captions, and text summary with browser/editor/Eitri labels; verify the sequence covers intake, editing, saving, testing, and results without inventing behavior.
- [x] 3.3 Implement deliberate video loading and playback controls, retaining a no-JavaScript alternative; verify keyboard playback and inspect network requests to confirm initial loading and scrolling fetch neither video nor original GIFs.

## 4. Guide content and executable examples

- [x] 4.1 Verify one accessible beginner problem supported by current URL routing and store its URL, sample input, expected output, and equivalent solutions in shared guide fixtures; verify all four solutions describe the same problem and result.
- [x] 4.2 Write prerequisite and Eitri installation/verification instructions for macOS, Linux, and Windows, including language toolchains and PATH troubleshooting; verify commands and official documentation against the current required versions and native shell conventions.
- [x] 4.3 Create complete C++ workspace/configuration/example variants with explicit parent-directory creation, no library dependencies, and manual editor opening; verify generated configurations through Eitri's validator and compile/run the example to the documented output.
- [x] 4.4 Create complete Python workspace/configuration/example variants using the same minimal setup contract; verify generated configurations through Eitri's validator and run the example to the documented output.
- [x] 4.5 Create complete Rust workspace/configuration/example variants using the same minimal setup contract; verify generated configurations through Eitri's validator and compile/run the example to the documented output.
- [x] 4.6 Create complete OCaml workspace/configuration/example variants using the same minimal setup contract; verify generated configurations through Eitri's validator and compile/run the example to the documented output.
- [x] 4.7 Add fixture checks that construct temporary workspaces and use the actual Go configuration reader/validator, plus example-output checks; verify failures detect a missing base file, invalid referenced paths, and incorrect sample output without altering Go application behavior.
- [x] 4.8 Write all seven guide stages with actions, expected outcomes, and adjacent recovery sections; verify coverage of missing commands/files, compilation failures, absent intake, and saves that do not run tests, with optional editor/library setup only after first-run success.

## 5. Guided browsing and command interactions

- [x] 5.1 Generate complete labeled guide variants into HTML and implement language/platform filtering; verify all four languages and three platforms have consistent commands, paths, code, and configuration and remain navigable without JavaScript.
- [x] 5.2 Implement `lang`/`os` URL selection with visible C++/Linux defaults, independent fallback for unsupported values, and back/forward synchronization; verify direct/shared links and navigation history preserve consistent selections.
- [x] 5.3 Implement numbered progression and the compact “Show all steps” mode; verify selected content is complete in both modes and step changes remain keyboard accessible.
- [x] 5.4 Add command/configuration copy controls and accessible feedback; verify exact copied text, genuine success feedback, rejected/unavailable clipboard handling, and manually selectable fallback text.

## 6. Integration and local delivery

- [ ] 6.1 Exercise the documented first-run workflow through real problem intake, file generation, manual editing, save-triggered testing, result inspection, and solution copy; record automated/manual evidence and the language/platform matrix, leaving unavailable checks explicitly open.
- [x] 6.2 Review the combined site for beginner clarity and factual accuracy; verify a reader can identify what runs locally, the role of each application, the setup action, and the difference between sample success and judge acceptance without consulting the full README configuration.
- [ ] 6.3 Run cross-page integration checks for theme persistence, keyboard navigation, mobile layouts, JavaScript-disabled reading, media loading, and direct guide links; verify no browser calls reach local Eitri endpoints and no backend/account is required.
- [x] 6.4 Run `go test ./...` and `go vet ./...` after any verification-harness additions, regenerate the static output, and inspect the final diff; verify Go runtime behavior and dependencies remain unchanged and generated content matches its sources.
- [x] 6.5 Deliver a working local preview and portable `website/dist/` output with a concise verification summary; verify public hosting/domain configuration has not been performed and update task completion only for implemented, verified work.

## 7. Expanded content foundation and positioning

- [x] 7.1 Use current runtime contracts and read-only research to identify Eitri’s generic execution/transformation model; keep research provenance out of public pages.
- [x] 7.2 Extend the existing static generator to produce `/setups/`, `/handbook/`, and `/reference/` with shared navigation, metadata, Monaspace, dark-first themes, tables of contents, and stable section anchors; verify all five routes load directly.
- [x] 7.3 Replace the homepage lead with “Your language. Your editor. Your workflow.”, remove the four-language hero row, update FAQ/metadata, and provide beginner/advanced actions while preserving the concrete workflow and lazy video with captions off by default.
- [x] 7.4 Label existing first-run walkthroughs as starter examples and link to setups/custom-language guidance; preserve all four languages, three OS variants, query parameters, progression, and back/forward behavior.

## 8. Concept-based language and toolchain guidance

- [x] 8.1 Explain source preparation, optional compilation and execution as the common configuration model; link to starter walkthroughs and the adaptation guide.
- [x] 8.2 Explain direct compilation with a focused compiler fragment and several illustrative languages, rather than personal configurations.
- [x] 8.3 Explain interpreted execution and optional syntax checks; show how runtime dependencies and sample timing affect configuration.
- [x] 8.4 Explain user-supplied project helpers and temporary/persistent build workspaces without publishing personal scripts, package pins or setup instructions.
- [x] 8.5 Explain operating-system, architecture, toolchain and judge constraints as independent configuration choices; omit personal workspace repair notes.
- [x] 8.6 Render Languages & toolchains with stable concept anchors, copyable fragments and complete static content; verify no research-workspace catalog or source disclosures are published.

## 9. Pipeline handbook and adaptation guide

- [x] 9.1 Write the full intake-to-submission pipeline and illustrate editable source, configuration, and generated output using small self-contained examples; explain template.modifier versus code.modifier without implying browser-side execution.
- [x] 9.2 Document snippet matching/dependency order, header/code blocks, .Header/.LibFiles, and template helpers; verify representative transformation outputs against current Eitri behavior.
- [x] 9.3 Explain temporary build workspace contents, command/path semantics, dependency provisioning, custom build helpers, and compiled_file submission artifacts; link the explanation to relevant language setups.
- [x] 9.4 Add `/handbook/#adapt-your-language` covering templates/rules, source transformation, optional compilation, execution, dependencies, validation, sample tests, and submission-output inspection.
- [x] 9.5 Add task-based editor integration, judge filename rules, and troubleshooting; distinguish OS/toolchain constraints from judge intake and submission support, and recheck cp-submit compatibility claims against current source.

## 10. Complete configuration reference

- [x] 10.1 Inventory every current YAML field and write anchored reference entries with type, purpose, required/optional status, actual defaults, path/command behavior, and examples; compare coverage against configuration structs and validation.
- [x] 10.2 Document each template context and registered helper with available values, usage examples, and limitations; cross-link reference, handbook, and concept explanations and use Context7 for external toolchain/API syntax.

## 11. Expansion validation and local delivery

- [x] 11.1 Validate internal configuration fixtures and self-contained source transformations in temporary workspaces, keeping research fixtures outside generated site output.
- [x] 11.2 Verify all five routes, concept anchors, reference coverage, existing guide URLs and exclusion of research-workspace content/assets from the generated site.
- [ ] 11.3 Review all pages in both themes at mobile/desktop widths, with keyboard access, JavaScript disabled, and reduced motion; verify theme persistence, readable code, lazy media/captions defaults, and absence of local Eitri API calls.
- [x] 11.4 Update the existing AGENTS.md website instructions for new routes/content/tests, run website checks and appropriate Go checks for any harness edits, regenerate portable output, and record expansion evidence separately; create no other Markdown files and perform no deployment or runtime/sibling-repository changes.

## 12. Authorized Cloudflare publication

- [x] 12.1 Configure an assets-only Cloudflare deployment for eitri.veryshyjelly.in, production metadata, sitemap/robots, custom 404 and fixture exclusion; verify the production build and Wrangler dry run.
- [x] 12.2 Verify the existing Cloudflare account/zone and absence of an existing Eitri domain binding; publish and attach only the requested custom domain.
- [x] 12.3 Verify live HTTPS pages and assets, directory redirects, 404 behavior and fixture exclusion; record deployment identity and browser playback evidence.

## Implementation evidence (original website; expansion not yet implemented)

- Static source is in `website/`; `node website/build.mjs` generates two pages and twelve complete guide variants in ignored `website/dist/`. `npm --prefix website run dev` starts the loopback preview.
- `npm --prefix website test` passed: seven Node tests cover guide coverage, URL fallback, static/no-JavaScript content availability, deferred media, storage failure, theme preference, and clipboard success/rejection. These interaction tests use controlled DOM doubles; they do not substitute for all browser accessibility checks.
- All twelve YAML configurations pass the actual Go reader/validator, including negative checks for missing template and library paths. C++, Python, Rust, and OCaml each passed the official sample and three boundary cases on macOS.
- `python3 website/tests/smoke.py` passed with local network permission: the real CLI received synthetic Competitive Companion JSON over HTTP, routed the filename, created the file, observed a save, ran the sample, and rendered AC. Browser extension clicking, native Windows/Linux execution, and the CLI desktop clipboard remain unverified, so task 6.1 remains open.
- Manual Zen browser review verified the dark homepage, light-theme persistence into the guide, Rust selection and matching URL, all-steps/one-step switching, successful website copy feedback, 320-pixel guide layout, and playing video with captions and native controls. In-app browser access was unavailable; native browser review was used.
- Blocked browser storage and clipboard failure are covered by controlled interaction tests. Full manual JavaScript-disabled, reduced-motion, network-panel, screen-reader, and cross-browser review remains pending under task 6.3. Static HTML includes all guide content; no video source exists before playback and site code makes no local Eitri API calls.
- `go test ./...`, `go vet ./...`, JS syntax checks, and strict OpenSpec validation passed. Existing Go runtime source and dependencies are unchanged; the new Go file is a standalone website-fixture verifier.
- Original GIFs are preserved. The site uses a roughly 1.2 MB MP4 and 93 KB JPEG poster. The recording is an existing advanced AtCoder workflow; the first-run guide consistently uses Codeforces Watermelon.
- The local preview was reviewed; portable output is available in `website/dist/`. Public deployment and domain setup have not been performed. Production canonical/absolute social-image URLs remain deployment-time configuration.

- Typography follow-up: the user requested Monaspace throughout. Applied self-hosted variable Monaspace Neon to all site text/controls/code/captions, preloaded it on both pages, included the upstream SIL license, adjusted heading sizes for its metrics, and outlined social-preview lettering in Monaspace. Build and seven website checks pass. A fresh visual browser pass was attempted but native browser automation repeatedly timed out; visual verification of this follow-up remains pending under 6.3.

- Submission-extension follow-up: added optional cp-submit guidance to advanced homepage content and the final guide step, covering CSES/AtCoder/Codeforces, installation, login/open-browser prerequisites, the c trigger, polling, verdict checks, compatibility, and manual fallback. Corrected the FAQ and benefit copy so they no longer imply that extensions cannot automate submission. No extension was installed and no real submission was made. Current source compatibility gaps are recorded in design.md and disclosed in the optional guide.


## Expansion implementation evidence (initial catalog approach superseded below)

- Implemented all five static routes with shared Monaspace typography/navigation/theme behavior, the generic-workflow homepage positioning, twelve annotated setups, a pipeline handbook/custom-language guide, and a configuration reference. All content is generated into static HTML; existing guide selectors/query URLs and captions-off default are preserved.
- Recipe provenance is pinned per branch in website/data/setups.json. Configurations were adapted only in website content: corrected ARM's template path and LLVM's output name, removed unsupported historical keys, and explicitly selected Nim source libraries to avoid the branch's cpio/cpio.nim duplicate basename. cp_x remains clean and personal scripts are unchanged. Helper examples use portable paths.
- npm --prefix website test passes ten Node checks, twelve starter configuration fixtures, twelve recipe/path-manifest checks, three real Eitri transformation/output checks with C++/Rust/LLVM compilation and execution, and four starter programs with sample/boundary cases on macOS. Recipe validation uses placeholder library files and is explicitly labeled as such; full personal workflows and unavailable native platforms are not claimed verified.
- Documentation checks verify every internal route/anchor/asset link, all twelve recipe/source entries, all twenty-four YAML schema entries, and all ten registered helper names. Generated HTML nesting/balanced-tag inspection passed for all five pages.
- go test ./... and go vet ./... pass. The real macOS/Python HTTP/PTY integration smoke was rerun successfully for intake, routing, source creation, save-triggered testing, and AC display. Browser extension clicking, real judge submission, native Windows/Linux flows, and CLI desktop clipboard remain outside that evidence; 6.1 stays open.
- Manual Zen review confirmed the updated dark homepage, setup navigation, handbook/light-theme persistence, reference navigation at a 320-pixel responsive viewport, existing Rust/macOS shared guide selection, clipboard success feedback, and keyboard focus on the copied code block. Responsive mode was closed and the website restored to dark afterward. Reused the existing local preview on port 4173; no user server was stopped.
- Complete manual coverage of every page in both themes, JavaScript-disabled browser mode, reduced motion, network-panel inspection, and cross-browser/screen-reader review remains pending under 11.3 and 6.3. Static content availability, guarded storage/copy behavior, deferred video, and absence of localhost requests in website interactions are covered by source/automated checks.
- Context7 documentation was consulted for Go templates and Gleam module execution. The current cp-submit source still expects a Codeforces languageId absent from Eitri's payload; optional compatibility guidance and manual fallback remain visible. Fixed the first-run scenario's unconditional manual-submission wording.
- Local deliverable is website/dist/. No Go runtime source/dependencies, sibling repositories, or public hosting were changed. No additional Markdown files were created.


## User clarification: explain Eitri, not research workspaces

The user clarified that cp_x and personal scripts were provided to inform understanding, not to be published as particular setups. This clarification supersedes the catalog approach and its historical evidence above. Sections 7–11 and the active capability requirements now describe concept-based language/toolchain guidance.

- Replaced the twelve-workspace catalog with explanations of direct compilation, interpreted execution, project helpers, source transformation and portability. Languages illustrate capabilities; no fixed language list is implied.
- Removed public repository attribution, source links/revisions, clone instructions, personal helper scripts and workspace-specific repair/verification disclosures. Preserved the deeper pipeline handbook and complete reference, with self-contained explanatory examples.
- Moved research-derived configurations into internal test fixtures; the build no longer publishes them or personal scripts, and clears generated output before rebuilding to remove obsolete assets.
- Added checks for concept anchors and absence of research-workspace material, alongside existing link, schema/helper and executable-example verification. Monaspace and captions-off behavior are preserved.

- Correction verification: the complete website suite passes (ten Node checks, starter/internal configuration validation, three transformation compile/run checks and four starter programs). All public pages contain no research-workspace references or personal helpers; obsolete catalog assets are absent. Strict OpenSpec validation passes.

- Source-shaping clarity follow-up: replaced abstract capability claims with two concrete stages. Shows boilerplate plus template.modifier producing an editable file, then editable Rust/helper files plus lib/code configuration producing one prepared source. Explains marker extraction, matching, defaults, test/copy timing and preservation of editable files. The prepared Rust example shares its source/output fixture with real Eitri transformation and compile/run verification.


## Cloudflare publication evidence

- The user explicitly authorized public deployment at https://eitri.veryshyjelly.in/. Used the existing authenticated Cloudflare account and active veryshyjelly.in zone, with no prior Eitri Worker/custom domain/DNS record found during inspection.
- Deployed the assets-only Worker eitri-website with Wrangler 4.162.0. Version: 28ae032d-d62b-4993-88ee-dd4bc87e6d07. Cloudflare reports the custom domain enabled in production, with certificate provisioning; workers.dev and preview URLs are disabled.
- Website checks pass: eleven Node tests plus all existing configuration, transformation/compilation and starter-program fixtures. Wrangler dry run and strict OpenSpec validation pass. Deployment source/config is website/wrangler.jsonc; npm --prefix website run deploy rebuilds and invokes pinned Wrangler.
- HTTPS verification retrieved all five pages and confirmed byte equality with generated HTML plus correct canonical URLs. Verified stylesheet, JavaScript, Monaspace font, social artwork, captions, sitemap and video bytes. The video endpoint returned its full body (200) for a range request; native Zen playback was verified with loaded media and active Pause control. No byte-range support is claimed by this check.
- Verified /handbook redirects to /handbook/ and unknown routes, starter fixture JSON, old recipe-fixture JSON and internal test fixture paths return the custom 404. Existing Cloudflare-managed crawler policy is prepended to robots.txt; the site's sitemap directive remains present and account policy was unchanged.
- Opened the production homepage in Zen and verified correct content/navigation and native workflow playback. No unrelated hostnames, sibling repositories or Go runtime source were changed. Remaining earlier manual/native-platform checks are still tracked separately.


## 13. Queued reference documentation: HTTP test cases

The user requested that HTTP test-case documentation live in the site's reference, not the README. This is queued work; no website implementation or deployment is included in the README update.

- [ ] 13.1 Add an HTTP reference section at `/reference/#http-tests` for `POST http://127.0.0.1:6174/tests`, verified against `server/server.go` and its tests. Include a copyable curl example, required string fields `input` and `output` (empty strings allowed), one case per request, and the requirement for an active problem.
- [ ] 13.2 Document response and lifecycle semantics: `201 {"added":true}` after installation in the TUI; `400` invalid input, `405` wrong method, `409` no problem, `413` body over 4 MiB, and `503` timeout/shutdown. Explain the five-second processing deadline, cancellation of an active run, verdict reset to NA, manual rerun with `r`, in-memory lifetime, and that requests expiring before processing do not add a case. Confirm all details against current source before publication.
- [ ] 13.3 Check the reference anchor, example and response descriptions against endpoint tests, run website checks, then publish and verify the new reference section on the live site.

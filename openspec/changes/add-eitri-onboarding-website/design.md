# Design

## Context

See `proposal.md` for motivation and the three capability specs for observable requirements. The initial local website already provides a homepage and four beginner walkthroughs. The approved expansion corrects its impression of a fixed language list and its shallow treatment of advanced workflows. Eitri delegates build/run behavior to configuration; the varied `../cp_x` branches informed our understanding of this abstraction, but must not become a public setup catalog. Derive runtime claims from the current working tree, which has changed since the initial implementation.

The audience includes newcomers and experienced competitors. Retain the four guided starter languages and add concept-based language/toolchain explanations, a dedicated handbook, and a complete reference. Preserve Monaspace, dark-first themes, and captions off by default. Only local delivery is in scope. Maintain this change's existing OpenSpec artifacts; create no unrelated Markdown files.

## Goals / Non-Goals

**Goals:** Explain language/editor independence and the practical workflow; preserve an approachable first run; teach advanced configuration through concrete transformations and a complete reference; keep pages accessible, fast, portable, and verifiable.

**Non-Goals:** Change the Go application, sibling repositories, or personal helper scripts; generate downloadable workspace bundles; add a documentation search service; execute user code in the browser; integrate with localhost APIs; add accounts or analytics; or publish publicly.

## Decisions

### Static source with progressive enhancement

Keep the website in `website/` using HTML, CSS, and small JavaScript modules. Serve `/`, `/get-started/`, `/setups/`, `/handbook/`, and `/reference/` as real static pages without client-side routing or a runtime backend. Retain the dependency-free Node generator.

Store guide variants in structured JavaScript data keyed by language and platform. A small Node-based static generation script will emit complete HTML into `website/dist/` from those data and page templates, copying styles, scripts, and assets. Node is a development dependency only; the output needs only static hosting. Generate content during preparation, not on page load, so the no-JavaScript experience remains complete without manually maintaining duplicated instructions. Ignore generated output and document preview commands in the existing root `AGENTS.md`, not a new README.

### Information hierarchy and visual language

Use the exact lead “Your language. Your editor. Your workflow.” Follow it with a plain explanation of Eitri as a local terminal tool and the problem → edit → save → test loop. Remove the four-language hero row and fixed-list implications from the FAQ and metadata. Order sections as purpose/actions, demonstration, practical benefits, concise customization preview, common questions, and final setup action. Provide Get started, Languages & toolchains, How Eitri works, Configuration reference, and GitHub navigation; the logo links home. The primary hero action is Get started and the secondary is How Eitri works, with the video in the next section.

Describe language independence, editor choice, operating-system/toolchain constraints, judge intake, and submission separately. Named languages are examples, not a closed support list. Judge intake requires Competitive Companion support and suitable filename rules; submission depends on judge capabilities and optional compatible extensions. Do not imply every toolchain runs on every operating system or every judge accepts every language.

Use generous spacing, self-hosted Monaspace Neon throughout headings, body text, controls, captions, and code blocks, restrained warm accents, and authentic application imagery. Dark mode uses a charcoal background; light mode uses a warm off-white background. Use the official variable WOFF2 font with its SIL Open Font License, preload it on every page, and retain a monospace fallback. Define shared color tokens for both themes and validate text, controls, and code contrast. Avoid ornamental terminal animations and repeated generic feature cards.

### Guide selection and progression

Label the four walkthroughs as starter examples and link to `/setups/` and `/handbook/#adapt-your-language`. Preserve existing guide URLs and selection behavior; use concept explanations to cover other toolchains without multiplying the first-run matrix.

Support language values `cpp`, `python`, `rust`, and `ocaml`, and platform values `macos`, `linux`, and `windows`. Use `lang` and `os` URL parameters, defaulting to `cpp` and `linux` with the selectors clearly visible; do not silently infer the operating system. Unknown values fall back independently to those defaults. Preserve valid selections on navigation and handle browser back/forward changes.

Render all labeled variants into static HTML. JavaScript filters to the selected variant and enables the numbered stepper; “Show all steps” reveals the entire active variant. Without JavaScript, ordinary section links and grouped language/platform headings provide access to all instructions. Keep selection and display code separate from instructional content.

Each of the seven stages includes an action, expected outcome, and expandable recovery advice: prerequisites; Eitri install/verification; workspace preparation; minimal configuration; start/intake; edit/save/results; copy/manual submission. Use complete examples rather than leaving placeholders that cause startup validation to fail.

### Minimal setup and example correctness

Use one problem consistently across the four language guides. During content preparation, select and verify an accessible beginner problem supported by the existing URL rules; retain its URL, input, expected output, and equivalent solutions in the guide data and test fixtures. The exact problem is editorial content, not a new application contract.

Create the workspace and parent directories explicitly. Use an existing base source file, `template.modifier` and `code.modifier` that preserve `.Code`, no library include entries, `editor: ""`, `create_file: true`, and `run_on_save: true`. Explain manual file opening before offering optional editor automation. Do not copy starter templates wholesale: they reference personal library directories and external editor commands.

Adapt compiler commands, executable suffixes, interpreter commands, and shell examples to each platform. Base the Go requirement on `go.mod`, not the outdated README minimum. Document Go binary PATH troubleshooting and link verified official toolchain installation instructions. Verify current library/CLI guidance with Context7. Native Windows instructions must not silently substitute Unix shell commands or WSL workflows.

Keep install/verification commands, directory trees, config blocks, and example solutions in the same content model. Configuration fixtures must be tested through the actual `config.ReadConfig`/`config.Validate` implementation. Example programs must produce the documented output. Do not change application validation to accommodate incorrect website examples.

### Demonstration and asset loading

Inspect the existing GIFs, extract only relevant browser/editor/terminal sequences, and create optimized video, a poster, captions, and an adjacent text summary. Preserve the original assets. Label the applications and explain the transitions; do not fabricate product capabilities or present a simulated interaction as a real recording.

Initially render a poster and playback button without a video source request. Attach the video source only after activation, then expose native playback controls with captions available but off by default. Provide a standard media link/text alternative without JavaScript. This avoids initial GIF/video downloads and supports reduced motion without removing explanatory content.

### Theme, accessibility, and copy behavior

Set the default document theme to dark in HTML/CSS. Run a minimal guarded preference read before rendering page content, accepting only `dark` or `light` from local storage. Do not use the system color preference to override the chosen dark default. A labeled toggle updates both the document and stored preference when storage works; storage failure must not break the page.

Use semantic sections, a skip link, visible focus, sufficiently large controls, and native disclosure elements. Disable decorative transitions for reduced motion. Keep code overflow inside its block. Copy actions use browser clipboard APIs with accessible success/failure feedback and selectable text fallback; report success only after the operation succeeds.

### Languages and toolchains: explain the abstraction

The user clarified that cp_x and personal helper scripts were research material for understanding Eitri, not content to publish. This clarification supersedes the earlier twelve-recipe approach. Keep `/setups/` as a stable route, titled Languages & toolchains, and explain the tool through execution, direct compilation, interpreted runtimes, project build helpers, source transformation, and portability.

Use small configuration fragments to show compiler.source, optional compiler.compile/args, and compiler.run. Explain that a helper is a user-supplied executable and how it can provide dependencies or project structure in the temporary workspace. Do not reproduce personal helper scripts, package pins, setup-specific limitations, clone commands, source revisions, workspace repair notes or links to cp_x. Language names illustrate capabilities rather than enumerate fixed integrations. Retain generic editor/OS/judge constraints where they explain a real configuration decision.

Link readers to the full starter walkthroughs, the adaptation handbook and individual reference fields. All explanations and examples remain readable without JavaScript. Internal regression fixtures may retain research-derived configurations and provenance, but normal builds must not publish those fixtures or depend on the sibling checkout. Clean generated output before building so removed catalog assets cannot remain accessible.

### Handbook and complete configuration reference

Generate `/handbook/` with stable section links and a table of contents. Trace intake → filename selection → editable source creation → snippet resolution/header extraction → code transformation → temporary compiler input → compilation/execution → sample results → submission preparation. Use small self-contained source/configuration/generated-output examples; do not add a browser compiler or live execution engine.

Explain `template.modifier` versus `code.modifier`, library matching/dependency ordering, `@head`/`@code` blocks, `.Header`/`.LibFiles`, template helpers, and editable files versus generated compiler input. The temporary build workspace receives generated source, not an automatic copy of the entire project; explain how custom scripts supply project structure/dependencies and how `compiled_file` selects a generated submission artifact. Cover manual/automatic editor opening, judge URL rules, command/path troubleshooting, sample limitations, and cp-submit guidance rechecked against current source.

Provide `/handbook/#adapt-your-language`: choose an editable template and filename rules, define generated source and transformations, configure optional compilation and execution, supply dependencies/helpers, validate the workspace, test samples, and inspect submission output. Link each step to concept explanations and reference entries. Beginners can finish their existing walkthrough without this material.

Generate `/reference/` with an anchored entry for every current YAML field, all available template contexts, and registered helper functions. Document type, purpose, required/optional status, actual defaults, relevant context, path base/command semantics, and a small example. Derive claims from current validation, rendering, linking, editor launch, execution, and submission code; use Context7 for external syntax/toolchain documentation. Do not infer supported keys from old sample configurations. Share content/fixtures where useful to reduce drift.

## Risks / Trade-offs

- **Guide drift from application behavior** → Generate examples from shared data, validate fixtures against Go configuration code, and check claims against working-tree source.
- **Four languages across three platforms increase verification work** → Share the narrative while keeping command variants explicit; report the actual test matrix and never label an untested platform as verified.
- **Required compilers or target platforms are unavailable locally** → Run available automated checks, record remaining native/manual checks in the task checklist, and keep affected verification tasks open rather than claiming completion.
- **A static generator adds a development step** → Keep it dependency-light and deterministic; deployable output remains plain files with no runtime Node requirement.
- **Long no-JavaScript guide** → Provide a table of contents and clearly labeled variant groups; enhanced browsing shows one variant at a time.
- **Advanced material overwhelms the homepage** → Keep the home narrative concise and provide depth in directly linked setup, handbook, and reference pages.
- **Personal setups depend on hidden local state** → Explain helpers/prerequisites, normalize examples, and state exact verification evidence rather than universal portability claims.
- **Video conversion can lose readability** → Verify terminal text at target display sizes, provide captions/text equivalents, and ensure initial page load requests only the poster.

## Verification and Delivery

For the expansion, verify all five routes/section links, concept coverage and the absence of research-workspace content, complete reference coverage, representative generated-output fixtures, and existing guide URL compatibility. Exercise runnable examples in temporary workspaces with available toolchains; record unavailable compilers/platforms separately from failures. Extend website verification only. Historical completion evidence does not establish completion of the expansion.

Validate generated configurations and workspace fixtures; compile/run the example in each language; exercise the actual intake, file generation, save-triggered test, and copy flow. Use temporary fixtures and leave application behavior unchanged. Record automated versus manual evidence in this change's tasks.

Check both themes, theme persistence and blocked storage, valid/invalid query parameters, back/forward navigation, clipboard success/failure, JavaScript-disabled browsing, keyboard access, reduced motion, direct guide loading, and 320-pixel layouts. Inspect network activity to confirm no initial video/GIF request and no localhost Eitri calls.

Deliver the local preview and `website/dist/` as portable static output. Include titles, descriptions, favicon, and social preview metadata. Public hosting, canonical production URL, domain setup, and production social-image URL configuration are deferred to a separately requested deployment. No migration is required; rollback consists of reverting the isolated website additions.


## Optional submission extension follow-up

The user identified `https://github.com/veryshyjelly/cp-submit` as the optional automatic-submission extension for CSES, AtCoder, and Codeforces. Keep it outside prerequisites for the first run. Add a shared disclosure in advanced homepage content and at the final guide stage; distinguish Competitive Companion intake from cp-submit submission and explain browser/login prerequisites, the c trigger, roughly three-second polling, judge-side result checking, and manual fallback.

Source inspection confirms polling of `http://localhost:6174/submit` and handlers for the three judges. Current Eitri queues a submission when c is pressed. Current Codeforces extension code expects `languageId` and derives problem identifiers from `problemName`; Eitri sends no `languageId` and sends the human-readable title. This is a known compatibility gap, not a verified automatic-submission workflow. Link to project installation guidance rather than asserting that older browser-store builds match the fork. Application/extension repairs and real judge submissions are outside this website-copy update.


## Authorized Cloudflare publication follow-up

The user subsequently requested publication to eitri.veryshyjelly.in through their Cloudflare account. This explicit request supersedes earlier local-only/deferred-deployment statements for that hostname. Use an assets-only Cloudflare Worker with a custom-domain route, automatic DNS/certificate provisioning, directory-index URLs and a custom 404. Do not add application runtime code, bindings or unrelated DNS changes. Disable workers.dev/preview URLs. Generate production canonical/social URLs, robots.txt and sitemap.xml; exclude fixture JSON with .assetsignore. Keep deployment settings in website/wrangler.jsonc and document build/deploy commands in AGENTS.md. Verify live HTTPS routes, assets, redirects and fixture exclusion after deployment.

# Proposal

## Why

Eitri's configuration can overwhelm newcomers, while the initial website's four-language presentation understates the tool's flexibility. Eitri delegates compilation and execution to configured commands and adapts to different languages, editors, and workflows. The website must explain the everyday benefit, guide a first successful sample test, and teach the deeper configuration model using focused examples of the tool’s configuration model.

## What Changes

- Provide five static pages: homepage, Get started, Languages & toolchains, How Eitri works, and Configuration reference.
- Lead with “Your language. Your editor. Your workflow.”, explain the problem → edit → save → test loop, and retain the click-to-play demonstration with captions available but off by default.
- Present C++, Python, Rust, and OCaml as starter walkthroughs rather than a language-support boundary; preserve their complete platform-specific instructions and existing guide URLs.
- Explain compiled, interpreted and project-based execution, source transformation, editor choice and portability. Use languages as illustrations of these capabilities; do not publish a catalog of personal workspaces, repository provenance, clone commands or helper scripts.
- Add a dedicated handbook explaining the source/build/submission pipeline, language-specific transformations, snippets, custom toolchains, editor integration, and judge filename rules, including an “Adapt your own language” guide.
- Add a complete reference for current YAML fields, template contexts/functions, defaults, path rules, and command behavior derived from the working-tree implementation.
- Support dark and light themes, defaulting to dark on first visit and remembering explicit preferences.
- Keep optional complexity outside the first-run path while making advanced material directly accessible from navigation.
- Provide responsive, keyboard-accessible pages, readable no-JavaScript content, shareable guide selections, and copyable commands.
- Deliver a local preview and portable static output. The user subsequently authorized publication through their Cloudflare account at eitri.veryshyjelly.in.

## Capabilities

### New Capabilities

- `website-introduction`: Explain Eitri's flexibility through an authentic demonstration, accurate claims, and distinct beginner and advanced paths.
- `guided-first-run`: Provide four starter walkthroughs, language-independent toolchain guidance, an adaptation handbook, and a complete configuration reference.
- `website-experience`: Provide accessible navigation across five static pages, responsive presentation, theme preferences, and portable delivery.

### Modified Capabilities

None. No existing capability specs are present, and this change does not alter Eitri's CLI or runtime behavior.

## Impact

- New website source and assets live in `website/`; guide fixtures and verification code belong with the website.
- Use static HTML, CSS, and small JavaScript modules without a frontend framework or backend. No new Go runtime dependencies, database, accounts, or hosted code execution.
- Reuse and optimize existing demonstration assets without replacing the original repository GIFs.
- Guide examples must remain compatible with the actual configuration validator and execution behavior, including required directories, toolchains, and path handling.
- Treat `../cp_x` and local helper scripts as read-only research material for understanding Eitri. They are not website content. Keep public examples small and self-contained without changing sibling repositories, personal scripts, or the Go runtime. Distinguish configuration validation from compilation and end-to-end verification.
- The user explicitly authorized this change's OpenSpec Markdown artifacts as an exception to the earlier no-additional-Markdown restriction. Other Markdown generation remains outside scope. Cloudflare publication at the explicitly requested hostname is now authorized.

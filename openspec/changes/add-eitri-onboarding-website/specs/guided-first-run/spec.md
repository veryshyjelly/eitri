# Spec Delta

## Purpose

Help newcomers complete a first sample test and experienced users adapt Eitri through starter walkthroughs, language-independent toolchain explanations, a pipeline handbook, and a complete configuration reference.

## ADDED Requirements

### Requirement: Select a complete language and platform walkthrough

The getting-started page SHALL provide complete starter walkthroughs for C++, Python, Rust, and OCaml with operating-system-specific instructions. It SHALL label these as examples and link to the language/toolchain guide and custom-language guide. Existing guide URLs and selection behavior SHALL remain compatible. With JavaScript available it SHALL show the selected variant, provide a numbered progression, and offer a compact “Show all steps” view.

#### Scenario: Choose a language and operating system
- **WHEN** a visitor selects a supported language and operating system
- **THEN** prerequisite instructions, commands, filenames, configuration, and example code consistently match those selections
- **AND** unrelated variants do not clutter the active walkthrough

#### Scenario: Experienced visitor scans the guide
- **WHEN** a visitor selects “Show all steps”
- **THEN** the complete selected walkthrough is available in one compact view without stepping through each stage

### Requirement: Share guide selections

Guide language and platform selections SHALL be represented in the URL and restored when that URL is opened. Missing or unsupported values SHALL resolve to visible, valid defaults rather than broken or mismatched instructions.

#### Scenario: Open a shared guide
- **WHEN** a visitor opens a link containing supported language and platform selections
- **THEN** the guide displays those selections and their matching instructions

#### Scenario: Open an invalid guide link
- **WHEN** a guide URL contains unknown selection values
- **THEN** the page remains usable and displays its valid fallback selections

### Requirement: Guide installation and workspace preparation

The walkthrough SHALL cover prerequisites, Eitri installation and command verification, workspace creation, and a minimal configuration before starting Eitri. It SHALL explicitly create all directories and base source files needed by that configuration.

#### Scenario: Follow setup in a fresh workspace
- **WHEN** a visitor with the listed prerequisites follows the preparation steps
- **THEN** the resulting configuration passes Eitri's validator and all referenced workspace paths exist
- **AND** the guide explains how to recognize a successful installation

### Requirement: Avoid optional dependencies in the first run

The initial configuration SHALL require neither a personal snippet library nor automatic editor launching. The walkthrough SHALL tell visitors which generated file to open in their preferred editor and introduce editor automation and snippet linking only after first-run success.

#### Scenario: Start without a configured editor or library
- **WHEN** a visitor follows the minimal setup without a personal library or editor executable in PATH
- **THEN** they can still receive a problem, open its file manually, and run sample tests

### Requirement: Reach a successful sample test

Every language walkthrough SHALL use the same example problem and cover starting Eitri, sending the problem through Competitive Companion, editing the generated file, saving to test, inspecting expected results, and copying the prepared solution for manual submission, followed by optional cp-submit guidance.

#### Scenario: Complete the first-run path
- **WHEN** a visitor follows a language walkthrough with its listed prerequisites
- **THEN** they can compare the actual result with the documented expected result
- **AND** they can copy the solution and understand that submission to the judge is separate from local testing and can be manual or automated through a compatible optional extension

### Requirement: Explain actions, outcomes, and recovery

Each walkthrough step SHALL explain its action and expected outcome and provide a short expandable recovery section. Troubleshooting SHALL cover unavailable commands, missing files, compiler failures, absent incoming problems, and saves that do not trigger tests.

#### Scenario: Recover from a setup problem
- **WHEN** the expected outcome of a step does not occur
- **THEN** the visitor can open troubleshooting adjacent to that step and find concrete checks relevant to their selected environment

### Requirement: Copy commands with usable fallback

Command and configuration blocks SHALL support copying with success or failure feedback. Their text SHALL remain selectable when clipboard access is denied or unavailable.

#### Scenario: Clipboard access works
- **WHEN** a visitor activates a copy button and the clipboard operation succeeds
- **THEN** the selected block is copied exactly and accessible success feedback is presented

#### Scenario: Clipboard access fails
- **WHEN** the browser rejects or cannot perform a clipboard operation
- **THEN** the page does not claim success and provides feedback allowing the visitor to select and copy the text manually

### Requirement: Explain language and toolchain configuration through concepts

The website SHALL provide `/setups/`, titled Languages & toolchains, explaining direct compilation, interpreted execution, project build helpers, source transformation, and editor/platform/judge constraints. It SHALL use small illustrative configuration fragments and multiple language examples to explain Eitri's generic configuration model without implying a fixed language list.

#### Scenario: Understand how another language fits

- **WHEN** a visitor reads the language/toolchain page
- **THEN** they can identify generated source, optional build steps, execution commands and dependency requirements
- **AND** they can follow links to the starter guide, adaptation handbook and reference

### Requirement: Keep research workspaces out of public content

The website SHALL explain Eitri rather than publish particular personal setups. Research repositories and scripts SHALL remain internal guidance. Generated output SHALL NOT include cp_x attribution, workspace catalogs, branch/revision disclosures, clone instructions, personal build scripts, or research configuration fixtures.

#### Scenario: Browse language guidance

- **WHEN** a visitor reads any page or opens a published asset
- **THEN** the content teaches tool behavior and configuration choices using self-contained examples
- **AND** understanding or using the website does not require a personal workspace repository

### Requirement: Explain the transformation and execution pipeline

The website SHALL provide `/handbook/` tracing intake, filename selection, editable-source creation, snippet/header extraction, code transformation, temporary compiler input, compilation/execution, sample results, and submission preparation. It SHALL include source/configuration/generated-output self-contained examples with contrasting language-specific transformations.

#### Scenario: Understand source preparation

- **WHEN** a reader follows handbook examples
- **THEN** they can distinguish template.modifier from code.modifier, editable source from compiler input, and local snippets from the final solution
- **AND** they understand library matching/dependency order, @head/@code blocks, .Header, .LibFiles, and relevant template helpers

#### Scenario: Configure a nontrivial build

- **WHEN** a reader examines build and submission stages
- **THEN** they understand which files Eitri places in its temporary workspace, how helpers supply dependencies, and what compiled_file selects
- **AND** the examples do not imply the full project is automatically copied into that workspace

#### Scenario: Understand source transformation through a complete example

- **WHEN** a visitor reads the source-shaping explanation
- **THEN** they see the editable source, local helper, relevant matching/output configuration, and prepared output together
- **AND** the explanation states when template.modifier and code.modifier run, what .Code means in each stage, and that preparing output does not rewrite editable files
- **AND** it explains the default behavior and that snippet detection follows configured rules rather than inferred language imports

### Requirement: Guide adaptation of an existing workflow

The handbook SHALL provide `/handbook/#adapt-your-language` covering templates/filename rules, generated source, optional compilation, execution, dependencies/helpers, validation, sample testing, and submission-output inspection. It SHALL also explain manual/automatic editor opening, additional judge rules, troubleshooting, and optional cp-submit integration.

#### Scenario: Add an unlisted language

- **WHEN** a visitor follows the adaptation guide for their toolchain
- **THEN** they can map source, build/run commands, and dependencies to Eitri settings with links to concept explanations and reference entries
- **AND** they are not instructed to add a built-in language implementation to Eitri

### Requirement: Provide a complete implementation-grounded reference

The website SHALL provide `/reference/` documenting every current YAML field, available template context, and registered helper. Entries SHALL include types, purposes, required/optional status, actual defaults, relevant context, path/command semantics, and examples. Claims SHALL match current working-tree behavior rather than unsupported historical keys.

#### Scenario: Look up a configuration detail

- **WHEN** a visitor opens an anchored reference entry
- **THEN** they can determine where the setting applies and how to use it correctly
- **AND** links connect it to relevant handbook explanations and configuration concepts

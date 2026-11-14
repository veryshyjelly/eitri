# Spec Delta

## Purpose

Help contest newcomers and experienced competitors understand Eitri's local workflow and choose a clear next step without first learning its configuration system.

## ADDED Requirements

### Requirement: Explain the tool before its configuration

The homepage SHALL lead with “Your language. Your editor. Your workflow.” and explain the problem → edit → save → test loop, with Eitri running locally alongside an editor and Competitive Companion. It SHALL provide prominent “Get started” and “How Eitri works” actions, an accessible GitHub link, and a demonstration in the following section.

#### Scenario: First visit
- **WHEN** a visitor opens the homepage
- **THEN** the first content explains the local workflow and its benefit without requiring knowledge of YAML, snippet linking, or compiler configuration
- **AND** the visitor can navigate to setup or the demonstration

### Requirement: Demonstrate the complete everyday workflow

The homepage SHALL provide an authentic, captioned demonstration showing problem intake, solution editing, save-triggered tests, and displayed results. Browser, editor, and Eitri SHALL be clearly identified.

#### Scenario: Watch the demonstration
- **WHEN** a visitor starts the demonstration
- **THEN** the sequence explains how the three applications work together
- **AND** captions and accompanying text convey the workflow without relying on audio

### Requirement: Load demonstration media on demand

The demonstration SHALL initially show a poster and playback control. Captions SHALL be off by default and available through the player controls. It SHALL provide playback controls and SHALL NOT download the demonstration video or original large GIFs during initial page load.

#### Scenario: Browse without playing
- **WHEN** a visitor loads or scrolls through the homepage without starting playback
- **THEN** the poster and explanatory text are available without downloading demonstration video or original GIF payloads

#### Scenario: Control playback
- **WHEN** a visitor starts the video
- **THEN** they can pause and resume it using keyboard-accessible controls

### Requirement: Progress from benefits to optional complexity

The homepage SHALL connect the demonstration to bringing sample tests into a workspace, testing on save, and copying a prepared solution. It SHALL introduce customization after the benefits, link to deeper examples and explanations, and end with a setup action. Advanced material SHALL be directly accessible through navigation without completing the beginner walkthrough.

#### Scenario: Follow the beginner narrative
- **WHEN** a visitor reads the homepage in order
- **THEN** they encounter purpose, demonstration, practical benefits, optional advanced features, common questions, and the final setup action in that order
- **AND** advanced material does not obstruct the beginner path

### Requirement: Describe implemented behavior accurately

The website SHALL answer common questions about supported languages, prerequisites, editor choice, local execution, and sample-test limitations. It SHALL distinguish manual submission from optional cp-submit integration for CSES, AtCoder, and Codeforces, including version compatibility. It SHALL NOT promise built-in automatic submission, enforced memory limits, guaranteed judge acceptance, or browser-hosted execution.

#### Scenario: Assess whether Eitri fits
- **WHEN** a visitor reads the common questions or product claims
- **THEN** they can distinguish local sample testing and manual submission preparation from remote judging
- **AND** language support and prerequisites match the implementation and provided guides


### Requirement: Explain optional automatic submission

The website SHALL link to cp-submit installation instructions as an optional extension for CSES, AtCoder, and Codeforces. It SHALL distinguish problem intake from submission, explain that pressing c queues a solution for the extension, and disclose known compatibility limitations without claiming verified integration.

#### Scenario: Explore automatic submission
- **WHEN** a visitor expands the optional cp-submit guidance
- **THEN** they see its installation link, supported judges, browser/login prerequisites, submission trigger, compatibility note, and manual fallback
- **AND** they understand that pressing c with the extension enabled can submit to the judge while saving only runs local tests

### Requirement: Present a configurable workflow without a fixed language boundary

The homepage SHALL lead with “Your language. Your editor. Your workflow.” and explain the problem → edit → save → test loop. It SHALL describe configured compilation/execution and editor independence, remove the four-language hero row, and label named configurations as examples rather than an exhaustive support list. FAQ and metadata SHALL use the same positioning.

#### Scenario: Visitor uses an unlisted language

- **WHEN** a visitor reads the homepage and language FAQ
- **THEN** they understand that additional languages can be configured through commands and source transformations
- **AND** they can reach illustrative language examples and a custom-language guide

### Requirement: Provide direct beginner and advanced paths

Navigation SHALL provide Get started, Languages & toolchains, How Eitri works, and Configuration reference. The homepage SHALL preview customization without making advanced documentation a prerequisite for the first run.

#### Scenario: Experienced visitor adapts a workflow

- **WHEN** a visitor wants to use an existing language, editor, or library setup
- **THEN** they can directly access concept explanations, the handbook, and the reference
- **AND** beginners can still follow a concise first-run path

### Requirement: Explain integration boundaries separately

The website SHALL distinguish language configuration, editor integration, operating-system/toolchain requirements, problem intake, and remote submission. It SHALL explain that intake depends on Competitive Companion and filename rules, while submission depends on judge support and optional compatible extensions. It SHALL NOT imply universal toolchain portability or universal judge language acceptance.

#### Scenario: Assess a custom environment

- **WHEN** a visitor evaluates an editor, operating system, language, or judge
- **THEN** they can identify the settings and external prerequisites for that choice
- **AND** finite example lists do not imply a fixed limit on Eitri's configuration model

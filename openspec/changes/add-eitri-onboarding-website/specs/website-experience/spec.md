# Spec Delta

## Purpose

Provide a calm, accessible and responsive website experience that remains useful without JavaScript and can be previewed locally as portable static files.

## ADDED Requirements

### Requirement: Default to dark and remember explicit theme choice

The website SHALL use dark mode when no valid saved preference exists, provide a light/dark toggle, and preserve explicit choices across pages and visits when browser storage is available. The saved theme SHALL be applied before first paint to avoid an incorrect-theme flash.

#### Scenario: First visit on a light-system device
- **WHEN** a visitor has no saved website theme and their operating system prefers light mode
- **THEN** the website initially appears in dark mode

#### Scenario: Return after choosing light mode
- **WHEN** a visitor selects light mode and later opens any page with browser storage available
- **THEN** light mode is applied from the first render

#### Scenario: Browser storage is unavailable
- **WHEN** preference storage cannot be read or written
- **THEN** the page remains usable, defaults to dark on a fresh load, and allows theme changes for the current page

### Requirement: Provide accessible and responsive presentation

All pages SHALL use self-hosted Monaspace throughout their interface, restrained warm accents, semantic structure, sufficient contrast in both themes, keyboard-accessible interactions, visible focus indicators, and layouts usable on narrow and wide screens. They SHALL respect reduced-motion preferences.

#### Scenario: Navigate using a keyboard
- **WHEN** a visitor navigates links, theme controls, guide selectors, copy buttons, and expandable content with a keyboard
- **THEN** each control is operable, focus is visible, and meaningful status feedback is available to assistive technology

#### Scenario: Read on a narrow screen
- **WHEN** a visitor opens any page at a 320-pixel viewport width
- **THEN** content and controls remain usable without page-level horizontal scrolling
- **AND** long code blocks can scroll within their own containers

#### Scenario: Reduced motion is requested
- **WHEN** a visitor has enabled reduced motion
- **THEN** decorative animations and animated transitions are suppressed and the demonstration still requires deliberate playback

### Requirement: Preserve essential content without JavaScript

The homepage explanation, navigation, complete walkthroughs, language/toolchain guidance, handbook, and reference SHALL remain readable when JavaScript is disabled. Filtering, theme persistence, and clipboard buttons SHALL enhance rather than gate essential content.

#### Scenario: Browse without JavaScript
- **WHEN** a visitor disables JavaScript
- **THEN** they can understand Eitri and read the complete walkthrough for their language and platform using ordinary links and document navigation

### Requirement: Support independent static pages

The website SHALL expose the homepage at `/`, Get started at `/get-started/`, Languages & toolchains at `/setups/`, How Eitri works at `/handbook/`, and Configuration reference at `/reference/`, with distinct titles/descriptions, favicon, and social preview metadata. All pages SHALL be directly accessible through consistent navigation. Documentation pages SHALL provide tables of contents and stable section anchors.

#### Scenario: Open the guide directly
- **WHEN** a visitor loads `/get-started/` without first visiting the homepage
- **THEN** the guide and its assets load correctly and navigation back to the homepage is available

### Requirement: Keep delivery local and execution separate

The website SHALL be deliverable as portable static files and a local preview without a backend, account, or database. It SHALL NOT execute submissions or contact a visitor's local Eitri server. Following the user's explicit publication request, it SHALL deploy to their Cloudflare account at eitri.veryshyjelly.in.

#### Scenario: Preview the finished website
- **WHEN** the static output is served locally
- **THEN** all pages and their interactions work without application-server services
- **AND** no requests are made to the visitor's local Eitri endpoints

#### Scenario: Complete implementation
- **WHEN** the implementation is handed off
- **THEN** a local preview and portable static output are available and the explicitly authorized Cloudflare domain serves the production site

### Requirement: Support direct links to advanced explanations

Setup, handbook, and reference section URLs SHALL load directly with their assets and shared navigation, without requiring client-side routing or a prior homepage visit.

#### Scenario: Open a shared documentation section

- **WHEN** a visitor opens an advanced section URL directly
- **THEN** the target section loads and navigation offers both beginner and advanced paths
- **AND** its essential content remains accessible without JavaScript


### Requirement: Publish the authorized production domain

The production website SHALL be served over HTTPS at eitri.veryshyjelly.in using the user's Cloudflare account. It SHALL preserve all five routes, use production canonical/social URLs, exclude test fixtures from uploaded assets, and provide a custom 404 page and sitemap. It SHALL NOT alter unrelated hostnames.

#### Scenario: Visit the published website

- **WHEN** a visitor opens the production domain or a direct documentation URL
- **THEN** the intended static page and its assets load over HTTPS
- **AND** unknown pages and internal fixture paths return 404 rather than exposing test data

# Reusable Wizard Module

## Feature Name

Reusable Multi-Step Wizard

## Description

Provide a generic wizard engine for multi-step flows, including opening sessions and registering/programming arrows.

## User Story

> As a developer, I want a reusable wizard framework so that I can easily implement guided multi-step processes without duplicating code.

## Acceptance Criteria

1. WizardHost component renders steps based on a `WizardConfig` array.
2. Each step has a unique `name`, a `component`, and optional guards/validation.
3. Support "Next", "Back", and "Cancel" navigation.
4. Session flow: SessionDetailsStep -> (optional) TargetCalibrationStep.
5. Arrow flow: ArrowDetailsStep -> (optional) ProgramArrowStep.

## UI Requirements

The UI for this feature must follow the global [UI & Look-and-Feel Requirements document](./00_UI_requirements.md), including:
Tailwind CSS with the extended World Archery color palette (wa-reflex, wa-pink, wa-yellow, wa-green, wa-red, wa-sky, wa-black).
Dark mode only theme using the specified palette and typography (Inter font).
Rounded corners for all elements (cards, panels, buttons, inputs).
Layout, table, form, button, and indicator styles exactly as defined in the global document.
Desktop-first responsiveness.
Any deviation from these standards must be documented and approved before implementation.

## Technical Requirements

- Must use: **Vue 3**, **TypeScript**, **Tailwind CSS**
- Integrate with these files:
  - [frontend/index.html](https://github.com/jpmolinamatute/arch-stats/blob/main/frontend/index.html)
  - [frontend/src/App.vue](https://github.com/jpmolinamatute/arch-stats/blob/main/frontend/src/App.vue)
  - [frontend/src/main.ts](https://github.com/jpmolinamatute/arch-stats/blob/main/frontend/src/main.ts)

## API references

- None

## Backend files

- None

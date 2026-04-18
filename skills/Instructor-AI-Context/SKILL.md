# Instructor Suite Architecture Context

This document provides context on the **Instructor Suite** architecture within the UML Comparator application. It is intended for future AI agents to understand the responsibilities and design paradigms of the instructor modules.

## 1. Goal and Purpose
The Instructor Suite consolidates various, previously isolated backend tools (`cmd/cipher`, `cmd/compare`, `cmd/grade_batch`, `cmd/exam_gui`) into a single, cohesive Lorca-based GUI (`instructor_suite.exe`). It's designed to give the lecturer/admin a smooth, user-friendly experience for generating exams, encrypting solutions, batch grading students, and inspecting graphs with admin privileges.

## 2. Core Architecture Loop
The architecture follows a strict decoupled UI logic:
1. **Model/Domain**: Located in `gui/domain/instructor_interfaces.go`. It holds boundary interfaces like `InstructorController` and `InstructorView`.
2. **Service Layer**: Located in `instructor/instructor_service.go`. The `.go` file handles the "Heavy Lifting" operations natively using the core architecture components (e.g. `AppBuilder`, `grader`, `comparator`).
    - `EncryptSolution`
    - `BuildExamTool`
    - `GradeBatch`
    - `CompareUML`
3. **Controller**: Located in `gui/controller/instructor_controller.go`. Handles interactions and asynchronous delegating.
4. **View/GUI**: Located in `gui/view/instructor_view.go`. A Lorca Web UI. It binds JS functions to Go methods securely and injects a dark styled DOM mimicking `visualizer/template.go` (light modern aesthetic).

## 3. Best Practices & Rules
- **No Direct App Actions in JS**: The GUI shouldn't execute direct file manipulations. Always bind functions (e.g., `goSelectFile`, `goExecEncrypt`) that invoke the `InstructorController`.
- **Output Paths**: The Instructor Suite uses **Folder based Selection** for outputs then explicitly adds default filenames to reduce typing effort. Keep this convention.
- **In-App Notifications**: We abandoned the `zenity.Info` pattern for basic alerts. Future edits should try utilizing the `<div id="notification">` pipeline for non-disruptive feedback. Zenity is only for File Selection and Critical Errors.
- **SOLID Dependency**: Features must implement the `InstructorService` and be instantiated from the main entry point `cmd/instructor/main.go`. DO NOT add cross-module dependencies bypassing the Service layer.

## 4. Redundant Modules Tracking
During the Instructor Suite integration (April 2026), the legacy `cmd/grade_batch` and `cmd/cipher` were rendered obsolete for general usage and their execution was stripped from `universal_builder.go`. Whenever editing batch logic or encrypt logic, **DO NOT** edit those legacy CLI commands, instead adjust `instructor/instructor_service.go`.

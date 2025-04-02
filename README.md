# PromptToGo

A terminal-based interactive prompt builder for Large Language Models (LLMs), built with [Huh?](https://github.com/charmbracelet/huh).

## Overview

PromptToGo guides you through creating structured prompts for LLMs with clearly delineated sections:

- **Goal**: What you want the LLM to achieve.
- **Return Format**: How you want the information returned.
- **Warnings**: Things to be cautious about or to avoid.
- **Context Dump**: Additional context or background information (e.g., relevant document text).

The application provides an intuitive interface to fill in each section. Based on your goal, it intelligently offers guidance and suggestions to help you build a more effective prompt.

## Features

- **Interactive Multi-Step Form**: Guides you through filling out Goal, Return Format, Warnings, and Context Dump.
- **Archetype Detection**: Analyzes your stated **Goal** to detect common tasks like "Summarization" or "Question Answering".
- **Contextual Guidance**: Based on the detected task type, the tool displays helpful tips on how to improve your prompt for better results (e.g., suggesting specificity for Q&A, or length considerations for summaries).
- **Intelligent Enrichment**: If a task type is detected, a second optional step allows you to:
    - **Refine Goal**: See and edit a suggested, more structured goal based on the detection (e.g., framing a question appropriately for Q&A).
    - **Select Return Format**: Choose from common return formats relevant to the task (e.g., "Bulleted list", "Direct answer with citations"), while always having the option to keep your original input.
    - **Add Common Warnings**: Easily select pre-defined warnings/constraints relevant to the task (e.g., "Do not infer", "Cite sources", "Avoid jargon") which are appended to any warnings you manually entered.
- **Confirmation Loop**: Before finalizing, you confirm the prompt details. If you select "No", the process restarts, allowing you to easily iterate and refine your input.
- **Clear Output**: Presents the final, structured prompt ready for use.

## Getting Started

Head to [releases](https://github.com/masonbrott/prompttogo/releases) and download the latest version for your platform.

Run the application:

```bash
./prompttogo
```

```powershell
.\prompttogo.exe
```

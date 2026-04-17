# AI-Context CLI

A simple and lightning-fast CLI tool written in Go. Designed for developers who are tired of hitting token limits in IDE-based AI extensions and want a quick way to generate project context for Web-based AI chats (like Gemini, ChatGPT, or Claude).

> [!Note]
> This tool was born out of pure laziness. I was tired of manual copy-pasting and wasting my precious time (and tokens!) just to give AI some context. So, I built this to do the heavy lifting for me.

## Why use this? (The Pain Point)
Pasting entire codebases into an AI often leads to "hallucinations" or unwanted side effects because the AI loses focus. On the other hand, IDE-based AIs often run out of tokens quickly.

This tool extracts your current **Project Structure** and combines it with a pre-configured **English Prompt Template** (to save tokens!). You can instantly pipe the output to your clipboard and paste it into any web-based AI. The AI will then analyze your structure and ask *only* for the specific files it needs to complete your task.

**Key Features:**
- Automatically ignores unnecessary directories (e.g., `.git`, `node_modules`, `.next`, `dist`).
- Provide tasks via a command-line flag or an interactive prompt.
- Pipe directly to your OS clipboard—no manual highlighting or copying required.

---

## Installation

Since it's written in Go, you can build it into a single executable binary and run it anywhere!

1. Clone this repository or save the `main.go` file to your local machine.
2. Build the binary:
   ```bash
   go build -o ai-context main.go
   ```
3. Move the binary to your System Path so you can run it from any directory:
- Mac/Linux:
  ```bash
  sudo mv ai-context /usr/local/bin/
  ```
- Windows: Move the ai-context.exe file to a folder that is included in your Environment Variables (e.g., C:\Windows or a custom C:\Tools added to PATH).

---

## Usage

Open your terminal in the root folder of the project you are currently working on.

### Method 1: Using the -task Flag (One-liner)
Specify what you want the AI to do right in the command.

- Mac/Linux:
    ```bash
    ai-context -task "add login api" | pbcopy
    ```

- Window:
    ```bash
    ai-context -task "add login api" | clip
    ```

### Method 2: Interactive Mode
If you have a longer task or just want to run the command quickly, type the command without flags. The program will prompt you for your task.

- Mac/Linux:
    ```bash
    ai-context | pbcopy
    ```

- Windows:
    ```bash
    ai-context | clip
    ```

(When you hit Enter, the terminal will wait and ask: What is your task?:. Type your task, press Enter again, and the generated context will be instantly sent to your clipboard.)

---

## Output Example
After running the command, your clipboard will contain text similar to this (ready to `Ctrl+V` or `Cmd+V` into your AI chat):

```text
Here is my project structure:

- .github
  - workflows
    - lint_test.yml
    - unit_test.yml
- .gitignore
- README.md
- cmd
  - logspy
    - main.go
- go.mod
- internal
  - handler
    - handler.go

Currently, I am working on this project. I want to: add login api.
To avoid any side effects, please analyze this structure and tell me exactly which files you need to see the code from?
```
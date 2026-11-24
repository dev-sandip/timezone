🌍 Timezone

A lightweight, interactive terminal-based timezone converter built with Go and Bubble Tea.

Convert time from different timezones to Asia/Kathmandu directly from your terminal.

✨ Features
Interactive terminal UI (TUI).
Live timezone search and filtering.
Supports AM/PM and 24-hour time formats.
Search and select IANA timezones (CST, EST, PST, UTC, IST, etc.).
Automatic daylight saving time handling.
Displays source and converted UTC offsets.
Uses Go's embedded timezone database.
🛠️ Tech Stack
Go
Bubble Tea
Lipgloss
📦 Installation

Clone the repository:

bash
git clone https://github.com/dev-sandip/timezone.git
cd timezone

Install dependencies:

bash
go mod tidy

🚀 Run from Source
bash
go run .

🔨 Build
Linux
bash
go build -o timezone .
./timezone

Windows (PowerShell / CMD)
powershell
go build -o timezone.exe .
.\timezone.exe

🌐 Install as a System Command
Linux (Fedora / Ubuntu)

Build and install:

bash
go build -o timezone .
sudo install -m 755 timezone /usr/local/bin/timezone

Run from anywhere:

bash
timezone

Windows

Build:

powershell
go build -o timezone.exe .

Create a directory and copy the executable:

powershell
New-Item -ItemType Directory -Force C:\Tools
Copy-Item .\timezone.exe C:\Tools\timezone.exe

Add C:\Tools to your Windows PATH environment variable.

Open a new terminal and run:

powershell
timezone

⌨️ Keyboard Shortcuts
Key Action
TAB Next input field
SHIFT + TAB Previous input field
CTRL + A Toggle AM/PM
ENTER Continue / Select
UP / DOWN Navigate timezones
CTRL + U Clear timezone search
ESC Go back
CTRL + C Quit
🎯 Default Target Timezone
text
Asia/Kathmandu (NPT / UTC+05:45)

📁 Project Structure
text
timezone/
├── go.mod
├── go.sum
└── main.go

📜 License

MIT License. See LICENSE for details.

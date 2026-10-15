# Skills and sandbox

Agent skills run in a sandbox configured for the workspace. Supported backends include local Docker and E2B. Select a trusted image, configure network policy, and install the skill into that backend before enabling it for an agent.

The chat side panel provides an interactive terminal for the session sandbox on Docker and E2B. An E2B terminal can reattach after a page reload. A Docker terminal ends its shell when it disconnects (background jobs started with `nohup` keep running), and reconnecting opens a new shell.

The skill list supports case-insensitive search by name and description; all space-separated keywords must match.

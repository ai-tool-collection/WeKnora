# Sandbox deployment

Configure Docker or E2B as a workspace sandbox. Docker requires access to a compatible daemon and a trusted runtime image; E2B requires its API credentials. Limit host filesystem mounts, network access, and resource use according to your deployment policy. Test the sandbox connection before enabling agent skills.

## Interactive terminal

The terminal connects to the session sandbox over WebSocket. The browser first exchanges an authenticated POST for a short-lived ticket and passes it in the handshake query.

The Docker terminal is a TTY exec, which the Engine API cannot reattach:

- While the terminal is open, the sandbox stays active and is not reclaimed by `DockerIdleTTL`. After it closes, normal idle reclamation applies.
- On disconnect, the shell and its child processes receive SIGHUP. Foreground jobs end; jobs started with `nohup` keep running.
- The frontend does not reconnect automatically. **Reconnect** opens a new shell.
- If the container was reclaimed and removed, a new sandbox starts and earlier files are not kept.

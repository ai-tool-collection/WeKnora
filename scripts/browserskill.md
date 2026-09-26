# Optional local browser integration

Knowledge Hub can connect to a locally hosted BrowserSkill daemon and browser
extension. This component is optional and disabled unless configured.

The repository does not download the BrowserSkill source or binaries. Review and
provide a local source checkout at the pinned commit in
`scripts/browserskill-release.json`, then run:

```sh
BROWSERSKILL_SOURCE_DIR=/path/to/reviewed/source \
  scripts/build_browserskill.sh artifacts/browserskill
```

The build requires Rust/Cargo, a C compiler, and Node.js. The extension and
daemon must come from the same source commit. Keep the included upstream license
with any redistributed artifacts.

For a prebuilt local installation, set `BROWSERSKILL_BINARY` to the daemon path
and `BROWSERSKILL_EXTENSION_PATH` to the extension archive. Configure
`BROWSERSKILL_PUBLIC_URL` and `BROWSERSKILL_INTERNAL_URL` for remote deployment.

The optional UI channel uses `ui.task_preview` and `ui.task_focus` on an
authenticated socket. The server limits those methods to the current user's
owned browser task.

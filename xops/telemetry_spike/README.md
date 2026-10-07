# Isolated Flutter telemetry candidate spike

Use the workspace-local Flutter 3.47.5 toolchain described in
[deployment operations](../../deploy/README.md). This is not the game app and
adds no production dependency or network exporter.

Run `flutter test`, then `flutter build web --no-web-resources-cdn` in this
directory. Serve `build/web` on a loopback HTTP server and press **Run bounded
measurement**. Each mode warms 1,000 spans and measures 30 blocks of 1,000 span
start/end pairs; blocks yield to the event loop. Read the report as per-block
mean microseconds per span, not individual-span tail latency or frame overhead.

The no-op provider is the unregistered global API default. The recording provider
has no exporter/processors. Every span has the constant name `synthetic`, without
attributes, gameplay state, tokens, URLs or exception content. Results are local
screen text only. No telemetry is sent anywhere.

The pinned package's publisher documents traces beta, metrics alpha and logs
unimplemented, with web platform support. The measured production-selection
limits and results are recorded in [validation](../../deploy/VALIDATION.md).
Native-device/export/lifecycle tests remain required before broader adoption.

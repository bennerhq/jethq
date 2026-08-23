# jethq

A fork of [lkarlslund/jetkvm-desktop](https://github.com/lkarlslund/jetkvm-desktop). Thanks Lars!

A native desktop client for JetKVM with local discovery, direct connect, remote control, and core settings in one window.

## What It Does

- Finds JetKVM devices on your local network
- Connects directly by hostname, mDNS name, or IP
- Shows the remote video feed in a native desktop window
- Sends keyboard and mouse input to the target machine
- Supports mouse back/forward side buttons in the native client, unlike the browser UI
- Prompts for a password when the device requires it
- Exposes the main settings and connection stats without opening the browser UI

## Getting Started

Open the launcher:

```bash
jethq
```

Connect straight to a known device:

```bash
jethq jetkvm.local
jethq 192.168.1.50
jethq http://192.168.1.50
```

If the device requires a password, the app will ask for it.

## JetKVM

This is a separate desktop client for the JetKVM ecosystem. For the upstream JetKVM repositories, see `github.com/jetkvm`.
For the upstream desktop client this is forked from, see `github.com/lkarlslund/jetkvm-desktop`.
# no-root Notes
Testing out logic to identify if a go program is actually be run by root via nefarious means.

## Edge Cases

- Shell scripts: Someone could wrap your program in a script
- Systemd services: Programs run via systemd might not show obvious privilege escalation
- Container escapes: Docker/podman might obscure the real process tree
- SSH with forced commands: ssh user@host 'your-program'
- Cron jobs: Running as different users via cron

## Commands that are banned

- `su`
- `sudo`
- `doas`
- `super`
- `calife`
- `pkexec`
- `systemd-run`
- `machinectl`
- `runuser`

## Order of Operations

1. Get Go process info
2. Get current user
3. Get list of processes
4. Attempt to find the command that started calling this process by traversing up the PPID tree.

## Traversal Logic

This logic needs more fleshing out and edge cases, but the general gist is below.

1. Check owner of PPID of Go process.
2. If same as current user, continue traversing up.
3. If root, check if command is on the banned list. Then check if command has setuid set. Final fallback is to check if command line portion of the process contains the current user. If any of this criteria matches, throw an error because we've figured out that someone is using the run user to try and run this program as the current user.

- Add logic to search for all files based on command name and then check those. If any meet the criteria above along with if it is even executable. For instance, if we have a sudo in some non-standard folder, but it is not executable, ignore it.

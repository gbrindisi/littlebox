## 1. Dockerfile Changes

- [x] 1.1 Remove sudo package from apt-get install list
- [x] 1.2 Remove sudoers-firewall COPY and chmod commands
- [x] 1.3 Delete the sudoers-firewall file
- [x] 1.4 Change USER directive from `agent` to `root`

## 2. Entrypoint Changes

- [x] 2.1 Remove sudo call from firewall initialization (run init-firewall.sh directly)
- [x] 2.2 Add UID/GID resolution for agent user (using `id -u agent` and `id -g agent`)
- [x] 2.3 Update setpriv exec to include --reuid, --regid, and --init-groups flags
- [x] 2.4 Update comments to reflect new privilege flow (root -> agent)

## 3. Firewall Script Cleanup

- [x] 3.1 Remove sudoers cleanup code from init-firewall.sh (lines that rm /etc/sudoers.d/*)
- [x] 3.2 Update script comments to reflect it runs as root directly

## 4. Test Updates

- [x] 4.1 Remove `no_new_privileges: false` workarounds from test configs
- [x] 4.2 Verify existing tests pass with new privilege model
- [x] 4.3 Add test verifying container works with no_new_privileges: true (default)

## 5. Documentation and Cleanup

- [x] 5.1 Update README security section to reflect new privilege model
- [x] 5.2 Remove no_new_privileges: false from example Agentfiles (root Agentfile, test-environment/Agentfile)
- [x] 5.3 Update init.go template comments (remove no_new_privileges workaround mention)

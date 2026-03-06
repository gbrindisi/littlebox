## ADDED Requirements

### Requirement: Allowed IP export
The security module SHALL export the allowed IP list to a file for LD_PRELOAD library access.

#### Scenario: Export after ipset population
- **WHEN** init-firewall.sh finishes populating the allowed_ips ipset
- **THEN** it exports the ipset contents to /run/sandbox/allowed_ips

#### Scenario: File permissions
- **WHEN** /run/sandbox/allowed_ips is created
- **THEN** it is readable by all users (mode 444)

#### Scenario: Directory creation
- **WHEN** init-firewall.sh exports the allowed IPs
- **THEN** it creates /run/sandbox directory if it does not exist

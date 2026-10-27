---
organization: Turbot
category: ["software development"]
icon_url: "/images/plugins/turbot/nomad.svg"
brand_color: "#00CA8E"
display_name: "Nomad"
short_name: "nomad"
description: "Steampipe plugin to query nodes, jobs, deployments and more from Nomad."
og_description: "Query Nomad with SQL! Open source CLI. No DB required."
og_image: "/images/plugins/turbot/nomad-social-graphic.png"
engines: ["steampipe", "sqlite", "postgres", "export"]
---

# Nomad + Steampipe

[Nomad](https://www.nomadproject.io/) is a simple and flexible scheduler and orchestrator for managing containers and non-containerized applications across on-prem and clouds at scale.

[Steampipe](https://steampipe.io) is an open-source zero-ETL engine to instantly query cloud APIs using SQL.

List your Nomad jobs:

```sql
select
  id,
  name,
  status,
  dispatched,
  namespace,
  priority,
  region
from
  nomad_job;
```

```
+------+------+---------+------------+-----------+----------+--------+
| id   | name | status  | dispatched | namespace | priority | region |
+------+------+---------+------------+-----------+----------+--------+
| docs | docs | pending | false      | default   | 50       | global |
+------+------+---------+------------+-----------+----------+--------+
```

## Documentation

- **[Table definitions & examples →](/plugins/turbot/nomad/tables)**

## Quick start

### Install

Download and install the latest Nomad plugin:

```sh
steampipe plugin install nomad
```

### Credentials

| Item        | Description                                                                                                                                                                              |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Credentials | Nomad requires an `address` and `namespace` or `address`, `namespace` and [Secret ID](https://developer.hashicorp.com/nomad/tutorials/access-control/access-control-tokens) for all requests.    |
| Permissions | The permission scope of Secret IDs is set by the Admin at the creation time of the [ACL tokens](https://developer.hashicorp.com/nomad/tutorials/web-ui/web-ui-access).                                                                                           |
| Radius      | Each connection represents a single Nomad Installation.                                                                                                                                  |
| Resolution  | 1. Credentials explicitly set in a steampipe config file (`~/.steampipe/config/nomad.spc`)<br />2. Credentials specified in environment variables, e.g., `NOMAD_ADDR` and `NOMAD_TOKEN`. |

### Configuration

Installing the latest nomad plugin will create a config file (`~/.steampipe/config/nomad.spc`) with a single connection named `nomad`:

Configure your account details in `~/.steampipe/config/nomad.spc`:

```hcl
connection "nomad" {
  plugin = "nomad"

  # Address is required for requests. Required.
  # This can also be set via the NOMAD_ADDR environment variable.
  # address = "http://18.118.164.168:4646"

  # The secret ID of ACL token is required for ACL-enabled Nomad servers. Optional.
  # For more information about the ACL Token, please see https://developer.hashicorp.com/nomad/tutorials/access-control/access-control-tokens.
  # This can also be set via the NOMAD_TOKEN environment variable.
  # secret_id = "c178b810-8b18-6f38-016f-725ddec5d58"

  # Namespace is required for Nomad Enterprise access. Optional.
  # API will execute with default namespace if this parameter is not set.
  # This can also be set via the NOMAD_NAMESPACE environment variable.
  # "*" indicates all the namespaces available.
  # namespace = "*"

  # TLS handshake timeout in seconds. Optional.
  # Increase this if you see "net/http: TLS handshake timeout" errors on slow
  # or remote links. Default: 30.
  # tls_timeout_seconds = 60

  # Maximum concurrent connections to the Nomad HTTP API. Optional.
  # Keep this below the agent's limits.http_max_conns_per_client (default 100),
  # otherwise Nomad resets the excess connections ("connection reset by peer").
  # Default: 6.
  # max_connections = 6
}

```

- `tls_timeout_seconds`: TLS handshake timeout for requests to the Nomad API. Defaults to 30 seconds; increase it when querying a remote cluster over a slow link.
- `max_connections`: Maximum concurrent TCP connections to the Nomad API. Defaults to 6. Nomad agents close connections from one IP beyond [`limits.http_max_conns_per_client`](https://developer.hashicorp.com/nomad/docs/configuration#http_max_conns_per_client) (default 100), which shows up as `connection reset by peer`; keep this value below that limit.
- `secret_id` parameter is only required to query the ACL tables like `nomad_acl_auth_method`, `nomad_acl_binding_rule`, `nomad_acl_policy`, `nomad_acl_role` and `nomad_acl_token` tables.
- `namespace` parameter is only required to query the `nomad_namespace` table.

Alternatively, you can also use the standard Nomad environment variable to obtain credentials **only if other arguments (`address`, `token`, and `namespace`) are not specified** in the connection:

```sh
export NOMAD_ADDR=http://18.118.144.168:4646
export NOMAD_TOKEN=c178b810-8b18-6f38-016f-725ddec5d58
export NOMAD_NAMESPACE=*
```



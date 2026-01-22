# ali-nuke

A command-line tool to delete all resources from an Alibaba Cloud account.

> **WARNING:** This tool is highly destructive and irreversible. It cannot distinguish between production and non-production resources. Do not run this on any account where you cannot afford to lose all resources. Always back up critical data and configurations before execution.

## Table of Contents

- [Supported Resources](#supported-resources)
- [Usage](#usage)
  - [Basic Command](#basic-command)
  - [Command Line Options](#command-line-options)
  - [Dry Run Mode (Default)](#dry-run-mode-default)
  - [Actual Deletion](#actual-deletion)
- [Configuration File](#configuration-file)
  - [Example Configuration](#example-configuration)
  - [Configuration Sections](#configuration-sections)
- [Alibaba Cloud Regions](#alibaba-cloud-regions)
- [Authentication](#authentication)
  - [Creating an Access Key](#creating-an-access-key)
- [Resource Deletion Order](#resource-deletion-order)
- [Exit Codes](#exit-codes)
- [License](#license)

## Supported Resources

| Resource Type | Product | Description |
|---------------|---------|-------------|
| `ECSInstance` | Elastic Compute Service | Virtual machine instances |
| `VPC` | Virtual Private Cloud | Virtual private networks |
| `VSwitch` | Virtual Private Cloud | Subnets within VPCs |

## Usage

### Basic Command

```bash
ali-nuke nuke \
  --config config.yaml \
  --access-key-id <YOUR_ACCESS_KEY_ID> \
  --access-key-secret <YOUR_ACCESS_KEY_SECRET>
```

### Command Line Options

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--config` | `-c` | Yes | Path to the configuration file |
| `--access-key-id` | | Yes | Alibaba Cloud Access Key ID |
| `--access-key-secret` | | Yes | Alibaba Cloud Access Key Secret |
| `--no-dry-run` | | No | Actually delete resources (default is dry-run mode) |

### Dry Run Mode (Default)

By default, `ali-nuke` runs in **dry-run mode**, which scans and lists all resources without deleting them:

```bash
ali-nuke nuke \
  --config config.yaml \
  --access-key-id <YOUR_ACCESS_KEY_ID> \
  --access-key-secret <YOUR_ACCESS_KEY_SECRET>
```

Example output:
```
Fetching available regions...
Scanning 28 regions (excluded 1)...
Scan complete: Found 70 resources in total. To be removed 70, Filtered 0
┌────────────────┬─────────┬─────────────────────┬────────┐
│     REGION     │ PRODUCT │       ID / NAME     │ STATUS │
├────────────────┼─────────┼─────────────────────┼────────┤
│ eu-central-1   │ VSwitch │ sub1                │ Ready  │
│ eu-central-1   │ VPC     │ MyVPC               │ Ready  │
│ cn-hangzhou    │ VPC     │ Production-VPC      │ Ready  │
└────────────────┴─────────┴─────────────────────┴────────┘
Dry run complete.
```

### Actual Deletion

To actually delete resources, add the `--no-dry-run` flag:

```bash
ali-nuke nuke \
  --config config.yaml \
  --access-key-id <YOUR_ACCESS_KEY_ID> \
  --access-key-secret <YOUR_ACCESS_KEY_SECRET> \
  --no-dry-run
```

You will be prompted to type `yes` to confirm the deletion.

## Configuration File

The configuration file (YAML format) allows you to exclude specific regions, resource types, or individual resources from deletion.

### Example Configuration

```yaml
# Regions to exclude from scanning
# All other available Alibaba Cloud regions will be scanned
regions:
  excludes:
    - cn-hongkong      # Skip Hong Kong region
    - ap-southeast-1   # Skip Singapore region

# Resource types to exclude from deletion
resource-types:
  excludes:
    - VPC              # Don't delete any VPCs
    - ECSInstance      # Don't delete any ECS instances

# Specific resource IDs or names to exclude from deletion
resource-ids:
  excludes:
    - resourceType: VPC
      id: vpc-abc123456789      # Exclude by resource ID
    - resourceType: ECSInstance
      id: i-bp1234567890abcdef  # Exclude by instance ID
    - resourceType: VSwitch
      id: production-subnet     # Exclude by resource name
```

### Configuration Sections

#### `regions`

Exclude specific Alibaba Cloud regions from scanning. All other regions will be scanned.

```yaml
regions:
  excludes:
    - cn-hongkong
    - us-west-1
```

Available regions are fetched dynamically from the Alibaba Cloud API, ensuring compatibility with newly added regions.

#### `resource-types`

Exclude entire resource types from deletion.

```yaml
resource-types:
  excludes:
    - ECSInstance
    - VPC
    - VSwitch
```

#### `resource-ids`

Exclude specific resources by their ID or name.

```yaml
resource-ids:
  excludes:
    - resourceType: VPC
      id: vpc-production-001
    - resourceType: ECSInstance
      id: i-critical-server
```

## Alibaba Cloud Regions

The tool automatically discovers all available Alibaba Cloud regions using the `DescribeRegions` API. Common regions include:

| Region ID | Location |
|-----------|----------|
| `cn-hangzhou` | China (Hangzhou) |
| `cn-shanghai` | China (Shanghai) |
| `cn-beijing` | China (Beijing) |
| `cn-shenzhen` | China (Shenzhen) |
| `cn-hongkong` | China (Hong Kong) |
| `ap-southeast-1` | Singapore |
| `ap-northeast-1` | Japan (Tokyo) |
| `eu-central-1` | Germany (Frankfurt) |
| `eu-west-1` | UK (London) |
| `us-east-1` | US (Virginia) |
| `us-west-1` | US (Silicon Valley) |

## Authentication

`ali-nuke` requires an Alibaba Cloud Access Key with sufficient permissions to list and delete resources. 

### Creating an Access Key

1. Log in to the [Alibaba Cloud Console](https://www.alibabacloud.com/)
2. Navigate to **AccessKey Management**
3. Create a new AccessKey pair
4. Store the Access Key ID and Secret securely

> **Security Tip:** Use a dedicated RAM user with only the necessary permissions instead of your root account credentials.

## Resource Deletion Order

When deleting resources, dependencies matter. For example:
- VSwitches must be deleted before their parent VPCs
- ECS instances must be deleted before their associated VSwitches

The tool handles these dependencies by attempting deletion in parallel and retrying failed resources.

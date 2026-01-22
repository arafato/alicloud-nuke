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

### Elastic Compute Service (ECS)

| Resource Type | Description |
|---------------|-------------|
| `ECSInstance` | Virtual machine instances |
| `Disk` | Cloud disks (excludes system disks and attached disks) |
| `Snapshot` | Disk snapshots |
| `Image` | Custom images (excludes public/marketplace images) |
| `SecurityGroup` | Security groups for network access control |
| `NetworkInterface` | Elastic Network Interfaces (ENI), excludes primary ENIs |
| `KeyPair` | SSH key pairs |
| `LaunchTemplate` | Launch templates for instance creation |
| `AutoSnapshotPolicy` | Automatic snapshot policies |
| `Command` | Cloud Assistant commands (user-created only) |
| `DeploymentSet` | Deployment sets for distributed instances |

### Virtual Private Cloud (VPC)

| Resource Type | Description |
|---------------|-------------|
| `VPC` | Virtual private networks |
| `VSwitch` | Subnets within VPCs |
| `RouteTable` | Custom route tables (system route tables are automatically excluded) |
| `RouterInterface` | Router interfaces for VPC peering connections |
| `NatGateway` | NAT Gateways for internet access |
| `EIP` | Elastic IP addresses |
| `CommonBandwidthPackage` | Shared bandwidth packages |
| `ForwardEntry` | DNAT entries for NAT Gateways |
| `SnatEntry` | SNAT entries for NAT Gateways |
| `HaVip` | High Availability Virtual IPs |
| `FlowLog` | VPC flow logs |
| `VpnGateway` | VPN Gateways |
| `VpnConnection` | IPsec VPN connections |
| `CustomerGateway` | Customer gateways for VPN |
| `SslVpnServer` | SSL VPN servers |
| `SslVpnClientCert` | SSL VPN client certificates |

### Network Attached Storage (NAS)

| Resource Type | Description |
|---------------|-------------|
| `NASFileSystem` | NAS file systems |
| `NASMountTarget` | NAS mount targets |

### Auto Scaling (ESS)

| Resource Type | Description |
|---------------|-------------|
| `ScalingGroup` | Auto Scaling groups |
| `ScalingConfiguration` | Auto Scaling configurations |

### Container Services

| Resource Type | Description |
|---------------|-------------|
| `ContainerRegistryRepo` | Container Registry repositories |

### Load Balancing

| Resource Type | Description |
|---------------|-------------|
| `SLB` | Classic Load Balancer instances |
| `ALB` | Application Load Balancer instances |
| `NLB` | Network Load Balancer instances |

### Databases

| Resource Type | Description |
|---------------|-------------|
| `RDSInstance` | RDS database instances (MySQL, PostgreSQL, SQL Server, MariaDB) |
| `RedisInstance` | Redis (KVStore) instances |
| `MongoDBInstance` | MongoDB (ApsaraDB for MongoDB) instances |
| `PolarDBCluster` | PolarDB clusters (MySQL, PostgreSQL, Oracle compatible) |

### Object Storage

| Resource Type | Description |
|---------------|-------------|
| `OSSBucket` | OSS buckets (including all objects and versions) |

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
    - NASFileSystem    # Don't delete any NAS file systems

# Specific resource IDs or names to exclude from deletion
resource-ids:
  excludes:
    - resourceType: VPC
      id: vpc-abc123456789      # Exclude by resource ID
    - resourceType: ECSInstance
      id: i-bp1234567890abcdef  # Exclude by instance ID
    - resourceType: VSwitch
      id: production-subnet     # Exclude by resource name
    - resourceType: SecurityGroup
      id: sg-production         # Exclude by security group name
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
    # ECS Resources
    - ECSInstance
    - Disk
    - Snapshot
    - Image
    - SecurityGroup
    - NetworkInterface
    - KeyPair
    - LaunchTemplate
    - AutoSnapshotPolicy
    - Command
    - DeploymentSet
    # VPC Resources
    - VPC
    - VSwitch
    - RouteTable
    - RouterInterface
    - NatGateway
    - EIP
    - CommonBandwidthPackage
    - ForwardEntry
    - SnatEntry
    - HaVip
    - FlowLog
    - VpnGateway
    - VpnConnection
    - CustomerGateway
    - SslVpnServer
    - SslVpnClientCert
    # NAS Resources
    - NASFileSystem
    - NASMountTarget
    # Auto Scaling Resources
    - ScalingGroup
    - ScalingConfiguration
    # Container Resources
    - ContainerRegistryRepo
    # Load Balancer Resources
    - SLB
    - ALB
    - NLB
    # Database Resources
    - RDSInstance
    - RedisInstance
    - MongoDBInstance
    - PolarDBCluster
    # Object Storage Resources
    - OSSBucket
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

When deleting resources, dependencies matter. The tool uses a **wave-based retry system** to handle resource dependencies automatically:

1. **Wave 1**: All resources are deleted in parallel
2. Resources that fail with dependency errors (e.g., `DependencyViolation`) are marked for retry
3. **Wave 2+**: After a 10-second interval, failed resources are retried
4. This continues until all resources are deleted or a 10-minute timeout is reached

### Common Dependency Chains

| Delete First | Then Delete |
|--------------|-------------|
| ECS Instances | Disks, Security Groups, VSwitches, Network Interfaces, Key Pairs |
| Snapshots | (should be deleted before Images that depend on them) |
| Images | (can be deleted independently) |
| Disks | (unattached disks can be deleted independently) |
| SSL VPN Client Certs | SSL VPN Servers |
| SSL VPN Servers | VPN Gateways |
| VPN Connections | VPN Gateways, Customer Gateways |
| VPN Gateways | VPCs |
| Forward Entries (DNAT) | NAT Gateways |
| SNAT Entries | NAT Gateways |
| NAT Gateways | VSwitches, EIPs |
| EIPs | (can be deleted independently after unassociation) |
| NAS Mount Targets | NAS File Systems |
| NAS File Systems | VSwitches |
| Network Interfaces (ENI) | VSwitches |
| Security Groups | VPCs |
| VSwitches | VPCs |
| Custom Route Tables | VPCs |
| Router Interfaces | VPCs |
| HA VIPs | VSwitches |
| Launch Templates | (can be deleted independently) |
| Auto Snapshot Policies | (can be deleted independently) |
| Commands | (can be deleted independently) |
| Deployment Sets | (must delete instances first) |
| Flow Logs | (can be deleted independently) |
| Scaling Configurations | Scaling Groups |
| Scaling Groups | VPCs, VSwitches |
| Container Registry Repos | (can be deleted independently) |
| SLB (Classic Load Balancer) | VPCs, VSwitches |
| ALB (Application Load Balancer) | VPCs, VSwitches |
| NLB (Network Load Balancer) | VPCs, VSwitches |
| RDS Instances | VPCs, VSwitches, Security Groups |
| Redis Instances | VPCs, VSwitches |
| MongoDB Instances | VPCs, VSwitches |
| PolarDB Clusters | VPCs, VSwitches |
| OSS Buckets | (can be deleted independently, objects deleted first) |

> **Note:** System route tables (created automatically with VPCs) are excluded from deletion as they are managed by Alibaba Cloud and deleted when the parent VPC is removed.

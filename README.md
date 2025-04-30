# Cloud Resource Suggestion API

## Overview
This project provides actionable cloud resource suggestions and resource listings to help businesses eliminate waste, enhance efficiency, and maximize business value.

## API Endpoints

### `/suggestions`
Returns a list of actionable suggestions for cloud resources, including business context and potential savings.

**Example Response:**

```json
[
    {
        "resource_id": "vm-2",
        "resource_type": "VM",
        "message": "VM 'vm-2' is over-provisioned. Consider rightsizing to reduce spend.",
        "estimated_savings_usd": 60,
        "severity": "Info",
        "priority": 3,
        "timestamp": "2025-04-30T23:43:34.059806655+05:30",
        "action": "Resize down",
        "details": {
            "current_type": "t3.xlarge",
            "recommended_type": "t3.large"
        },
        "docs_link": "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html"
    },
    {
        "resource_id": "s-1",
        "resource_type": "Storage",
        "message": "Storage 's-1' is idle and can be moved to a lower-cost storage class to eliminate waste.",
        "estimated_savings_usd": 10,
        "severity": "Warning",
        "priority": 2,
        "timestamp": "2025-04-30T23:43:34.059892099+05:30",
        "action": "Move to infrequent access tier",
        "details": {
            "business_impact": "Idle storage can be archived or deleted to save costs.",
            "owner": "Data Science",
            "region": "us-east-1",
            "storage_class": "standard"
        },
        "docs_link": "https://docs.aws.amazon.com/AmazonS3/latest/userguide/storage-class-intro.html"
    },
    {
        "resource_id": "db-1",
        "resource_type": "Database",
        "message": "Database 'db-1' has a high number of connections. Consider scaling up or load balancing.",
        "severity": "Critical",
        "priority": 0,
        "timestamp": "2025-04-30T23:43:37.062300877+05:30",
        "action": "Scale up or load balance",
        "details": {
            "business_impact": "High connection count; scaling or balancing can prevent outages and improve business continuity.",
            "connections": 180,
            "owner": "Analytics"
        },
        "docs_link": "https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_WorkingWithConnections.html"
    },
    {
        "resource_id": "db-1",
        "resource_type": "Database",
        "message": "Database 'db-1' has high CPU usage. Consider query optimization or upgrading instance.",
        "severity": "Warning",
        "priority": 2,
        "timestamp": "2025-04-30T23:43:38.062583743+05:30",
        "action": "Optimize or upgrade",
        "details": {
            "business_impact": "High DB CPU usage; optimizing or upgrading can improve performance and user experience.",
            "cpu_usage": 73.78918739496602,
            "owner": "Analytics"
        },
        "docs_link": "https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/MonitoringOverview.html"
    },
    {
        "resource_id": "vm-2",
        "resource_type": "VM",
        "message": "VM 'vm-2' is underutilized (CPU < 10%) for 14 days. Consider resizing or terminating to eliminate waste.",
        "estimated_savings_usd": 45,
        "severity": "Critical",
        "priority": 0,
        "timestamp": "2025-04-30T23:43:40.064291152+05:30",
        "action": "Resize or terminate",
        "details": {
            "business_impact": "No recent activity; freeing this VM will save significant costs.",
            "current_type": "t3.large",
            "owner": "Engineering",
            "recommended_type": "t3.small",
            "region": "us-east-1"
        },
        "docs_link": "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html"
    },
    {
        "resource_id": "vm-1",
        "resource_type": "VM",
        "message": "VM 'vm-1' is over-provisioned. Consider rightsizing to reduce spend.",
        "estimated_savings_usd": 60,
        "severity": "Info",
        "priority": 3,
        "timestamp": "2025-04-30T23:43:40.063893232+05:30",
        "action": "Resize down",
        "details": {
            "current_type": "t3.xlarge",
            "recommended_type": "t3.large"
        },
        "docs_link": "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html"
    },
    {
        "resource_id": "vm-1",
        "resource_type": "VM",
        "message": "VM 'vm-1' is underutilized (CPU < 10%) for 14 days. Consider resizing or terminating to eliminate waste.",
        "estimated_savings_usd": 45,
        "severity": "Critical",
        "priority": 0,
        "timestamp": "2025-04-30T23:43:41.065104332+05:30",
        "action": "Resize or terminate",
        "details": {
            "business_impact": "No recent activity; freeing this VM will save significant costs.",
            "current_type": "t3.large",
            "owner": "Finance Team",
            "recommended_type": "t3.small",
            "region": "us-east-1"
        },
        "docs_link": "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html"
    },
    {
        "resource_id": "db-1",
        "resource_type": "Database",
        "message": "Database 'db-1' is over-provisioned. Consider downsizing to reduce waste.",
        "estimated_savings_usd": 25,
        "severity": "Info",
        "priority": 3,
        "timestamp": "2025-04-30T23:43:42.066902928+05:30",
        "action": "Downsize instance",
        "details": {
            "business_impact": "Over-provisioned DB; downsizing will reduce waste and save costs.",
            "current_size": "db.m5.large",
            "engine": "Postgres",
            "owner": "Analytics",
            "recommended_size": "db.t3.medium"
        },
        "docs_link": "https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstanceClass.html"
    }
]
```

### `/resources`
Returns a list of cloud resources and their key properties.

**Example Response:**

```json
[
    {
        "ID": "vm-1",
        "CPUUsage": 56.44094154514207,
        "CostPerHour": 0.05,
        "PreviousCostPerHour": 0,
        "Owner": "Finance Team",
        "LastActive": 1746034638
    },
    {
        "ID": "vm-2",
        "CPUUsage": 59.1998047137639,
        "CostPerHour": 0.1,
        "PreviousCostPerHour": 0,
        "Owner": "Engineering",
        "LastActive": 1746034643
    },
    {
        "ID": "s-1",
        "UsedGB": 13.402588400774338,
        "CostPerGB": 0.02,
        "PreviousCostPerGB": 0,
        "LastAccessed": 1746034643,
        "Owner": "Data Science"
    },
    {
        "ID": "db-1",
        "Connections": 182,
        "CPUUsage": 75.69160268055239,
        "CostPerHr": 0.2,
        "PreviousCostPerHr": 0,
        "Owner": "Analytics"
    }
]
```

## Mission Alignment
All suggestions and features are designed to help businesses gain control over cloud spending, eliminate waste, enhance efficiency, and maximize business value.

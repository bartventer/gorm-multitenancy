# Multitenancy Implementation

This document outlines the approach for implementing multitenancy in the gorm-multitenancy package.

## Key Terms

- **Multitenancy**: A software architecture that allows multiple tenants (customers or users) to share the same application instance while ensuring data isolation.
- **Tenant**: A customer or user of the application, possessing a unique set of data.
- **Shared Data**: Data common to all tenants and shared across the application.
- **Tenant-Specific Data**: Data exclusive to a particular tenant, isolated from other tenants' data.
- **Schema**: A logical container for database objects (tables, views, etc.) in a database management system. PostgreSQL uses schemas, while MySQL uses databases for tenant isolation.
- **Migration**: The process of updating the database schema to reflect the application's latest data model.
- **Tenant Context**: The operational environment for executing tenant-specific operations, such as setting the schema or database for a specific tenant.

## Overview

This section details the essential operations for managing a multitenant architecture, facilitating tenant onboarding and offboarding, and managing shared and tenant-specific data structures.

### Default Shared Schema

Each database system has a default shared schema or database for storing shared data structures, accessible to all tenants. This shared infrastructure includes tables, views, and other database objects that are common across all tenants.

### Tenant Identifier

A unique identifier (e.g., `tenant_id`) distinguishes each tenant, facilitating the isolation of tenant-specific data and the management of tenant-specific operations. This identifier is used to create and manage tenant-specific schemas or databases and to query tenant-specific data.

## Operations

### Shared Operations

- **MigrateSharedModels**: Applies migrations to models shared across all tenants, ensuring the availability of shared infrastructure or common tables to all tenants.

### Tenant-Specific Operations

Operations requiring a tenant identifier for execution, enabling tenant-specific data management:

- **MigrateTenantModels**: Applies migrations to tenant-specific models, setting up or updating the necessary database structures for a tenant's data.
- **UseTenant**: Configures the schema or database for a specific tenant to perform tenant-specific operations. It includes a cleanup function that reverts to the default shared schema or database.
- **OffboardTenant**: Handles database cleanup when a tenant is removed, potentially involving the deletion of tenant-specific data and resource reclamation.

## Implementation Details

### PostgreSQL

| Feature | Description |
|---------|------------|
| Isolation Method | Utilizes schemas for tenant isolation. |
| Shared Data Storage | The `public` schema stores shared data structures. |
| Tenant-Specific Data Storage | Resides in a schema named after the tenant's identifier (e.g., `tenant_id`). |
| Data Access | The `search_path` is adjusted to the tenant's schema for querying tenant-specific data and is reset post-operation. |
| Transaction Management | Transactions are used to ensure atomicity and consistency during schema creation and migration. |
| Advisory Locks | Uses transaction-level advisory locks to prevent concurrent migrations that may interfere with each other. |

#### References

- [PostgreSQL Schemas](https://www.postgresql.org/docs/current/ddl-schemas.html)

#### MigrateSharedModels

```sql
-- Start transaction
BEGIN;

-- Acquire advisory lock to prevent concurrent migrations (automatically released at the end of the transaction)
SELECT pg_advisory_xact_lock(1);

-- Apply migrations
CREATE TABLE IF NOT EXISTS shared_table (id SERIAL PRIMARY KEY, name VARCHAR(255));

-- Commit transaction
COMMIT;
```

#### MigrateTenantModels

```sql
-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS tenant_id;

-- Start transaction
BEGIN;

-- Acquire advisory lock to prevent concurrent migrations (automatically released at the end of the transaction)
SELECT pg_advisory_xact_lock(1);

-- Set search path to tenant's schema
SET search_path TO tenant_id;

-- Apply migrations
CREATE TABLE IF NOT EXISTS tenant_specific_table (id SERIAL PRIMARY KEY, name VARCHAR(255));

-- Reset search path
RESET search_path;

-- Commit transaction
COMMIT;
```

#### OffboardTenant

```sql
-- Drop schema
DROP SCHEMA IF EXISTS tenant_id CASCADE;
```

#### UseTenant

```sql
-- Set search path to tenant's schema
SET search_path TO tenant_id;

-- Cleanup function to reset search path
RESET search_path;
```

### MySQL

| Feature | Description |
|---------|-------|
| Isolation Method | Employs databases for tenant isolation. |
| Shared Data Storage | The `public` database contains shared data structures. |
| Tenant-Specific Data Storage | Stored in a database named after the tenant's identifier (e.g., `tenant_id`). |
| Data Access | The database context is switched to the tenant's database for querying tenant-specific data and is reset post-operation. |
| Transaction Management | Transactions are used to ensure atomicity and consistency during schema creation and migration. |
| Advisory Locks | Used to prevent concurrent migrations that may interfere with each other. |

#### References

- [MySQL Databases](https://dev.mysql.com/doc/refman/8.0/en/create-database.html)
- [MySQL Use Database](https://dev.mysql.com/doc/refman/8.0/en/use.html)

#### MigrateSharedModels

```sql
-- Acquire advisory lock to prevent concurrent migrations
SELECT GET_LOCK('tenant_id', -1);

-- Start transaction
START TRANSACTION;

-- Apply migrations
CREATE TABLE IF NOT EXISTS shared_table (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255));

-- Commit transaction
COMMIT;

-- Release advisory lock
SELECT RELEASE_LOCK('tenant_id');
```

#### MigrateTenantModels

```sql
-- Check if database exists
SHOW DATABASES LIKE 'tenant_id';

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS tenant_id;

-- Acquire advisory lock to prevent concurrent migrations
SELECT GET_LOCK('tenant_id', -1);

-- Start transaction
START TRANSACTION;

-- Use database
USE tenant_id;

-- Apply migrations
CREATE TABLE IF NOT EXISTS tenant_specific_table (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255));

-- Reset database (switch to default database)
USE public;

-- Commit transaction
COMMIT;

-- Release advisory lock
SELECT RELEASE_LOCK('tenant_id');
```

#### OffboardTenant

```sql
-- Drop database
DROP DATABASE IF EXISTS tenant_id;
```

#### UseTenant

```sql
-- Use database
USE tenant_id;

-- Cleanup function to reset database
USE public;
```
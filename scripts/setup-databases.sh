#!/bin/bash

# Database Setup Script for Alibaba Cloud
# This script sets up all required databases and runs migrations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[SETUP]${NC} $1"
}

# Load environment variables
ENV_FILE="${1:-configs/generated/.env}"
if [ ! -f "$ENV_FILE" ]; then
    print_error "Environment file not found: $ENV_FILE"
    exit 1
fi

print_status "Loading environment variables from: $ENV_FILE"
set -a
source "$ENV_FILE"
set +a

print_header "Setting up databases for JobFirst services..."

# Function to check if MySQL is running
check_mysql() {
    print_status "Checking MySQL connection..."
    mysql -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" -p"${DB_PASSWORD}" -e "SELECT 1;" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        print_status "MySQL connection successful"
        return 0
    else
        print_error "MySQL connection failed"
        return 1
    fi
}

# Function to check if PostgreSQL is running
check_postgresql() {
    print_status "Checking PostgreSQL connection..."
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d postgres -c "SELECT 1;" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        print_status "PostgreSQL connection successful"
        return 0
    else
        print_error "PostgreSQL connection failed"
        return 1
    fi
}

# Function to check if MongoDB is running
check_mongodb() {
    print_status "Checking MongoDB connection..."
    mongo --host "${MONGODB_HOST}:${MONGODB_PORT}" --username "${MONGODB_USER}" --password "${MONGODB_PASSWORD}" --authenticationDatabase admin --eval "db.adminCommand('ping')" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        print_status "MongoDB connection successful"
        return 0
    else
        print_error "MongoDB connection failed"
        return 1
    fi
}

# Function to check if Redis is running
check_redis() {
    print_status "Checking Redis connection..."
    redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" -a "${REDIS_PASSWORD}" ping > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        print_status "Redis connection successful"
        return 0
    else
        print_error "Redis connection failed"
        return 1
    fi
}

# Setup MySQL database
setup_mysql() {
    print_header "Setting up MySQL database..."
    
    if check_mysql; then
        print_status "Creating MySQL database: ${DB_NAME}"
        mysql -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" -p"${DB_PASSWORD}" -e "CREATE DATABASE IF NOT EXISTS ${DB_NAME} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
        
        print_status "Running MySQL migrations..."
        if [ -f "database/database_migration_script.sql" ]; then
            mysql -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" -p"${DB_PASSWORD}" "${DB_NAME}" < database/database_migration_script.sql
            print_status "MySQL migrations completed"
        else
            print_warning "MySQL migration script not found: database/database_migration_script.sql"
        fi
    else
        print_error "Skipping MySQL setup due to connection failure"
    fi
}

# Setup PostgreSQL database
setup_postgresql() {
    print_header "Setting up PostgreSQL database..."
    
    if check_postgresql; then
        print_status "Creating PostgreSQL database: ${POSTGRES_DATABASE}"
        PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d postgres -c "CREATE DATABASE ${POSTGRES_DATABASE};" || print_warning "Database may already exist"
        
        print_status "PostgreSQL setup completed"
    else
        print_error "Skipping PostgreSQL setup due to connection failure"
    fi
}

# Setup MongoDB database
setup_mongodb() {
    print_header "Setting up MongoDB database..."
    
    if check_mongodb; then
        print_status "Creating MongoDB database: ${MONGODB_DATABASE}"
        mongo --host "${MONGODB_HOST}:${MONGODB_PORT}" --username "${MONGODB_USER}" --password "${MONGODB_PASSWORD}" --authenticationDatabase admin --eval "db = db.getSiblingDB('${MONGODB_DATABASE}'); db.createCollection('users'); db.createCollection('sessions');"
        
        print_status "MongoDB setup completed"
    else
        print_error "Skipping MongoDB setup due to connection failure"
    fi
}

# Setup Redis
setup_redis() {
    print_header "Setting up Redis..."
    
    if check_redis; then
        print_status "Testing Redis operations..."
        redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" -a "${REDIS_PASSWORD}" set "jobfirst:setup:test" "success" > /dev/null 2>&1
        redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" -a "${REDIS_PASSWORD}" get "jobfirst:setup:test" > /dev/null 2>&1
        redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" -a "${REDIS_PASSWORD}" del "jobfirst:setup:test" > /dev/null 2>&1
        
        print_status "Redis setup completed"
    else
        print_error "Skipping Redis setup due to connection failure"
    fi
}

# Main setup function
main() {
    print_header "Starting database setup for JobFirst services..."
    
    # Setup each database
    setup_mysql
    setup_postgresql
    setup_mongodb
    setup_redis
    
    print_header "Database setup completed!"
    
    # Final verification
    print_status "Performing final verification..."
    echo ""
    echo "Database Status:"
    echo "================"
    check_mysql && echo "✅ MySQL: Ready" || echo "❌ MySQL: Failed"
    check_postgresql && echo "✅ PostgreSQL: Ready" || echo "❌ PostgreSQL: Failed"
    check_mongodb && echo "✅ MongoDB: Ready" || echo "❌ MongoDB: Failed"
    check_redis && echo "✅ Redis: Ready" || echo "❌ Redis: Failed"
    
    print_status "Setup script completed!"
}

# Run main function
main "$@"

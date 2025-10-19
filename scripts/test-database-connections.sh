#!/bin/bash

# Database Connection Test Script
# This script tests all database connections to ensure they're working

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
    echo -e "${BLUE}[TEST]${NC} $1"
}

# Load environment variables
ENV_FILE="${1:-configs/.env}"
if [ ! -f "$ENV_FILE" ]; then
    print_error "Environment file not found: $ENV_FILE"
    exit 1
fi

print_status "Loading environment variables from: $ENV_FILE"
set -a
source "$ENV_FILE"
set +a

print_header "Testing database connections..."

# Test MySQL connection
test_mysql() {
    print_header "Testing MySQL connection..."
    echo "Host: ${DB_HOST}:${DB_PORT}"
    echo "User: ${DB_USER}"
    echo "Database: ${DB_NAME}"
    
    mysql -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" -p"${DB_PASSWORD}" -e "SELECT VERSION() as mysql_version, DATABASE() as current_database;" 2>/dev/null
    if [ $? -eq 0 ]; then
        print_status "✅ MySQL connection successful"
        return 0
    else
        print_error "❌ MySQL connection failed"
        return 1
    fi
}

# Test PostgreSQL connection
test_postgresql() {
    print_header "Testing PostgreSQL connection..."
    echo "Host: ${POSTGRES_HOST}:${POSTGRES_PORT}"
    echo "User: ${POSTGRES_USER}"
    echo "Database: ${POSTGRES_DATABASE}"
    
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DATABASE}" -c "SELECT version();" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        print_status "✅ PostgreSQL connection successful"
        return 0
    else
        print_error "❌ PostgreSQL connection failed"
        return 1
    fi
}

# Test MongoDB connection
test_mongodb() {
    print_header "Testing MongoDB connection..."
    echo "Host: ${MONGODB_HOST}:${MONGODB_PORT}"
    echo "User: ${MONGODB_USER}"
    echo "Database: ${MONGODB_DATABASE}"
    
    mongo --host "${MONGODB_HOST}:${MONGODB_PORT}" --username "${MONGODB_USER}" --password "${MONGODB_PASSWORD}" --authenticationDatabase admin --eval "db.adminCommand('ping')" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        print_status "✅ MongoDB connection successful"
        return 0
    else
        print_error "❌ MongoDB connection failed"
        return 1
    fi
}

# Test Redis connection
test_redis() {
    print_header "Testing Redis connection..."
    echo "Host: ${REDIS_HOST}:${REDIS_PORT}"
    
    result=$(redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" -a "${REDIS_PASSWORD}" ping 2>/dev/null)
    if [ "$result" = "PONG" ]; then
        print_status "✅ Redis connection successful"
        return 0
    else
        print_error "❌ Redis connection failed"
        return 1
    fi
}

# Main test function
main() {
    print_header "Starting database connection tests..."
    echo ""
    
    # Test each database
    mysql_result=0
    postgresql_result=0
    mongodb_result=0
    redis_result=0
    
    test_mysql || mysql_result=1
    echo ""
    test_postgresql || postgresql_result=1
    echo ""
    test_mongodb || mongodb_result=1
    echo ""
    test_redis || redis_result=1
    echo ""
    
    # Summary
    print_header "Database Connection Test Summary"
    echo "========================================"
    
    if [ $mysql_result -eq 0 ]; then
        echo "✅ MySQL: Connected"
    else
        echo "❌ MySQL: Failed"
    fi
    
    if [ $postgresql_result -eq 0 ]; then
        echo "✅ PostgreSQL: Connected"
    else
        echo "❌ PostgreSQL: Failed"
    fi
    
    if [ $mongodb_result -eq 0 ]; then
        echo "✅ MongoDB: Connected"
    else
        echo "❌ MongoDB: Failed"
    fi
    
    if [ $redis_result -eq 0 ]; then
        echo "✅ Redis: Connected"
    else
        echo "❌ Redis: Failed"
    fi
    
    echo "========================================"
    
    # Overall result
    total_failures=$((mysql_result + postgresql_result + mongodb_result + redis_result))
    if [ $total_failures -eq 0 ]; then
        print_status "🎉 All database connections successful!"
        exit 0
    else
        print_error "⚠️  $total_failures database connection(s) failed"
        exit 1
    fi
}

# Run main function
main "$@"

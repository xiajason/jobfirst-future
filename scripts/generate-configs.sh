#!/bin/bash

# Configuration Generator Script
# This script generates environment-specific configuration files from templates

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check if environment file exists
ENV_FILE="${1:-configs/templates/aliyun.env.template}"
if [ ! -f "$ENV_FILE" ]; then
    print_error "Environment file not found: $ENV_FILE"
    exit 1
fi

print_status "Generating configuration files from: $ENV_FILE"

# Load environment variables
set -a
source "$ENV_FILE"
set +a

# Create output directory
OUTPUT_DIR="configs/generated"
mkdir -p "$OUTPUT_DIR"

# Function to process template file
process_template() {
    local template_file="$1"
    local output_file="$2"
    
    if [ ! -f "$template_file" ]; then
        print_warning "Template file not found: $template_file"
        return 1
    fi
    
    print_status "Processing template: $template_file -> $output_file"
    
    # Use envsubst to replace variables
    envsubst < "$template_file" > "$output_file"
    
    if [ $? -eq 0 ]; then
        print_status "Generated: $output_file"
    else
        print_error "Failed to generate: $output_file"
        return 1
    fi
}

# Generate configuration files for each service
services=(
    "user-service"
    "resume-service"
    "company-service"
    "notification-service"
    "template-service"
    "statistics-service"
    "banner-service"
    "dev-team-service"
    "job-service"
)

# Process each service template
for service in "${services[@]}"; do
    template_file="configs/templates/${service}-config.yaml.template"
    output_file="$OUTPUT_DIR/${service}-config.yaml"
    
    process_template "$template_file" "$output_file"
done

# Generate API Gateway configuration
process_template "configs/templates/api-gateway-config.yaml.template" "$OUTPUT_DIR/api-gateway-config.yaml"

# Generate environment file
print_status "Generating environment file: $OUTPUT_DIR/.env"
cp "$ENV_FILE" "$OUTPUT_DIR/.env"

print_status "Configuration generation completed!"
print_status "Generated files in: $OUTPUT_DIR"

# List generated files
echo ""
echo "Generated configuration files:"
ls -la "$OUTPUT_DIR"

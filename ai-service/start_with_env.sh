#!/bin/bash
cd /opt/services/ai-service-1/current

# Load environment variables from .env file
export $(grep -v '^#' .env | xargs)

# Activate virtual environment
source venv/bin/activate

# Start the service
python ai_service_with_zervigo.py

#!/bin/bash

echo "🛑 Stopping RuleStack development environment..."

docker-compose down

echo "✅ Development environment stopped."
echo ""
echo "💾 Data persisted in Docker volume 'rulestack_postgres_data'"
echo "   To remove all data: docker-compose down -v"
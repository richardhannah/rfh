#!/bin/bash

echo "🚀 Starting RuleStack development environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Create necessary directories
mkdir -p storage
mkdir -p tmp

# Start services
echo "🐳 Starting Docker services..."
docker-compose up --build -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
echo "   - Postgres"
docker-compose exec -T postgres pg_isready -U rulestack_user -d rulestack_dev

echo "   - API server"
timeout=60
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if curl -s http://localhost:8080/v1/health > /dev/null; then
        break
    fi
    sleep 2
    elapsed=$((elapsed + 2))
done

if [ $elapsed -ge $timeout ]; then
    echo "❌ API server did not start within $timeout seconds"
    docker-compose logs api
    exit 1
fi

echo "✅ Development environment is ready!"
echo ""
echo "📋 Services:"
echo "   🐘 Postgres:  localhost:5432"
echo "   🌐 API:       http://localhost:8080"
echo ""
echo "📊 Useful commands:"
echo "   View logs:    docker-compose logs -f"
echo "   Stop:         docker-compose down"
echo "   Rebuild:      docker-compose up --build"
echo ""
echo "🔧 API will automatically reload when you change Go files!"

# Show development token
echo ""
echo "🔑 Development setup:"
echo "   Run: go run ./scripts/setup-dev.go"
echo "   Then: ./rfh registry add local http://localhost:8080 dev-token-12345"
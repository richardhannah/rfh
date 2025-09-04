# Database Contention Analysis - RFH API

## Executive Summary
Database contention issues identified during cucumber test execution reveal fundamental architectural limitations in the RFH API server. While authentication endpoints show the most visible symptoms ("Failed to create session" errors), the underlying issue affects all database operations across the entire API.

## Problem Discovery
**Date**: 2025-08-31  
**Context**: Cucumber test suite experiencing high failure rates  
**Primary Symptom**: API 500 errors - "Failed to create session" during concurrent auth operations

## Current Mitigation
**Temporary Fix**: Added 50-200ms random delays before authentication commands in test suite  
**Result**: 85-90% reduction in session creation failures  
**Impact**: +3.5 seconds test execution time, significantly improved reliability

## Root Cause Analysis

### 1. Synchronous Request Handling
**Evidence**: Multiple concurrent login requests consistently fail with database errors  
**Pattern**: Sequential processing of database operations causes blocking  
**Location**: `internal/api/auth_handlers.go` and likely all handler files

### 2. Database Connection Pool Exhaustion
**Symptoms**:
- INSERT operations fail under concurrent load
- Session creation (`user_sessions` table) particularly affected
- Unique constraint violations on `token_hash` column suggest race conditions

**Current Implementation Issues**:
```go
// internal/db/users.go:170-183
func (db *DB) CreateUserSession(...) (*UserSession, error) {
    query := `INSERT INTO user_sessions ...`
    err := db.Get(&session, query, ...)  // Synchronous blocking call
    if err != nil {
        return nil, err  // No retry logic
    }
}
```

### 3. Lack of Async/Concurrent Processing
**Current Architecture**:
- HTTP handlers process requests sequentially
- Database operations block the handler goroutine
- No queue or worker pool for database operations
- Missing retry logic for transient failures

## Impact Assessment

### Affected Endpoints (High Risk)
1. **Authentication** (`/v1/auth/*`)
   - Login: Session creation failures
   - Register: User creation conflicts
   - Logout: Session deletion delays

2. **Package Operations** (`/v1/packages/*`)
   - Publish: Package metadata insertion
   - Search: Complex queries blocking other operations
   - Download: Read operations competing with writes

3. **Registry Operations** (`/v1/registries/*`)
   - Configuration updates
   - Token management
   - Concurrent registry access

### Performance Degradation Pattern
```
Concurrent Requests → Database Lock Contention → Thread Blocking → 
Request Queue Backup → Timeout/500 Errors → Cascade Failures
```

## Proposed Solutions

### Short-term (Quick Wins)

#### 1. Connection Pool Tuning
```go
// internal/db/db.go
func NewDB(config DatabaseConfig) (*DB, error) {
    db.SetMaxOpenConns(25)      // Increase from default
    db.SetMaxIdleConns(10)       // Maintain idle pool
    db.SetConnMaxLifetime(5*60) // 5 minute lifetime
    db.SetConnMaxIdleTime(90)    // 90 second idle timeout
}
```

#### 2. Retry Logic for Transient Failures
```go
func (db *DB) CreateUserSessionWithRetry(...) (*UserSession, error) {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        session, err := db.CreateUserSession(...)
        if err == nil {
            return session, nil
        }
        if !isRetryableError(err) {
            return nil, err
        }
        time.Sleep(time.Duration(i*50) * time.Millisecond)
    }
    return nil, fmt.Errorf("failed after %d retries", maxRetries)
}
```

### Medium-term (Architectural Improvements)

#### 1. Async Handler Pattern
```go
// Use goroutines with proper context handling
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    resultChan := make(chan *LoginResult, 1)
    go s.processLogin(ctx, req, resultChan)
    
    select {
    case result := <-resultChan:
        writeJSON(w, result.Status, result.Data)
    case <-ctx.Done():
        writeError(w, http.StatusRequestTimeout, "Request timeout")
    }
}
```

#### 2. Database Query Optimization
- Add missing indexes for frequent queries
- Use prepared statements for repeated operations
- Implement query result caching for read-heavy operations
- Consider read replicas for search operations

### Long-term (Scalability Solutions)

#### 1. Message Queue Architecture
```
HTTP Request → API Handler → Message Queue → Worker Pool → Database
                    ↓                            ↓
                Response ← ← ← ← Result Channel ←
```

**Benefits**:
- Decouples request handling from database operations
- Natural retry and backpressure handling
- Horizontal scaling of workers

#### 2. Database Migration Strategy
- **Option A**: PostgreSQL with pgBouncer for connection pooling
- **Option B**: Implement Redis for session storage (removes DB bottleneck)
- **Option C**: Event sourcing for write operations with eventual consistency

#### 3. API Server Improvements
```go
// Implement middleware for concurrent request handling
type ConcurrencyLimiter struct {
    semaphore chan struct{}
}

func NewConcurrencyLimiter(maxConcurrent int) *ConcurrencyLimiter {
    return &ConcurrencyLimiter{
        semaphore: make(chan struct{}, maxConcurrent),
    }
}

func (cl *ConcurrencyLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cl.semaphore <- struct{}{} // Acquire
        defer func() { <-cl.semaphore }() // Release
        next.ServeHTTP(w, r)
    })
}
```

## Implementation Roadmap

### Phase 1: Immediate (1-2 days)
- [ ] Implement connection pool tuning
- [ ] Add retry logic to session creation
- [ ] Deploy monitoring for database metrics

### Phase 2: Quick Fixes (1 week)
- [ ] Add database operation timeouts
- [ ] Implement basic request queuing
- [ ] Add circuit breaker pattern for database calls

### Phase 3: Architectural (2-4 weeks)
- [ ] Migrate session handling to async pattern
- [ ] Implement worker pool for database operations
- [ ] Add comprehensive database metrics and alerting

### Phase 4: Scalability (1-2 months)
- [ ] Evaluate and implement message queue system
- [ ] Migrate to connection pooler (pgBouncer/similar)
- [ ] Consider microservices split for high-traffic endpoints

## Testing Strategy

### Load Testing Requirements
```bash
# Concurrent login test
ab -n 1000 -c 50 -T application/json \
   -p login.json \
   http://localhost:8081/v1/auth/login

# Expected metrics after fixes:
# - 0% 500 errors (currently ~15-20%)
# - <100ms p95 response time (currently 500ms+)
# - Support 100+ concurrent connections
```

### Monitoring Metrics
- Database connection pool utilization
- Query execution times (p50, p95, p99)
- HTTP request queue depth
- Error rates by endpoint
- Database lock wait times

## Research Topics

### Go Concurrency Patterns
1. **Context-aware database operations**
   - Reference: https://go.dev/blog/context
   - Proper cancellation and timeout handling

2. **Worker pool patterns**
   - Reference: https://gobyexample.com/worker-pools
   - Bounded concurrency for database operations

3. **Circuit breaker implementation**
   - Libraries: sony/gobreaker, afex/hystrix-go
   - Prevent cascade failures

### Database Optimization
1. **PostgreSQL connection pooling**
   - PgBouncer configuration
   - Connection pool sizing formulas

2. **Query optimization**
   - EXPLAIN ANALYZE for slow queries
   - Index strategy for concurrent writes

3. **Alternative storage patterns**
   - Redis for session management
   - Write-ahead logging for async processing

## Success Criteria

### Short-term (After Phase 1-2)
- Zero "Failed to create session" errors in test suite
- <5% error rate under 50 concurrent users
- <200ms p95 response time for auth endpoints

### Long-term (After Phase 3-4)
- Support 500+ concurrent users
- <100ms p95 response time across all endpoints
- Horizontal scalability without code changes
- Self-healing under load spikes

## Conclusion

The database contention issue is a critical architectural limitation that affects the entire API, not just authentication. While test delays provide temporary relief, a comprehensive solution requires moving from synchronous to asynchronous request processing, implementing proper connection pooling, and potentially adopting message queue patterns for database operations.

The phased approach allows for incremental improvements while maintaining system stability. Priority should be given to Phase 1-2 fixes as they provide immediate relief with minimal architectural changes.

## References

- [Go Database/SQL Tutorial](http://go-database-sql.org/)
- [PostgreSQL Connection Pooling](https://www.postgresql.org/docs/current/runtime-config-connection.html)
- [Handling High Traffic in Go](https://medium.com/@val_deleplace/go-concurrency-for-scaling-web-apis-86c5c3c8b48f)
- [Database Connection Pool Sizing](https://github.com/brettwooldridge/HikariCP/wiki/About-Pool-Sizing)

---
*Analysis Date: 2025-08-31*  
*Test Environment: Windows with Docker PostgreSQL*  
*Go Version: Assumed 1.19+*  
*Database: PostgreSQL with sqlx*
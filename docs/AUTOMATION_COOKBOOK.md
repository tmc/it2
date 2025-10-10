# it2 Automation Cookbook

Real-world automation recipes for common development and system administration tasks using the it2 CLI.

## Table of Contents

- [Development Workflows](#development-workflows)
- [Server Management](#server-management)
- [Database Operations](#database-operations)
- [CI/CD Integration](#cicd-integration)
- [Monitoring & Logging](#monitoring--logging)
- [Team Collaboration](#team-collaboration)
- [Advanced Automation](#advanced-automation)

## Development Workflows

### Recipe: Full-Stack Development Environment

Create a complete development environment with frontend, backend, and database monitoring.

```bash
#!/bin/bash
# setup-fullstack-env.sh
# Creates a comprehensive development environment

PROJECT_NAME="MyApp"
PROJECT_DIR="$HOME/projects/myapp"

echo "🚀 Setting up $PROJECT_NAME development environment..."

# Create main project tab
MAIN_TAB=$(it2 tab create "Development" --badge "Main" --format json | jq -r '.id')
it2 session send-text "cd $PROJECT_DIR && git status"

# Split for editor (keep main session for git operations)
EDITOR_SESSION=$(it2 session split --vertical --profile "Development" --quiet)
it2 session send-text "$EDITOR_SESSION" "cd $PROJECT_DIR && code ."

# Split for backend server
BACKEND_SESSION=$(it2 session split --horizontal --profile "Development" --quiet)
it2 session send-text "$BACKEND_SESSION" "cd $PROJECT_DIR/backend && npm run dev"

# Create separate tab for frontend
FRONTEND_TAB=$(it2 tab create "Development" --badge "Frontend" --format json | jq -r '.id')
it2 session send-text "cd $PROJECT_DIR/frontend && npm start"

# Split frontend tab for hot reloading logs
FRONTEND_LOGS=$(it2 session split --horizontal --profile "Development" --quiet)
it2 session send-text "$FRONTEND_LOGS" "cd $PROJECT_DIR/frontend && npm run test:watch"

# Create database monitoring tab
DB_TAB=$(it2 tab create "Database" --badge "DB" --format json | jq -r '.id')
it2 session send-text "cd $PROJECT_DIR && docker-compose logs -f postgres"

# Split for database shell access
DB_SHELL=$(it2 session split --vertical --profile "Database" --quiet)
it2 session send-text "$DB_SHELL" "cd $PROJECT_DIR && make db-shell"

# Create testing tab
TEST_TAB=$(it2 tab create "Testing" --badge "Tests" --format json | jq -r '.id')
it2 session send-text "cd $PROJECT_DIR && npm run test:coverage"

# Split for E2E tests
E2E_SESSION=$(it2 session split --horizontal --profile "Testing" --quiet)
it2 session send-text "$E2E_SESSION" "cd $PROJECT_DIR && npm run test:e2e:watch"

echo "✅ Development environment ready!"
echo "   - Main: Git operations and project overview"
echo "   - Frontend: React dev server and test runner"
echo "   - Database: Logs and shell access"
echo "   - Testing: Unit tests and E2E tests"
```

### Recipe: Git Workflow Automation

Automate common Git workflows across multiple repositories.

```bash
#!/bin/bash
# git-workflow-automation.sh
# Manages Git operations across multiple repositories

REPOS=("frontend" "backend" "mobile" "infrastructure")
BASE_DIR="$HOME/projects"

# Create Git workflow tab
GIT_TAB=$(it2 tab create "Git Workflow" --badge "Git" --format json | jq -r '.id')

# Function to create repository session
create_repo_session() {
    local repo=$1
    local is_first=$2

    if [ "$is_first" = "true" ]; then
        # Use main session for first repo
        SESSION_ID=$(it2 session current)
    else
        # Split for additional repos
        SESSION_ID=$(it2 session split --horizontal --profile "Development" --quiet)
    fi

    # Set up the repository session
    it2 session send-text "$SESSION_ID" "cd $BASE_DIR/$repo"
    it2 session send-text "$SESSION_ID" "git status --porcelain"
    it2 session send-text "$SESSION_ID" "echo '=== $repo Repository ==='"

    echo "$SESSION_ID"
}

echo "🔀 Setting up Git workflow for ${#REPOS[@]} repositories..."

# Create sessions for each repository
first=true
for repo in "${REPOS[@]}"; do
    session_id=$(create_repo_session "$repo" "$first")
    echo "📁 $repo: $session_id"
    first=false
done

# Create Git command helper
cat > /tmp/git-all-repos.sh << 'EOF'
#!/bin/bash
# Helper script for running Git commands across all repositories

command="$1"
if [ -z "$command" ]; then
    echo "Usage: git-all-repos.sh <git-command>"
    exit 1
fi

# Get all sessions in current tab
sessions=$(it2 session list --format json | jq -r '.[] | select(.tab_title | contains("Git Workflow")) | .id')

for session in $sessions; do
    echo "Executing '$command' in session $session"
    it2 session send-text "$session" "git $command"
done
EOF

chmod +x /tmp/git-all-repos.sh

echo "✅ Git workflow ready!"
echo "   Use: /tmp/git-all-repos.sh 'status' to check all repos"
echo "   Use: /tmp/git-all-repos.sh 'pull' to update all repos"
```

## Server Management

### Recipe: Multi-Server Monitoring Dashboard

Create a comprehensive server monitoring setup.

```bash
#!/bin/bash
# server-monitoring.sh
# Creates a multi-server monitoring dashboard

SERVERS=("web01.prod" "web02.prod" "db01.prod" "cache01.prod")
ENVIRONMENTS=("production" "staging" "development")

# Create monitoring tab
MONITOR_TAB=$(it2 tab create "Monitoring" --badge "🖥️ Servers" --format json | jq -r '.id')

# Function to create server monitoring session
setup_server_monitoring() {
    local server=$1
    local is_first=$2

    if [ "$is_first" = "true" ]; then
        SESSION_ID=$(it2 session current)
    else
        SESSION_ID=$(it2 session split --horizontal --profile "SSH" --quiet)
    fi

    # Connect and start monitoring
    it2 session send-text "$SESSION_ID" "ssh $server"
    sleep 2  # Wait for connection
    it2 session send-text "$SESSION_ID" "htop"

    echo "$SESSION_ID"
}

echo "🖥️ Setting up server monitoring dashboard..."

first=true
for server in "${SERVERS[@]}"; do
    session_id=$(setup_server_monitoring "$server" "$first")
    echo "📊 Monitoring $server: $session_id"
    first=false
done

# Create system status tab
STATUS_TAB=$(it2 tab create "System Status" --badge "📈 Status" --format json | jq -r '.id')

# Main system overview
it2 session send-text "ssh web01.prod"
sleep 2
it2 session send-text "watch -n 5 'uptime; echo; df -h; echo; free -h'"

# Split for network monitoring
NETWORK_SESSION=$(it2 session split --vertical --profile "SSH" --quiet)
it2 session send-text "$NETWORK_SESSION" "ssh web01.prod"
sleep 2
it2 session send-text "$NETWORK_SESSION" "watch -n 10 'ss -tuln | head -20'"

# Split for log monitoring
LOG_SESSION=$(it2 session split --horizontal --profile "SSH" --quiet)
it2 session send-text "$LOG_SESSION" "ssh web01.prod"
sleep 2
it2 session send-text "$LOG_SESSION" "sudo tail -f /var/log/nginx/access.log"

echo "✅ Server monitoring dashboard ready!"
```

### Recipe: Deployment Pipeline Monitoring

Monitor deployment across multiple environments.

```bash
#!/bin/bash
# deployment-monitoring.sh
# Monitors deployments across environments

ENVIRONMENTS=("staging" "production")
SERVICES=("api" "web" "worker")

# Create deployment monitoring tab
DEPLOY_TAB=$(it2 tab create "Deployment" --badge "🚀 Deploy" --format json | jq -r '.id')

# Function to monitor deployment
monitor_deployment() {
    local env=$1
    local service=$2
    local session_id=$3

    # Monitor deployment logs
    it2 session send-text "$session_id" "kubectl logs -f deployment/$service-$env -n $env"
}

echo "🚀 Setting up deployment monitoring..."

# Create sessions for each environment
for env in "${ENVIRONMENTS[@]}"; do
    echo "Setting up $env environment monitoring..."

    # Create tab for environment
    ENV_TAB=$(it2 tab create "Deploy-$env" --badge "$env" --format json | jq -r '.id')

    # Main session: deployment status
    it2 session send-text "kubectl get pods -n $env -w"

    first_service=true
    for service in "${SERVICES[@]}"; do
        if [ "$first_service" = "false" ]; then
            SERVICE_SESSION=$(it2 session split --horizontal --profile "Kubernetes" --quiet)
        else
            SERVICE_SESSION=$(it2 session current)
            first_service=false
        fi

        monitor_deployment "$env" "$service" "$SERVICE_SESSION"
    done
done

# Create deployment control tab
CONTROL_TAB=$(it2 tab create "Deploy Control" --badge "🎛️ Control" --format json | jq -r '.id')

# Deployment commands session
cat > /tmp/deploy-commands.sh << 'EOF'
#!/bin/bash
# Deployment command helpers

deploy_staging() {
    echo "🚀 Deploying to staging..."
    kubectl apply -f k8s/staging/ -n staging
    kubectl rollout status deployment/api-staging -n staging
}

deploy_production() {
    echo "🚀 Deploying to production..."
    read -p "Are you sure you want to deploy to production? (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
        kubectl apply -f k8s/production/ -n production
        kubectl rollout status deployment/api-production -n production
    fi
}

rollback_staging() {
    echo "⏪ Rolling back staging..."
    kubectl rollout undo deployment/api-staging -n staging
}

rollback_production() {
    echo "⏪ Rolling back production..."
    read -p "Are you sure you want to rollback production? (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
        kubectl rollout undo deployment/api-production -n production
    fi
}

# Show available commands
echo "Available commands:"
echo "  deploy_staging    - Deploy to staging"
echo "  deploy_production - Deploy to production"
echo "  rollback_staging  - Rollback staging"
echo "  rollback_production - Rollback production"
EOF

it2 session send-text "source /tmp/deploy-commands.sh"

echo "✅ Deployment monitoring ready!"
```

## Database Operations

### Recipe: Multi-Database Administration

Manage multiple databases with monitoring and backup operations.

```bash
#!/bin/bash
# database-admin.sh
# Multi-database administration setup

DATABASES=("users_db" "orders_db" "analytics_db")
DB_HOST="db.example.com"

# Create database administration tab
DB_TAB=$(it2 tab create "Database Admin" --badge "💾 DB" --format json | jq -r '.id')

echo "💾 Setting up database administration environment..."

# Main session: database monitoring
it2 session send-text "# Database Administration Dashboard"
it2 session send-text "watch -n 30 'echo \"=== Database Status ===\"; for db in ${DATABASES[@]}; do echo \"$db:\"; psql -h $DB_HOST -d \$db -c \"SELECT count(*) as active_connections FROM pg_stat_activity WHERE datname='\$db';\"; done'"

# Split for SQL console
SQL_SESSION=$(it2 session split --vertical --profile "Database" --quiet)
it2 session send-text "$SQL_SESSION" "psql -h $DB_HOST -d ${DATABASES[0]}"

# Split for backup operations
BACKUP_SESSION=$(it2 session split --horizontal --profile "Database" --quiet)

# Create backup script
cat > /tmp/db-backup.sh << 'EOF'
#!/bin/bash
# Database backup utilities

backup_all_databases() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_dir="/backups/$timestamp"

    mkdir -p "$backup_dir"

    for db in "$@"; do
        echo "📦 Backing up $db..."
        pg_dump -h "$DB_HOST" -d "$db" | gzip > "$backup_dir/${db}_${timestamp}.sql.gz"
        echo "✅ $db backup complete"
    done

    echo "🎉 All backups complete in $backup_dir"
}

restore_database() {
    local db=$1
    local backup_file=$2

    echo "⚠️  Restoring $db from $backup_file"
    read -p "This will overwrite $db. Continue? (yes/no): " confirm

    if [ "$confirm" = "yes" ]; then
        echo "🔄 Restoring $db..."
        gunzip -c "$backup_file" | psql -h "$DB_HOST" -d "$db"
        echo "✅ Restore complete"
    fi
}

# Show available functions
echo "Database backup utilities loaded:"
echo "  backup_all_databases db1 db2 db3"
echo "  restore_database db_name backup_file.sql.gz"
EOF

it2 session send-text "$BACKUP_SESSION" "source /tmp/db-backup.sh"

# Create performance monitoring tab
PERF_TAB=$(it2 tab create "DB Performance" --badge "📊 Perf" --format json | jq -r '.id')

# Query performance monitoring
it2 session send-text "psql -h $DB_HOST -d ${DATABASES[0]}"
it2 session send-text "\\timing on"
it2 session send-text "-- Database Performance Dashboard"

# Split for slow query monitoring
SLOW_QUERY_SESSION=$(it2 session split --horizontal --profile "Database" --quiet)
it2 session send-text "$SLOW_QUERY_SESSION" "# Monitoring slow queries..."
it2 session send-text "$SLOW_QUERY_SESSION" "tail -f /var/log/postgresql/postgresql-slow.log"

echo "✅ Database administration environment ready!"
```

## CI/CD Integration

### Recipe: Jenkins Pipeline Monitoring

Monitor Jenkins builds and deployments.

```bash
#!/bin/bash
# jenkins-monitoring.sh
# Monitor Jenkins CI/CD pipelines

JENKINS_URL="https://jenkins.example.com"
PROJECTS=("frontend" "backend" "mobile-app")

# Create Jenkins monitoring tab
JENKINS_TAB=$(it2 tab create "Jenkins CI/CD" --badge "🔧 CI/CD" --format json | jq -r '.id')

echo "🔧 Setting up Jenkins pipeline monitoring..."

# Function to monitor project builds
monitor_project_builds() {
    local project=$1
    local session_id=$2

    # Monitor build logs
    it2 session send-text "$session_id" "# Monitoring $project builds"
    it2 session send-text "$session_id" "watch -n 30 'curl -s \"$JENKINS_URL/job/$project/lastBuild/api/json\" | jq \".number, .result, .timestamp\"'"
}

# Main session: overall build status
it2 session send-text "# Jenkins Pipeline Dashboard"
it2 session send-text "watch -n 60 'echo \"=== Build Status ===\"; for project in ${PROJECTS[@]}; do echo \"$project:\"; curl -s \"$JENKINS_URL/job/\$project/lastBuild/api/json\" | jq -r \"\(.number) - \(.result // \\\"RUNNING\\\")\"; done'"

# Create sessions for each project
first=true
for project in "${PROJECTS[@]}"; do
    if [ "$first" = "false" ]; then
        PROJECT_SESSION=$(it2 session split --horizontal --profile "Development" --quiet)
    else
        PROJECT_SESSION=$(it2 session current)
        first=false
    fi

    monitor_project_builds "$project" "$PROJECT_SESSION"
done

# Create Jenkins control tab
CONTROL_TAB=$(it2 tab create "Jenkins Control" --badge "🎛️ Control" --format json | jq -r '.id')

# Jenkins control functions
cat > /tmp/jenkins-control.sh << 'EOF'
#!/bin/bash
# Jenkins control utilities

trigger_build() {
    local project=$1
    echo "🚀 Triggering build for $project..."
    curl -X POST "$JENKINS_URL/job/$project/build"
    echo "✅ Build triggered"
}

get_build_status() {
    local project=$1
    echo "📊 Getting status for $project..."
    curl -s "$JENKINS_URL/job/$project/lastBuild/api/json" | jq '{number, result, timestamp, duration}'
}

get_build_log() {
    local project=$1
    local build_number=${2:-"lastBuild"}
    echo "📋 Getting build log for $project #$build_number..."
    curl -s "$JENKINS_URL/job/$project/$build_number/consoleText"
}

# Show available commands
echo "Jenkins control utilities:"
echo "  trigger_build <project>"
echo "  get_build_status <project>"
echo "  get_build_log <project> [build_number]"
EOF

it2 session send-text "source /tmp/jenkins-control.sh"

echo "✅ Jenkins monitoring setup complete!"
```

## Monitoring & Logging

### Recipe: Comprehensive Log Monitoring

Monitor logs across multiple services and environments.

```bash
#!/bin/bash
# log-monitoring.sh
# Comprehensive log monitoring setup

SERVICES=("nginx" "api" "worker" "database")
LOG_LEVELS=("ERROR" "WARN" "INFO")

# Create log monitoring tab
LOG_TAB=$(it2 tab create "Log Monitoring" --badge "📋 Logs" --format json | jq -r '.id')

echo "📋 Setting up comprehensive log monitoring..."

# Function to setup log monitoring for a service
setup_service_logs() {
    local service=$1
    local session_id=$2

    case $service in
        "nginx")
            it2 session send-text "$session_id" "tail -f /var/log/nginx/error.log | grep --color=always -E 'ERROR|WARN|$'"
            ;;
        "api")
            it2 session send-text "$session_id" "kubectl logs -f deployment/api --tail=100 | grep --color=always -E 'ERROR|WARN|$'"
            ;;
        "worker")
            it2 session send-text "$session_id" "docker logs -f worker_container | grep --color=always -E 'ERROR|WARN|$'"
            ;;
        "database")
            it2 session send-text "$session_id" "tail -f /var/log/postgresql/postgresql.log | grep --color=always -E 'ERROR|WARN|$'"
            ;;
    esac
}

# Main session: aggregated error monitoring
it2 session send-text "# Aggregated Error Monitoring"
it2 session send-text "watch -n 10 'echo \"=== Error Summary (last 5 minutes) ===\"; find /var/log -name \"*.log\" -exec grep -l \"ERROR\" {} \\; | head -10'"

# Create sessions for each service
first=true
for service in "${SERVICES[@]}"; do
    if [ "$first" = "false" ]; then
        SERVICE_SESSION=$(it2 session split --horizontal --profile "Monitoring" --quiet)
    else
        SERVICE_SESSION=$(it2 session current)
        first=false
    fi

    setup_service_logs "$service" "$SERVICE_SESSION"
    echo "📊 Monitoring $service logs: $SERVICE_SESSION"
done

# Create log analysis tab
ANALYSIS_TAB=$(it2 tab create "Log Analysis" --badge "🔍 Analysis" --format json | jq -r '.id')

# Log analysis tools
cat > /tmp/log-analysis.sh << 'EOF'
#!/bin/bash
# Log analysis utilities

analyze_errors() {
    local hours=${1:-1}
    echo "🔍 Analyzing errors from last $hours hour(s)..."

    find /var/log -name "*.log" -type f -mmin -$((hours * 60)) \
        -exec grep -H "ERROR" {} \; | \
        awk -F: '{print $3}' | \
        sort | uniq -c | sort -rn | head -20
}

analyze_traffic() {
    local hours=${1:-1}
    echo "📊 Analyzing traffic from last $hours hour(s)..."

    tail -n 10000 /var/log/nginx/access.log | \
        awk '{print $1}' | sort | uniq -c | sort -rn | head -20
}

search_logs() {
    local pattern=$1
    local hours=${2:-1}
    echo "🔎 Searching for '$pattern' in last $hours hour(s)..."

    find /var/log -name "*.log" -type f -mmin -$((hours * 60)) \
        -exec grep -H "$pattern" {} \;
}

# Show available commands
echo "Log analysis utilities:"
echo "  analyze_errors [hours]     - Show error frequency"
echo "  analyze_traffic [hours]    - Show traffic patterns"
echo "  search_logs <pattern> [hours] - Search for pattern"
EOF

it2 session send-text "source /tmp/log-analysis.sh"

# Split for real-time log search
SEARCH_SESSION=$(it2 session split --vertical --profile "Monitoring" --quiet)
it2 session send-text "$SEARCH_SESSION" "# Real-time log search"
it2 session send-text "$SEARCH_SESSION" "# Use: tail -f /var/log/app.log | grep 'pattern'"

echo "✅ Log monitoring setup complete!"
```

## Team Collaboration

### Recipe: Code Review Dashboard

Setup for efficient code review workflows.

```bash
#!/bin/bash
# code-review-dashboard.sh
# Setup collaborative code review environment

REPOS=("frontend" "backend" "mobile")
GITHUB_USER="your-username"

# Create code review tab
REVIEW_TAB=$(it2 tab create "Code Review" --badge "👀 Review" --format json | jq -r '.id')

echo "👀 Setting up code review dashboard..."

# Function to setup repository for review
setup_repo_review() {
    local repo=$1
    local session_id=$2

    it2 session send-text "$session_id" "cd $HOME/projects/$repo"
    it2 session send-text "$session_id" "# $repo Repository Review"
    it2 session send-text "$session_id" "git fetch --all"
    it2 session send-text "$session_id" "git log --oneline -10"
}

# Main session: pull request overview
it2 session send-text "# Pull Request Dashboard"
it2 session send-text "gh pr list --state open"

# Create sessions for each repository
first=true
for repo in "${REPOS[@]}"; do
    if [ "$first" = "false" ]; then
        REPO_SESSION=$(it2 session split --horizontal --profile "Development" --quiet)
    else
        REPO_SESSION=$(it2 session current)
        first=false
    fi

    setup_repo_review "$repo" "$REPO_SESSION"
    echo "📁 $repo review session: $REPO_SESSION"
done

# Create PR review tools tab
TOOLS_TAB=$(it2 tab create "Review Tools" --badge "🛠️ Tools" --format json | jq -r '.id')

# Review helper functions
cat > /tmp/review-tools.sh << 'EOF'
#!/bin/bash
# Code review utilities

review_pr() {
    local pr_number=$1
    local repo=${2:-$(basename $(git rev-parse --show-toplevel))}

    echo "👀 Reviewing PR #$pr_number in $repo..."

    # Checkout PR branch
    gh pr checkout "$pr_number"

    # Show PR details
    gh pr view "$pr_number"

    # Show diff
    git diff HEAD~1
}

approve_pr() {
    local pr_number=$1
    local message=${2:-"LGTM! ✅"}

    echo "✅ Approving PR #$pr_number..."
    gh pr review "$pr_number" --approve --body "$message"
}

request_changes() {
    local pr_number=$1
    local message=$2

    echo "❌ Requesting changes on PR #$pr_number..."
    gh pr review "$pr_number" --request-changes --body "$message"
}

merge_pr() {
    local pr_number=$1

    echo "🔀 Merging PR #$pr_number..."
    read -p "Are you sure you want to merge? (yes/no): " confirm

    if [ "$confirm" = "yes" ]; then
        gh pr merge "$pr_number" --squash --delete-branch
        echo "✅ PR merged and branch deleted"
    fi
}

# Show available commands
echo "Code review utilities:"
echo "  review_pr <number> [repo]     - Review a pull request"
echo "  approve_pr <number> [message] - Approve a pull request"
echo "  request_changes <number> <message> - Request changes"
echo "  merge_pr <number>             - Merge a pull request"
EOF

it2 session send-text "source /tmp/review-tools.sh"

# Split for testing changes
TEST_SESSION=$(it2 session split --horizontal --profile "Development" --quiet)
it2 session send-text "$TEST_SESSION" "# Testing PR changes"
it2 session send-text "$TEST_SESSION" "# Use this session to run tests on PR branches"

echo "✅ Code review dashboard ready!"
```

## Advanced Automation

### Recipe: Incident Response Automation

Automate incident response procedures.

```bash
#!/bin/bash
# incident-response.sh
# Automated incident response setup

SERVICES=("web" "api" "database" "cache")
ALERT_CHANNELS=("#alerts" "#ops")

# Create incident response tab
INCIDENT_TAB=$(it2 tab create "Incident Response" --badge "🚨 Incident" --format json | jq -r '.id')

echo "🚨 Setting up incident response automation..."

# Main session: incident overview
it2 session send-text "# Incident Response Dashboard"
it2 session send-text "echo '🚨 Incident Response Active - $(date)'"

# Function to setup service monitoring
setup_service_incident_monitoring() {
    local service=$1
    local session_id=$2

    it2 session send-text "$session_id" "# Monitoring $service for incidents"

    case $service in
        "web")
            it2 session send-text "$session_id" "watch -n 10 'curl -s -o /dev/null -w \"%{http_code}\" https://app.example.com || echo \"❌ Web service down\"'"
            ;;
        "api")
            it2 session send-text "$session_id" "watch -n 10 'curl -s https://api.example.com/health | jq . || echo \"❌ API service down\"'"
            ;;
        "database")
            it2 session send-text "$session_id" "watch -n 30 'pg_isready -h db.example.com || echo \"❌ Database unreachable\"'"
            ;;
        "cache")
            it2 session send-text "$session_id" "watch -n 30 'redis-cli -h cache.example.com ping || echo \"❌ Cache unreachable\"'"
            ;;
    esac
}

# Create monitoring sessions for each service
first=true
for service in "${SERVICES[@]}"; do
    if [ "$first" = "false" ]; then
        SERVICE_SESSION=$(it2 session split --horizontal --profile "Monitoring" --quiet)
    else
        SERVICE_SESSION=$(it2 session current)
        first=false
    fi

    setup_service_incident_monitoring "$service" "$SERVICE_SESSION"
    echo "🔍 Monitoring $service: $SERVICE_SESSION"
done

# Create incident response tools tab
TOOLS_TAB=$(it2 tab create "Response Tools" --badge "🛠️ Response" --format json | jq -r '.id')

# Incident response automation
cat > /tmp/incident-response.sh << 'EOF'
#!/bin/bash
# Incident response automation tools

declare_incident() {
    local severity=$1
    local description=$2
    local timestamp=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

    echo "🚨 INCIDENT DECLARED - Severity: $severity"
    echo "Time: $timestamp"
    echo "Description: $description"

    # Create incident channel
    slack_channel="#incident-$(date +%Y%m%d-%H%M)"
    echo "📱 Incident channel: $slack_channel"

    # Notify stakeholders
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"🚨 INCIDENT: $description (Severity: $severity)\"}" \
        "$SLACK_WEBHOOK_URL"

    # Log incident
    echo "$timestamp | $severity | $description" >> /var/log/incidents.log
}

scale_service() {
    local service=$1
    local replicas=$2

    echo "⚡ Scaling $service to $replicas replicas..."
    kubectl scale deployment "$service" --replicas="$replicas"
    kubectl rollout status deployment/"$service"
    echo "✅ $service scaled to $replicas replicas"
}

failover_service() {
    local service=$1

    echo "🔄 Initiating failover for $service..."
    kubectl patch service "$service" -p '{"spec":{"selector":{"version":"backup"}}}'
    echo "✅ $service failed over to backup"
}

gather_diagnostics() {
    local service=$1
    local output_file="/tmp/diagnostics-$service-$(date +%s).log"

    echo "🔍 Gathering diagnostics for $service..."

    {
        echo "=== Service Status ==="
        kubectl get pods -l app="$service" -o wide

        echo "=== Recent Logs ==="
        kubectl logs -l app="$service" --tail=100

        echo "=== Resource Usage ==="
        kubectl top pods -l app="$service"

        echo "=== Events ==="
        kubectl get events --field-selector involvedObject.name="$service"

    } > "$output_file"

    echo "📋 Diagnostics saved to $output_file"
}

# Show available commands
echo "Incident response tools:"
echo "  declare_incident <severity> <description>"
echo "  scale_service <service> <replicas>"
echo "  failover_service <service>"
echo "  gather_diagnostics <service>"
EOF

it2 session send-text "source /tmp/incident-response.sh"

# Split for communication
COMM_SESSION=$(it2 session split --vertical --profile "Communication" --quiet)
it2 session send-text "$COMM_SESSION" "# Incident Communication"
it2 session send-text "$COMM_SESSION" "echo 'Use this session for stakeholder communication'"

echo "✅ Incident response automation ready!"
```

### Recipe: Performance Testing Automation

Automate performance testing and monitoring.

```bash
#!/bin/bash
# performance-testing.sh
# Automated performance testing setup

TEST_ENVIRONMENTS=("staging" "load-test")
ENDPOINTS=("/api/users" "/api/orders" "/api/products")

# Create performance testing tab
PERF_TAB=$(it2 tab create "Performance Tests" --badge "⚡ Perf" --format json | jq -r '.id')

echo "⚡ Setting up performance testing automation..."

# Main session: test overview
it2 session send-text "# Performance Testing Dashboard"
it2 session send-text "echo '⚡ Performance Testing Started - $(date)'"

# Function to setup load testing
setup_load_testing() {
    local endpoint=$1
    local session_id=$2

    it2 session send-text "$session_id" "# Load testing $endpoint"
    it2 session send-text "$session_id" "wrk -t12 -c400 -d30s --latency https://staging.example.com$endpoint"
}

# Create load testing sessions
first=true
for endpoint in "${ENDPOINTS[@]}"; do
    if [ "$first" = "false" ]; then
        ENDPOINT_SESSION=$(it2 session split --horizontal --profile "Testing" --quiet)
    else
        ENDPOINT_SESSION=$(it2 session current)
        first=false
    fi

    setup_load_testing "$endpoint" "$ENDPOINT_SESSION"
    echo "🎯 Load testing $endpoint: $ENDPOINT_SESSION"
done

# Create monitoring tab for performance metrics
METRICS_TAB=$(it2 tab create "Performance Metrics" --badge "📊 Metrics" --format json | jq -r '.id')

# System monitoring during tests
it2 session send-text "# System Performance Monitoring"
it2 session send-text "watch -n 5 'echo \"=== CPU & Memory ===\"; top -bn1 | head -20; echo; echo \"=== Network ===\"; netstat -i'"

# Split for application metrics
APP_METRICS_SESSION=$(it2 session split --vertical --profile "Monitoring" --quiet)
it2 session send-text "$APP_METRICS_SESSION" "# Application Metrics"
it2 session send-text "$APP_METRICS_SESSION" "watch -n 10 'curl -s https://staging.example.com/metrics | grep -E \"response_time|request_count|error_rate\"'"

# Split for database performance
DB_METRICS_SESSION=$(it2 session split --horizontal --profile "Database" --quiet)
it2 session send-text "$DB_METRICS_SESSION" "# Database Performance"
it2 session send-text "$DB_METRICS_SESSION" "watch -n 15 'psql -h db.staging.example.com -c \"SELECT query, calls, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;\"'"

# Create test automation tools
TOOLS_TAB=$(it2 tab create "Test Tools" --badge "🔧 Tools" --format json | jq -r '.id')

# Performance testing utilities
cat > /tmp/perf-testing.sh << 'EOF'
#!/bin/bash
# Performance testing utilities

run_load_test() {
    local url=$1
    local duration=${2:-"30s"}
    local connections=${3:-"100"}
    local threads=${4:-"12"}

    echo "⚡ Running load test on $url..."
    echo "Duration: $duration, Connections: $connections, Threads: $threads"

    wrk -t"$threads" -c"$connections" -d"$duration" --latency "$url" > "/tmp/load-test-$(date +%s).log"
    echo "📊 Load test complete. Results saved."
}

stress_test() {
    local url=$1
    local max_connections=${2:-"1000"}

    echo "🔥 Running stress test on $url..."

    for connections in 10 50 100 200 500 "$max_connections"; do
        echo "Testing with $connections connections..."
        wrk -t12 -c"$connections" -d10s "$url"
        sleep 30  # Cool down period
    done
}

baseline_test() {
    local url=$1

    echo "📏 Running baseline performance test..."

    # Single user test
    curl -w "@curl-format.txt" -o /dev/null -s "$url"

    # Light load test
    wrk -t1 -c1 -d30s --latency "$url"
}

# Create curl format file for detailed timing
cat > curl-format.txt << 'CURL_FORMAT'
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
CURL_FORMAT

echo "Performance testing utilities:"
echo "  run_load_test <url> [duration] [connections] [threads]"
echo "  stress_test <url> [max_connections]"
echo "  baseline_test <url>"
EOF

it2 session send-text "source /tmp/perf-testing.sh"

echo "✅ Performance testing automation ready!"
```

## Usage Tips

### Running Automation Scripts

1. **Make scripts executable**: `chmod +x script-name.sh`
2. **Customize variables**: Edit server names, directories, etc.
3. **Test incrementally**: Run parts of scripts to verify functionality
4. **Use environment variables**: For sensitive data like API keys

### Integration with CI/CD

Many of these recipes can be integrated into CI/CD pipelines:

```bash
# In your CI/CD pipeline
./setup-fullstack-env.sh
./run-performance-tests.sh
./deploy-and-monitor.sh
```

### Customization

Each recipe can be customized for your specific needs:
- Modify server lists and connection details
- Adjust monitoring intervals and thresholds
- Add custom commands and tools
- Integrate with your specific toolchain

### Security Considerations

- Store sensitive credentials in environment variables
- Use SSH keys instead of passwords
- Implement proper access controls
- Audit automation scripts regularly

This cookbook provides a foundation for advanced it2 automation. Combine and modify these recipes to create powerful automation workflows for your specific environment.
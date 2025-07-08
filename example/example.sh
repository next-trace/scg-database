#!/usr/bin/env bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo -e "${BLUE}=== SCG-Database End-to-End Example ===${NC}"
echo "This script demonstrates the complete workflow of the SCG-Database toolkit."
echo

# --- Configuration ---
CMD_PATH="../cmd/scg-db"
EXAMPLE_PATH="."
MODEL_DIR="./domain"
MODEL_PATH="${MODEL_DIR}/user/user.go"
MIGRATION_DIR="./database/migrations"
CONFIG_FILE="./config.yaml"
DB_FILE="scg_example.db"

# --- Validation ---
log_info "Validating environment..."

# Check if we're in the right directory
if [[ ! -f "../go.mod" ]]; then
    log_error "Please run this script from the example directory"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    log_error "Go is not installed or not in PATH"
    exit 1
fi

log_success "Environment validation passed"

# --- Step 0: Ensure a clean state ---
log_info "Cleaning up previous example runs..."
rm -f ${DB_FILE}
rm -rf ${MODEL_DIR}
rm -rf ${MIGRATION_DIR}
rm -f ${CONFIG_FILE}

log_success "Cleanup completed"

# --- Step 1: Generate Model and Migration via CLI ---
log_info "Step 1: Generating model and migration files via CLI..."

# Create the config file for the CLI
log_info "Creating configuration file..."
cat > ${CONFIG_FILE} <<EOL
database:
  default: gorm:sqlite
  connections:
    gorm:sqlite:
      dsn: ${DB_FILE}
  paths:
    models: "domain"
    migrations: "database/migrations"
EOL

if [[ -f ${CONFIG_FILE} ]]; then
    log_success "Configuration file created successfully"
else
    log_error "Failed to create configuration file"
    exit 1
fi

# Generate the User model
log_info "Generating User model..."
if [[ -f ${MODEL_PATH} ]]; then
    log_warning "User model already exists at: ${MODEL_PATH}"
    log_info "Using existing model file..."
else
    if go run ${CMD_PATH} make model User; then
        log_success "User model generated successfully"
    else
        log_error "Failed to generate User model"
        exit 1
    fi
fi

# Verify model file exists
if [[ -f ${MODEL_PATH} ]]; then
    log_success "Model file available at: ${MODEL_PATH}"
else
    log_error "Model file was not found"
    exit 1
fi

# Generate the migration files
log_info "Generating migration files..."
if go run ${CMD_PATH} migrate make create_users_table; then
    log_success "Migration files generated successfully"
else
    log_error "Failed to generate migration files"
    exit 1
fi

# Verify migration directory was created
if [[ -d ${MIGRATION_DIR} ]]; then
    log_success "Migration directory created at: ${MIGRATION_DIR}"
    MIGRATION_COUNT=$(ls ${MIGRATION_DIR}/*.sql 2>/dev/null | wc -l)
    log_info "Found ${MIGRATION_COUNT} migration files"
else
    log_error "Migration directory was not created"
    exit 1
fi

# --- Step 2: Populate the generated files ---
log_info "Step 2: Populating generated files with schema and fields..."

# Add Name and Email fields to the User model (if needed)
log_info "Checking User model fields..."
if grep -q "Name.*string" ${MODEL_PATH} && grep -q "Email.*string" ${MODEL_PATH}; then
    log_success "User model already has required fields (Name, Email)"
else
    log_info "Adding fields to User model..."
    if sed -i.bak '/\/\/ Define your fields here/a \
        Name  string\
        Email string `gorm:"uniqueIndex"`' ${MODEL_PATH}; then
        rm -f ${MODEL_PATH}.bak
        log_success "User model fields added successfully"
    else
        log_error "Failed to add fields to User model"
        exit 1
    fi
fi

# Fix the TableName method to return "users"
log_info "Fixing TableName method..."
if sed -i.bak 's/return ""/return "users"/' ${MODEL_PATH}; then
    rm -f ${MODEL_PATH}.bak
    log_success "TableName method fixed successfully"
else
    log_error "Failed to fix TableName method"
    exit 1
fi

# Replace BaseModel with a GORM-compatible struct
log_info "Making User model GORM-compatible..."
if sed -i.bak 's/contract\.BaseModel/ID uint `gorm:"primaryKey" json:"id"`/' ${MODEL_PATH}; then
    rm -f ${MODEL_PATH}.bak
    log_success "User model made GORM-compatible"
else
    log_error "Failed to make User model GORM-compatible"
    exit 1
fi

# Add required Model interface methods
log_info "Adding Model interface methods..."
cat >> ${MODEL_PATH} <<'EOL'

// Model interface implementation
func (m *User) PrimaryKey() string { return "id" }
func (m *User) GetID() any { return m.ID }
func (m *User) SetID(id any) { 
    if idVal, ok := id.(uint); ok {
        m.ID = idVal
    }
}
EOL

if [[ $? -eq 0 ]]; then
    log_success "Model interface methods added successfully"
else
    log_error "Failed to add Model interface methods"
    exit 1
fi

# Find the .up.sql file and add the table schema
log_info "Adding table schema to migration file..."
UP_MIGRATION_FILE=$(ls -t ${MIGRATION_DIR}/*.up.sql 2>/dev/null | head -n 1)

if [[ -z ${UP_MIGRATION_FILE} ]]; then
    log_error "No .up.sql migration file found"
    exit 1
fi

log_info "Using migration file: ${UP_MIGRATION_FILE}"

cat > ${UP_MIGRATION_FILE} <<EOL
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    email TEXT UNIQUE,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);
EOL

if [[ $? -eq 0 ]]; then
    log_success "Migration schema added successfully"
else
    log_error "Failed to add migration schema"
    exit 1
fi

# --- Step 3: Run the example Go program ---
log_info "Step 3: Running the example Go program..."
log_info "This will demonstrate database connection, migration, and repository operations..."

if go run ${EXAMPLE_PATH}/main.go; then
    log_success "Example program executed successfully!"
else
    log_error "Example program failed to execute"
    exit 1
fi

echo
log_success "=== Example Completed Successfully! ==="
echo
log_info "What was demonstrated:"
echo "  ✓ CLI model generation"
echo "  ✓ CLI migration generation"
echo "  ✓ File population and customization"
echo "  ✓ Database connection and migration"
echo "  ✓ Repository pattern usage"
echo "  ✓ CRUD operations"
echo "  ✓ Error handling"
echo
log_info "Generated files:"
echo "  - ${MODEL_PATH}"
echo "  - ${UP_MIGRATION_FILE}"
echo "  - ${CONFIG_FILE}"
echo "  - ${DB_FILE}"
echo
log_info "You can now examine these files to understand the SCG-Database workflow!"
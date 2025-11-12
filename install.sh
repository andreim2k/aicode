#!/bin/bash

# aicode Installation Script for macOS and Linux
# This script installs aicode and sets up the required configuration directories
# Supports: macOS (via Homebrew), Ubuntu/Debian (via apt), Fedora/RHEL (via dnf/yum)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/aicode"
SCRIPT_NAME="aicode"

# Functions
print_header() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  aicode Installation Script (macOS & Linux)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_step() {
    echo -e "\n${YELLOW}➜${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

check_requirements() {
    print_step "Checking requirements..."
    
    local missing_deps=()
    
    # Check for jq
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    else
        print_success "jq is installed"
    fi
    
    # Check for claude CLI
    if ! command -v claude &> /dev/null; then
        missing_deps+=("claude")
    else
        print_success "claude CLI is installed"
    fi
    
    # Check for Go compiler (needed for z-ai-proxy)
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    else
        print_success "Go compiler is installed"
    fi
    
    # Check bash version
    if [ "${BASH_VERSINFO[0]}" -lt 4 ]; then
        missing_deps+=("bash 4.0+")
    else
        print_success "bash ${BASH_VERSION%%.*} is installed"
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        return 1
    fi
    
    print_success "All requirements met"
    return 0
}

create_directories() {
    print_step "Creating directories..."
    
    if [ ! -d "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR"
        print_success "Created $INSTALL_DIR"
    else
        print_info "$INSTALL_DIR already exists"
    fi
    
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        print_success "Created $CONFIG_DIR"
    else
        print_info "$CONFIG_DIR already exists"
    fi
}

install_script() {
    print_step "Installing aicode script..."
    
    if [ ! -f "$SCRIPT_DIR/$SCRIPT_NAME" ]; then
        print_error "Script file not found: $SCRIPT_DIR/$SCRIPT_NAME"
        return 1
    fi
    
    cp "$SCRIPT_DIR/$SCRIPT_NAME" "$INSTALL_DIR/$SCRIPT_NAME"
    chmod +x "$INSTALL_DIR/$SCRIPT_NAME"
    print_success "Installed $SCRIPT_NAME to $INSTALL_DIR"
}

build_proxy() {
    print_step "Building proxy binaries..."
    
    # Change to script directory for building (needed for go.mod)
    cd "$SCRIPT_DIR" || {
        print_error "Failed to change to script directory"
        return 1
    }
    
    # Build z-ai-proxy
    if [ -f "$SCRIPT_DIR/z-ai-proxy.go" ]; then
        print_info "Building z-ai-proxy..."
        if ! go build -o "$CONFIG_DIR/z-ai-proxy" "$SCRIPT_DIR/z-ai-proxy.go" 2>&1 | tee /tmp/z-ai-proxy-build.log; then
            print_error "Failed to build z-ai-proxy"
            if [ -f /tmp/z-ai-proxy-build.log ]; then
                print_info "Build errors:"
                cat /tmp/z-ai-proxy-build.log
            fi
            print_info "Make sure Go is installed and dependencies are available (run: go mod tidy)"
            return 1
        fi
        chmod +x "$CONFIG_DIR/z-ai-proxy"
        print_success "Built and installed z-ai-proxy to $CONFIG_DIR"
        rm -f /tmp/z-ai-proxy-build.log 2>/dev/null || true
    else
        print_info "z-ai-proxy.go not found, skipping z-ai-proxy build"
    fi
    
    # Build x-ai-proxy
    if [ -f "$SCRIPT_DIR/x-ai-proxy.go" ]; then
        print_info "Building x-ai-proxy..."
        if ! go build -o "$CONFIG_DIR/x-ai-proxy" "$SCRIPT_DIR/x-ai-proxy.go" 2>&1 | tee /tmp/x-ai-proxy-build.log; then
            print_error "Failed to build x-ai-proxy"
            if [ -f /tmp/x-ai-proxy-build.log ]; then
                print_info "Build errors:"
                cat /tmp/x-ai-proxy-build.log
            fi
            print_info "Make sure Go is installed and dependencies are available (run: go mod tidy)"
            return 1
        fi
        chmod +x "$CONFIG_DIR/x-ai-proxy"
        print_success "Built and installed x-ai-proxy to $CONFIG_DIR"
        rm -f /tmp/x-ai-proxy-build.log 2>/dev/null || true
    else
        print_info "x-ai-proxy.go not found, skipping x-ai-proxy build"
    fi
}

check_path() {
    print_step "Checking PATH configuration..."
    
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        print_success "$INSTALL_DIR is in PATH"
        return 0
    fi
    
    print_info "$INSTALL_DIR is not in your PATH"
    
    # Detect shell and config file
    local shell_config=""
    if [ -n "${ZSH_VERSION:-}" ]; then
        shell_config="$HOME/.zshrc"
    elif [ -n "${BASH_VERSION:-}" ]; then
        shell_config="$HOME/.bashrc"
    fi
    
    if [ -z "$shell_config" ]; then
        print_error "Could not detect shell configuration file"
        print_info "Please manually add this to your shell config:"
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        return 1
    fi
    
    print_info "Adding $INSTALL_DIR to $shell_config..."
    
    if ! grep -q "\.local/bin" "$shell_config"; then
        echo "" >> "$shell_config"
        echo "# Added by aicode installer" >> "$shell_config"
        echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> "$shell_config"
        print_success "Updated $shell_config"
        print_info "Please run: source $shell_config"
    else
        print_success "$shell_config already configured"
    fi
}

setup_configuration() {
    print_step "Setting up configuration..."
    
    if [ -f "$CONFIG_DIR/settings.json" ]; then
        print_info "settings.json already exists, skipping..."
        return 0
    fi
    
    if [ -f "$SCRIPT_DIR/settings.json.example" ]; then
        cp "$SCRIPT_DIR/settings.json.example" "$CONFIG_DIR/settings.json"
        print_success "Created example settings.json"
        print_info "Edit $CONFIG_DIR/settings.json with your provider configuration"
    else
        print_error "settings.json.example not found"
        print_info "You'll need to create $CONFIG_DIR/settings.json manually"
    fi
}

print_completion() {
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Installation Complete!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Reload your shell configuration:"
    echo "     source ~/.zshrc  # or ~/.bashrc"
    echo ""
    echo "  2. Edit your configuration:"
    echo "     nano $CONFIG_DIR/settings.json"
    echo ""
    echo "  3. Test the installation:"
    echo "     aicode"
    echo ""
    echo "For more information, see:"
    echo "  • README.md"
    echo "  • docs/QUICKSTART.md"
    echo "  • docs/CONFIGURATION.md"
    echo ""
}

detect_platform() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macOS"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /etc/os-release ]; then
            . /etc/os-release
            echo "$ID"
        else
            echo "linux"
        fi
    else
        echo "unknown"
    fi
}

install_jq_platform() {
    local platform="$1"
    
    case "$platform" in
        macOS)
            echo "  brew install jq"
            ;;
        ubuntu|debian)
            echo "  sudo apt-get update && sudo apt-get install -y jq"
            ;;
        fedora|rhel|centos)
            echo "  sudo dnf install -y jq  # or: sudo yum install -y jq"
            ;;
        *)
            echo "  Visit: https://stedolan.github.io/jq/download/"
            ;;
    esac
}

install_go_platform() {
    local platform="$1"
    
    case "$platform" in
        macOS)
            echo "  brew install go"
            ;;
        ubuntu|debian)
            echo "  sudo apt-get update && sudo apt-get install -y golang-go"
            echo "  # Or install latest from: https://go.dev/dl/"
            ;;
        fedora|rhel|centos)
            echo "  sudo dnf install -y golang  # or: sudo yum install -y golang"
            echo "  # Or install latest from: https://go.dev/dl/"
            ;;
        *)
            echo "  Visit: https://go.dev/dl/"
            ;;
    esac
}

main() {
    print_header
    echo ""
    
    local platform=$(detect_platform)
    
    if [ "$platform" == "unknown" ]; then
        print_error "Unsupported operating system: $OSTYPE"
        print_info "This installer supports macOS and Linux (Ubuntu, Debian, Fedora, RHEL)"
        exit 1
    fi
    
    print_info "Detected platform: $platform"
    echo ""
    
    # Run installation steps
    if ! check_requirements; then
        print_error "Please install missing dependencies:"
        echo ""
        if ! command -v jq &> /dev/null; then
            echo "  To install jq:"
            install_jq_platform "$platform"
            echo ""
        fi
        if ! command -v claude &> /dev/null; then
            echo "  To install Claude CLI:"
            echo "  Visit: https://github.com/anthropics/claude-cli"
            echo ""
        fi
        if ! command -v go &> /dev/null; then
            echo "  To install Go:"
            install_go_platform "$platform"
            echo ""
        fi
        exit 1
    fi
    
    create_directories
    install_script
    build_proxy
    check_path
    setup_configuration
    print_completion
}

# Run main function
main "$@"

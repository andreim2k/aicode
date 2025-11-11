# Installation Guide

Complete installation instructions for `aicode` on macOS.

## System Requirements

- **OS**: macOS (10.12 or later)
- **Shell**: bash 4.0+ or zsh
- **Dependencies**:
  - `jq` (JSON query tool)
  - `claude` CLI (Anthropic)

## Quick Install (Recommended for macOS)

### Step 1: Clone or Download

```bash
# Clone from GitHub
git clone https://github.com/yourusername/aicode.git
cd aicode

# Or download the ZIP file and extract it
```

### Step 2: Run the Installer

```bash
chmod +x install.sh
./install.sh
```

The installer will:
- ✓ Check all system requirements
- ✓ Install aicode to `~/.local/bin`
- ✓ Create config directories
- ✓ Update your PATH
- ✓ Set up example configuration

### Step 3: Reload Your Shell

```bash
source ~/.zshrc
# or if using bash
source ~/.bashrc
```

### Step 4: Configure Providers

```bash
nano ~/.config/aicode/settings.json
```

Add your provider configurations (see [CONFIGURATION.md](./CONFIGURATION.md)).

### Step 5: Test

```bash
aicode
```

## Detailed Manual Installation

If you prefer manual installation or the automatic installer doesn't work:

### 1. Install Dependencies

#### Install jq

```bash
brew install jq
```

Verify:
```bash
jq --version
```

#### Install Claude CLI

Visit: https://github.com/anthropics/claude-cli

Or if available through package managers:
```bash
# Currently, Claude CLI must be installed from the official repository
# Follow the instructions at: https://github.com/anthropics/claude-cli
```

Verify:
```bash
claude --version
```

### 2. Copy the Script

```bash
mkdir -p ~/.local/bin
cp aicode ~/.local/bin/aicode
chmod +x ~/.local/bin/aicode
```

### 3. Update PATH

Add `~/.local/bin` to your PATH if not already present.

#### For zsh (default in macOS Catalina+)

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

#### For bash

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Verify PATH is set:
```bash
echo $PATH | grep ".local/bin"
```

### 4. Create Configuration Directory

```bash
mkdir -p ~/.config/aicode
```

### 5. Set Up Configuration

```bash
cp settings.json.example ~/.config/aicode/settings.json
nano ~/.config/aicode/settings.json
```

Edit the file and add your provider settings.

### 6. Test Installation

```bash
aicode
```

You should see your providers listed in an interactive menu.

## Troubleshooting

### "command not found: aicode"

**Cause**: `~/.local/bin` is not in your PATH

**Solution**:
1. Check if `~/.local/bin` exists: `ls -la ~/.local/bin`
2. Add to PATH: `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc`
3. Reload: `source ~/.zshrc`
4. Verify: `which aicode`

### "jq: command not found"

**Cause**: jq is not installed

**Solution**:
```bash
brew install jq
# Verify
jq --version
```

### "claude: command not found"

**Cause**: Claude CLI is not installed

**Solution**:
1. Visit: https://github.com/anthropics/claude-cli
2. Follow installation instructions for macOS
3. Verify: `claude --version`

### "Settings file not found"

**Cause**: Configuration file missing or in wrong location

**Solution**:
```bash
# Check if file exists
ls -la ~/.config/aicode/settings.json

# Create if missing
mkdir -p ~/.config/aicode
cp settings.json.example ~/.config/aicode/settings.json

# Edit with your settings
nano ~/.config/aicode/settings.json
```

### "Invalid JSON in settings.json"

**Cause**: JSON formatting errors

**Solution**:
1. Validate JSON: `jq . ~/.config/aicode/settings.json`
2. Check formatting at: https://jsonlint.com/
3. Common issues:
   - Missing commas between array elements
   - Unquoted strings
   - Trailing commas

Example of invalid JSON:
```json
{
  "provider": "test",
  "models": ["model1", "model2",]  // ← Trailing comma is invalid
}
```

### Script permission denied

**Cause**: Script doesn't have execute permission

**Solution**:
```bash
chmod +x ~/.local/bin/aicode
```

## Uninstalling

If you need to remove aicode:

```bash
# Remove the script
rm ~/.local/bin/aicode

# Optional: Remove configuration
rm -rf ~/.config/aicode

# Optional: Remove PATH entry from shell config
# Edit ~/.zshrc or ~/.bashrc and remove the PATH line
```

## Next Steps

- Read [QUICKSTART.md](./QUICKSTART.md) for basic usage
- Read [CONFIGURATION.md](./CONFIGURATION.md) for setup instructions
- Check [../README.md](../README.md) for full documentation

## Support

If you encounter issues:

1. Check this troubleshooting guide
2. Verify all dependencies are installed
3. Validate your JSON configuration
4. Check the official repository for issues

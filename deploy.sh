#!/bin/bash

# Reddit Poker Bot Deployment Script
# Deploys the bot to a Linux server via SSH

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="reddit-poker-bot"
REMOTE_DIR="/opt/reddit-poker-bot"
SERVICE_NAME="reddit-poker-bot.service"
SYSTEMD_DIR="/etc/systemd/system"

# Check for required arguments
if [ $# -lt 1 ]; then
    echo -e "${RED}Usage: $0 <user@server> [ssh-port]${NC}"
    echo "Example: $0 user@example.com"
    echo "Example: $0 user@example.com 2222"
    exit 1
fi

SERVER=$1
SSH_PORT=${2:-22}  # Default to port 22 if not specified

echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Reddit Poker Bot Deployment Script       ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════╝${NC}"
echo ""
echo "Target server: $SERVER"
echo "SSH port: $SSH_PORT"
echo ""

# Step 1: Build the binary for Linux
echo -e "${YELLOW}[1/6]${NC} Building binary for Linux..."
GOOS=linux GOARCH=amd64 go build -o ${BINARY_NAME}-linux .
echo -e "${GREEN}✓${NC} Binary built successfully"

# Step 2: Create temporary deployment directory
echo -e "${YELLOW}[2/6]${NC} Preparing deployment files..."
TEMP_DIR=$(mktemp -d)
cp ${BINARY_NAME}-linux $TEMP_DIR/$BINARY_NAME
cp $SERVICE_NAME $TEMP_DIR/

# Use .env if it exists, otherwise use .env.example
if [ -f .env ]; then
    echo "Using existing .env file"
    cp .env $TEMP_DIR/.env
else
    echo "Using .env.example as template"
    cp .env.example $TEMP_DIR/.env
fi

echo -e "${GREEN}✓${NC} Files prepared in $TEMP_DIR"

# Step 3: Create deployment script that will run on the server
cat > $TEMP_DIR/install.sh << 'EOF'
#!/bin/bash
set -e

BINARY_NAME="reddit-poker-bot"
REMOTE_DIR="/opt/reddit-poker-bot"
SERVICE_NAME="reddit-poker-bot.service"
SYSTEMD_DIR="/etc/systemd/system"
BOT_USER="reddit-bot"

echo "Installing Reddit Poker Bot..."

# Create user if it doesn't exist
if ! id "$BOT_USER" &>/dev/null; then
    echo "Creating $BOT_USER user..."
    sudo useradd -r -s /bin/false $BOT_USER
fi

# Create directory and set permissions
echo "Setting up directory..."
sudo mkdir -p $REMOTE_DIR
sudo cp $BINARY_NAME $REMOTE_DIR/
sudo cp .env $REMOTE_DIR/
sudo chown -R $BOT_USER:$BOT_USER $REMOTE_DIR
sudo chmod +x $REMOTE_DIR/$BINARY_NAME

# Install systemd service
echo "Installing systemd service..."
sudo cp $SERVICE_NAME $SYSTEMD_DIR/
sudo systemctl daemon-reload

# Stop service if running
if sudo systemctl is-active --quiet $SERVICE_NAME; then
    echo "Stopping existing service..."
    sudo systemctl stop $SERVICE_NAME
fi

# Enable and start service
echo "Starting service..."
sudo systemctl enable $SERVICE_NAME
sudo systemctl start $SERVICE_NAME

echo "Installation complete!"
echo ""
echo "Service status:"
sudo systemctl status $SERVICE_NAME --no-pager
echo ""
echo "To view logs: sudo journalctl -u $SERVICE_NAME -f"
EOF

chmod +x $TEMP_DIR/install.sh

# Step 4: Copy files to server
echo -e "${YELLOW}[3/6]${NC} Copying files to server..."
ssh -p $SSH_PORT $SERVER "mkdir -p ~/reddit-poker-bot-deploy"
scp -P $SSH_PORT $TEMP_DIR/* $SERVER:~/reddit-poker-bot-deploy/
echo -e "${GREEN}✓${NC} Files copied to server"

# Step 5: Run installation script on server
echo -e "${YELLOW}[4/6]${NC} Installing on server..."
ssh -p $SSH_PORT -t $SERVER "cd ~/reddit-poker-bot-deploy && bash install.sh"
echo -e "${GREEN}✓${NC} Installation complete"

# Step 6: Cleanup
echo -e "${YELLOW}[5/6]${NC} Cleaning up..."
ssh -p $SSH_PORT $SERVER "rm -rf ~/reddit-poker-bot-deploy"
rm -rf $TEMP_DIR
rm ${BINARY_NAME}-linux
echo -e "${GREEN}✓${NC} Cleanup complete"

# Step 7: Display service status
echo -e "${YELLOW}[6/6]${NC} Verifying deployment..."
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Deployment Complete!                      ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════╝${NC}"
echo ""
echo "The bot is now running as a systemd service."
echo ""
echo "Useful commands (run on server):"
echo "  - View status:  sudo systemctl status $SERVICE_NAME"
echo "  - View logs:    sudo journalctl -u $SERVICE_NAME -f"
echo "  - Stop:         sudo systemctl stop $SERVICE_NAME"
echo "  - Start:        sudo systemctl start $SERVICE_NAME"
echo "  - Restart:      sudo systemctl restart $SERVICE_NAME"
echo "  - Edit config:  sudo nano $REMOTE_DIR/.env (then restart)"
echo ""

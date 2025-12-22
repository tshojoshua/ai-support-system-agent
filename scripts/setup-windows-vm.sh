#!/bin/bash
# Quick script to spin up Windows VM in Azure for MSI build

set -e

echo "🔧 Setting up Windows Build VM in Azure..."

# Variables
RG_NAME="jtnt-build-rg"
VM_NAME="windows-build-vm"
LOCATION="eastus"
VM_SIZE="Standard_D2s_v3"
ADMIN_USER="jtntadmin"
ADMIN_PASS="SecureP@ssw0rd123!"  # Change this!

# Create resource group
echo "Creating resource group..."
az group create --name $RG_NAME --location $LOCATION

# Create Windows VM
echo "Creating Windows VM (this takes 3-5 minutes)..."
az vm create \
  --resource-group $RG_NAME \
  --name $VM_NAME \
  --image Win2022Datacenter \
  --size $VM_SIZE \
  --admin-username $ADMIN_USER \
  --admin-password "$ADMIN_PASS" \
  --public-ip-sku Standard

# Get public IP
PUBLIC_IP=$(az vm show -d -g $RG_NAME -n $VM_NAME --query publicIps -o tsv)

echo ""
echo "✅ Windows VM Ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Public IP: $PUBLIC_IP"
echo "Username: $ADMIN_USER"
echo "Password: $ADMIN_PASS"
echo ""
echo "Connect via RDP:"
echo "  xfreerdp /v:$PUBLIC_IP /u:$ADMIN_USER /p:'$ADMIN_PASS' /size:1920x1080"
echo ""
echo "Or use Remmina (GUI RDP client)"
echo ""
echo "After you're done building, delete VM:"
echo "  az group delete --name $RG_NAME --yes --no-wait"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

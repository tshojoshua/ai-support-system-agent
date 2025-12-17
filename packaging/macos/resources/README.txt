JTNT Agent - macOS Package

The JTNT Agent provides secure remote management and monitoring capabilities for macOS systems through the JTNT Hub platform.

Features:
• Secure mTLS communication with JTNT Hub
• Job execution with capability-based policies
• Real-time health and performance monitoring
• Automatic certificate rotation
• Signed self-update capability
• Comprehensive audit logging

Installation:
The agent will be installed to /usr/local/jtnt/agent/ and configured to start automatically via launchd.

Post-Installation:
After installation, you must enroll the agent with your JTNT Hub using an enrollment token:

  sudo jtnt-agent enroll --token YOUR_TOKEN

System Requirements:
• macOS 13.0 (Ventura) or later
• 50 MB free disk space
• Network connectivity to JTNT Hub

For more information:
https://docs.jtnt.us/agent

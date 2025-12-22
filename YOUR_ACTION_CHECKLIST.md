# EMERGENCY DEPLOYMENT - YOUR ACTION CHECKLIST
# Current Time: ~10:30 PM PST
# Target: Client deployment at 1:30 AM
# Decision: Option B - Deploy Tonight

## IMMEDIATE ACTIONS (Next 10 Minutes) - YOU DO THESE

### 1. Contact Agent Build Team ⏰ NOW
**Action:** Send them EMERGENCY_DEPLOYMENT.md
```bash
# Send via Slack/Email/Text - whatever gets fastest response
Subject: 🚨 EMERGENCY: Windows MSI Build - START NOW
Message: "Critical client deployment tonight. Need Windows MSI in 2.5 hours.
Full instructions attached: EMERGENCY_DEPLOYMENT.md
Reply immediately if you can execute."

Attachment: /home/tsho/ai-support-system/agent/EMERGENCY_DEPLOYMENT.md
```

**Who to contact:**
- [ ] Primary developer: _______________
- [ ] Backup developer: _______________
- [ ] DevOps lead: _______________

**Confirm receipt within 5 minutes or escalate**

---

### 2. Create Enrollment Tokens ⏰ NOW
```bash
# SSH to Hub server
ssh tsho@hub.jtnt.us

# OR via web UI: https://hub.jtnt.us/admin/tokens

# Create 2 tokens:

# Token 1: TEST (for tonight's testing)
# Name: Emergency-Test-Dec21
# Uses: 10
# Expires: Tomorrow (Dec 22)

# Token 2: PRODUCTION (for client machines) 
# Name: Client-Production-Dec21
# Uses: 50
# Expires: Jan 31, 2026

# SAVE BOTH TOKENS IN SECURE NOTE:
TEST_TOKEN=etok_________________________
PROD_TOKEN=etok_________________________
```

**Confirm tokens created:**
- [ ] TEST token: `etok_________________`
- [ ] PROD token: `etok_________________`

---

### 3. Alert Client ⏰ NOW
**Email/Call Client Contact**

```
Subject: JTNT Agent Deployment - Tonight 1:30 AM PST

Hi [Client Name],

We're proceeding with the JTNT Agent deployment tonight as planned.

Timeline:
• 1:30 AM: Begin installation on first test system
• 1:45 AM: Verify and proceed to remaining systems  
• 2:30 AM: All systems online and monitored
• 3:00 AM: Final verification complete

Please ensure:
✓ Someone available by phone/email if we need access
✓ All target systems powered on
✓ No critical work in progress
✓ Administrator passwords ready if needed

We'll send status updates every 30 minutes.

Contact: [Your phone]
Emergency: [Backup contact]

Thanks,
[Your Name]
```

**Client confirmation:**
- [ ] Client notified
- [ ] Client available at 1:30 AM: YES / NO
- [ ] Systems confirmed ready: YES / NO
- [ ] Contact info verified

---

### 4. Verify Hub Health ⏰ NOW
```bash
cd /home/tsho/ai-support-system/agent

# Quick health check
curl -s https://hub.jtnt.us/api/v1/health

# Verify endpoints (should return 400/401, NOT 404)
curl -X POST https://hub.jtnt.us/api/v1/agent/enroll \
  -H "Content-Type: application/json" \
  -d '{"token":"test"}' 2>&1 | head -3

curl https://hub.jtnt.us/api/v1/agent/jobs 2>&1 | head -3

curl https://hub.jtnt.us/api/v1/agent/diagnostics/next 2>&1 | head -3

# All should show error messages but NOT "404 Not Found"
```

**Hub health check:**
- [ ] /api/v1/health returns 200 OK
- [ ] /api/v1/agent/enroll returns 400/401 (NOT 404)
- [ ] /api/v1/agent/jobs returns 401 (NOT 404)
- [ ] /api/v1/agent/diagnostics/next returns 401 (NOT 404)

**If any return 404: ABORT IMMEDIATELY - Hub not ready**

---

## TIMELINE & CHECKPOINTS

### 11:00 PM - Status Check #1
**Expected:** "Build tools installed, binaries compiling"

**Your verification:**
```bash
# Check for message from agent team
# Verify Windows machine acquired
# Verify build started
```

- [ ] Message received: _________________
- [ ] On track: YES / NO / ISSUES
- [ ] Action needed: _________________

---

### 11:30 PM - CHECKPOINT 1: Binaries
**Expected:** "Binaries complete, packaging MSI"

**DECISION POINT:**
- Agent binaries compiled successfully?
- File sizes correct (10-15 MB each)?

**IF NO: Abort to tomorrow**

- [ ] jtnt-agent.exe built: YES / NO
- [ ] jtnt-agentd.exe built: YES / NO
- [ ] GO / NO-GO: __________

---

### 12:00 AM - Status Check #2
**Expected:** "MSI build in progress"

**Your verification:**
```bash
# Confirm MSI packaging underway
# No critical errors reported
```

- [ ] Message received: _________________
- [ ] On track: YES / NO / ISSUES
- [ ] Action needed: _________________

---

### 12:15 AM - CHECKPOINT 2: MSI File
**Expected:** "MSI complete, starting tests"

**DECISION POINT:**
- MSI file created successfully?
- File size correct (15-25 MB)?

**IF NO: Abort to tomorrow**

- [ ] MSI file exists: YES / NO
- [ ] File size OK: YES / NO
- [ ] GO / NO-GO: __________

---

### 12:30 AM - Status Check #3
**Expected:** "Testing in progress"

**Your verification:**
```bash
# Prepare to check Hub for test agent
# Monitor Hub logs for enrollment/heartbeat
```

- [ ] Test installation started: YES / NO
- [ ] TEST token sent to agent team: YES / NO

---

### 1:00 AM - CHECKPOINT 3: FINAL GO/NO-GO
**Expected:** "Testing complete - GO/NO-GO: [DECISION]"

**CRITICAL DECISION:**

**Agent team reports:**
- [ ] MSI installs: PASS / FAIL
- [ ] Service starts: PASS / FAIL
- [ ] Agent enrolls: PASS / FAIL
- [ ] Heartbeat active: PASS / FAIL
- [ ] No errors: PASS / FAIL

**Your verification:**
```bash
# Check Hub dashboard for test agent
# Verify agent visible and online
# Check Hub logs for heartbeat

ssh tsho@hub.jtnt.us
docker compose logs hub-api | grep -i heartbeat | tail -20
docker compose logs hub-api | grep -i enroll | tail -10

# Check database
docker compose exec hub-db psql -U postgres -d hub -c \
  "SELECT hostname, status, last_heartbeat FROM agents ORDER BY created_at DESC LIMIT 1;"
```

**Hub verification:**
- [ ] Agent visible in dashboard: YES / NO
- [ ] Agent status: ONLINE / OFFLINE
- [ ] Heartbeat recent (< 2 min): YES / NO
- [ ] No errors in Hub logs: YES / NO

**FINAL DECISION:**
- [ ] **🟢 GO - Proceed to client deployment**
- [ ] **🔴 NO-GO - Abort to tomorrow**

**If NO-GO:**
```bash
# Send abort message
# Notify client of reschedule
# Document failure reason
# Schedule tomorrow deployment
```

---

## CLIENT DEPLOYMENT (If GO at 1:00 AM)

### 1:30 AM - Deploy to Test Machine

**Client Info:**
- Test machine hostname: _________________
- Test machine IP: _________________
- Admin credentials: _________________

**Installation:**
```powershell
# Remote to client test machine or have them run:
msiexec /i JTNT-Agent-4.0.0.msi /qb `
  ENROLLMENT_TOKEN="[PROD_TOKEN]" `
  HUB_URL="https://hub.jtnt.us"

# Wait 2 minutes, verify
```

**Verification (1:35 AM):**
```bash
# Check Hub for new agent
# Verify status online
# Check heartbeat timestamp
```

- [ ] Service running on test machine: YES / NO
- [ ] Agent in Hub dashboard: YES / NO
- [ ] Heartbeat active: YES / NO
- [ ] **PROCEED to remaining machines: YES / NO**

**If test fails: STOP, rollback, reschedule**

---

### 1:45 AM - Deploy to Remaining Machines

**Client machines list:**
1. [ ] _____________ (IP: _________)
2. [ ] _____________ (IP: _________)
3. [ ] _____________ (IP: _________)
4. [ ] _____________ (IP: _________)
5. [ ] _____________ (IP: _________)

**Install on each:**
```powershell
msiexec /i JTNT-Agent-4.0.0.msi /qn `
  ENROLLMENT_TOKEN="[PROD_TOKEN]" `
  HUB_URL="https://hub.jtnt.us"
```

**Track progress:**
- [ ] Machine 1: INSTALLED / FAILED
- [ ] Machine 2: INSTALLED / FAILED
- [ ] Machine 3: INSTALLED / FAILED
- [ ] Machine 4: INSTALLED / FAILED
- [ ] Machine 5: INSTALLED / FAILED

---

### 2:00 AM - Linux Server (if applicable)

```bash
# If client has Linux server:
sudo dpkg -i jtnt-agent_1.0.0_amd64.deb

sudo jtnt-agent enroll \
  --token [PROD_TOKEN] \
  --hub-url https://hub.jtnt.us

systemctl status jtnt-agentd
```

- [ ] Linux agent installed: YES / NO / N/A
- [ ] Linux agent enrolled: YES / NO / N/A

---

### 2:30 AM - Final Verification

**Hub Dashboard Check:**
```bash
# Verify all agents online
# Check heartbeat timestamps
# Verify no errors in logs
```

**Expected agent count:**
- Windows agents: _____ (expected) / _____ (actual)
- Linux agents: _____ (expected) / _____ (actual)
- **Total:** _____ / _____

**All checks:**
- [ ] All agents visible in Hub
- [ ] All showing status: ONLINE
- [ ] All heartbeats < 2 minutes ago
- [ ] No errors in Hub logs
- [ ] No errors in agent logs (sample check)

**If all pass:**
```
🟢 DEPLOYMENT SUCCESSFUL
```

**Send client confirmation:**
```
Subject: JTNT Agent Deployment - Complete

All systems successfully deployed and monitored.

Agents installed: [X] Windows, [Y] Linux
Status: All online and reporting
Next heartbeat: < 2 minutes

We'll continue monitoring for the next hour. 
You'll receive a final report at 3:30 AM.

No action needed from your side.
```

---

### 3:00 AM - Post-Deployment Monitoring

**Monitor for 30-60 minutes:**
- Watch Hub logs for any errors
- Verify heartbeats continue regularly
- Check for any agent disconnections

**If any issues: Investigate immediately**

---

### 3:30 AM - Final Report

**Send to client:**
```
Subject: JTNT Agent Deployment - Final Report

Deployment Summary:
• Start time: 1:30 AM PST
• Completion: 2:30 AM PST
• Duration: 1 hour

Agents Deployed:
• Windows: [X] systems
• Linux: [Y] systems  
• Total: [Z] systems

Status: All agents online and reporting
Uptime: 1 hour with zero issues

Next Steps:
• Agents will report automatically 24/7
• You can view status at: https://hub.jtnt.us
• Support: team@jtnt.us

Thank you for your patience.
```

---

## ROLLBACK PLAN (If Deployment Fails)

**On each machine:**
```powershell
# Uninstall
msiexec /x JTNT-Agent-4.0.0.msi /qn

# Or manual
# Control Panel > Programs > Uninstall > JTNT Agent
```

**Cleanup:**
```powershell
# Remove files
Remove-Item "C:\Program Files\JTNT\Agent" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item "C:\ProgramData\JTNT\Agent" -Recurse -Force -ErrorAction SilentlyContinue

# Remove service (if stuck)
sc.exe delete JTNTAgent
```

**Client communication:**
```
Subject: JTNT Agent Deployment - Rescheduled

We encountered [issue description] during deployment.

To ensure quality, we've:
✓ Rolled back all changes
✓ Identified root cause
✓ Preparing fix for tomorrow

Rescheduled deployment:
• Date: Tomorrow, Dec 22
• Time: [time]
• Estimated duration: [time]

Your systems are unchanged and fully operational.
We apologize for the delay and will ensure 
a smooth deployment tomorrow.
```

---

## EMERGENCY CONTACTS

**Agent Build Team:**
- Primary: ________________ (phone: _______)
- Backup: ________________ (phone: _______)

**Hub/Infrastructure:**
- Primary: ________________ (phone: _______)

**Client:**
- Primary: ________________ (phone: _______)
- Emergency: ______________ (phone: _______)

**Escalation:**
- Manager: ________________ (phone: _______)

---

## ABORT DECISION CRITERIA

**ABORT if ANY of these occur:**

- [ ] Agent team cannot get Windows machine in 30 min
- [ ] Build tools won't install (15 min trying)
- [ ] Binaries fail to compile (Checkpoint 1 fail)
- [ ] MSI build fails (Checkpoint 2 fail)
- [ ] Test installation fails any check (Checkpoint 3 fail)
- [ ] Hub cannot see test agent
- [ ] Test agent shows errors
- [ ] Team confidence < 95%
- [ ] Any "gut feeling" something is wrong

**NO SHAME IN ABORTING. Client trust > meeting timeline.**

---

## CONFIDENCE CHECK

Before proceeding to client deployment, rate your confidence:

At 1:00 AM Checkpoint 3:
```
My confidence level: _____ % (need 95%+)

Reasons for confidence:
- _______________________________
- _______________________________
- _______________________________

Concerns (if any):
- _______________________________
- _______________________________

DECISION: GO / NO-GO
```

**If < 95% confident: NO-GO to tomorrow**

---

**Generated:** Dec 21, 2025 10:30 PM PST  
**Timeline:** 3.5 hours to deployment  
**Status:** READY TO EXECUTE

**🚨 START ACTIONS NOW 🚨**

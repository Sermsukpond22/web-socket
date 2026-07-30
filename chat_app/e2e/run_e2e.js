#!/usr/bin/env node

/**
 * Main E2E Test Suite Runner Entrypoint
 * Real-time Chat Application (Instagram DM Style)
 */

const { header, pass, fail, info, warn, colors } = require('./lib/logger');
const runTier1 = require('./tests/tier1_auth.test');
const runTier2 = require('./tests/tier2_friends.test');
const runTier3 = require('./tests/tier3_websocket.test');
const runTier4 = require('./tests/tier4_error_handling.test');

async function main() {
  const startTime = Date.now();
  header('REAL-TIME CHAT APPLICATION - E2E TEST SUITE');

  const baseUrl = process.env.BASE_URL || 'http://localhost:3000';
  const wsUrl = process.env.WS_URL || 'ws://localhost:3000';

  info(`Target Base HTTP URL: ${baseUrl}`);
  info(`Target Base WS URL:   ${wsUrl}`);

  const ctx = {
    baseUrl,
    wsUrl,
    userA: null,
    userB: null,
    userC: null
  };

  const results = {
    tier1: false,
    tier2: false,
    tier3: false,
    tier4: false
  };

  try {
    results.tier1 = await runTier1(ctx);
  } catch (err) {
    fail('Tier 1 Execution Error', err);
  }

  try {
    results.tier2 = await runTier2(ctx);
  } catch (err) {
    fail('Tier 2 Execution Error', err);
  }

  try {
    results.tier3 = await runTier3(ctx);
  } catch (err) {
    fail('Tier 3 Execution Error', err);
  }

  try {
    results.tier4 = await runTier4(ctx);
  } catch (err) {
    fail('Tier 4 Execution Error', err);
  }

  const durationSec = ((Date.now() - startTime) / 1000).toFixed(2);

  console.log("\n" + colors.bright + colors.cyan + "=".repeat(60) + colors.reset);
  console.log(colors.bright + colors.cyan + '  E2E TEST SUITE SUMMARY REPORT' + colors.reset);
  console.log(colors.bright + colors.cyan + "=".repeat(60) + colors.reset);

  const printTierStatus = (name, passed) => {
    const icon = passed ? `${colors.green}✔ PASS${colors.reset}` : `${colors.red}✖ FAIL${colors.reset}`;
    console.log(`  ${icon}  ${name}`);
  };

  printTierStatus('Tier 1: Auth & User Profile (Register, Login, /api/auth/me)', results.tier1);
  printTierStatus('Tier 2: Boundary & Friend System (Request, Pending, Accept, Restriction)', results.tier2);
  printTierStatus('Tier 3: Real-time WebSocket Messaging & DB Persistence', results.tier3);
  printTierStatus('Tier 4: Scenario & Error Handling (Invalid WS token, Duplicate Request)', results.tier4);

  const allPassed = Object.values(results).every(Boolean);

  console.log("\n" + colors.gray + `Total execution time: ${durationSec}s` + colors.reset);

  if (allPassed) {
    console.log(`\n${colors.bright}${colors.green}🎉 ALL E2E TEST TIERS PASSED SUCCESSFULLY!${colors.reset}\n`);
    process.exit(0);
  } else {
    console.log(`\n${colors.bright}${colors.red}❌ SOME E2E TEST TIERS FAILED. PLEASE REVIEW LOGS ABOVE.${colors.reset}\n`);
    process.exit(1);
  }
}

if (require.main === module) {
  main().catch(err => {
    console.error('Fatal E2E test runner failure:', err);
    process.exit(1);
  });
}

module.exports = { main };

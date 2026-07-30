/**
 * Tier 4: Scenario & Error Handling Tests
 */

const ApiClient = require('../lib/api_client');
const WsClient = require('../lib/ws_client');
const { subheader, pass, fail, info } = require('../lib/logger');

async function runTier4(ctx) {
  subheader('Tier 4: Scenario & Error Handling');
  const api = new ApiClient(ctx.baseUrl);
  const testResults = [];

  // Test 4.1: Connect to WS without token (missing token)
  try {
    const wsNoToken = new WsClient(ctx.wsUrl);
    let rejected = false;

    try {
      await wsNoToken.connect(null); // No token provided
      info('Connected without token; waiting for immediate close/error event');
      // If open event fired, check if server immediately closes connection
      await new Promise(r => setTimeout(r, 1000));
      if (!wsNoToken.isOpen) {
        rejected = true;
      }
    } catch (err) {
      rejected = true;
      info(`Connection correctly rejected: ${err.message}`);
    } finally {
      wsNoToken.close();
    }

    if (rejected) {
      pass('Connecting to WS with missing token rejected', 'Unauthorized connection denied');
      testResults.push(true);
    } else {
      throw new Error('WS connection succeeded without token (should be rejected)');
    }
  } catch (err) {
    fail('WS missing token rejection test', err);
    testResults.push(false);
  }

  // Test 4.2: Connect to WS with invalid token
  try {
    const wsInvalidToken = new WsClient(ctx.wsUrl);
    let rejected = false;

    try {
      await wsInvalidToken.connect('invalid_jwt_token_abcdef123456');
      await new Promise(r => setTimeout(r, 1000));
      if (!wsInvalidToken.isOpen) {
        rejected = true;
      }
    } catch (err) {
      rejected = true;
      info(`Connection correctly rejected with invalid token: ${err.message}`);
    } finally {
      wsInvalidToken.close();
    }

    if (rejected) {
      pass('Connecting to WS with invalid token rejected', 'Invalid token denied access');
      testResults.push(true);
    } else {
      throw new Error('WS connection succeeded with invalid token (should be rejected)');
    }
  } catch (err) {
    fail('WS invalid token rejection test', err);
    testResults.push(false);
  }

  // Test 4.3: Duplicate friend request handling
  try {
    if (!ctx.userA || !ctx.userB) {
      throw new Error('User A or User B missing from context');
    }

    // Send friend request again from User A to User B (they are already friends or request pending)
    const target = ctx.userB.id || ctx.userB.username;
    const resDup = await api.sendFriendRequest(ctx.userA.token, target);

    if (!resDup.ok || resDup.status === 400 || resDup.status === 409) {
      pass('Duplicate friend request handling verified', `Rejected with status ${resDup.status}`);
      testResults.push(true);
    } else {
      // Some APIs return 200 with { error: "..." } or { status: "already_exists" }
      const statusMsg = resDup.data?.error || resDup.data?.message || resDup.data?.status;
      if (statusMsg) {
        pass('Duplicate friend request handling verified', `Returned message: "${statusMsg}"`);
        testResults.push(true);
      } else {
        throw new Error(`Duplicate friend request returned success status 200 without error field: ${JSON.stringify(resDup.data)}`);
      }
    }
  } catch (err) {
    fail('Duplicate friend request handling test', err);
    testResults.push(false);
  }

  return testResults.every(Boolean);
}

module.exports = runTier4;
